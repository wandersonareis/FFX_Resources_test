package objectsfile_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
A guarda do ponteiro no fluxo legado (ObjectBinaryFile + KeyedStringFile),
que é o caminho do serviço em produção.

O ref é DIREÇÃO, não fato: um offset que não cai na fronteira de um bloco
(MID) ou que sai da string table (OOB) não tem texto — o sufixo de outro
texto jamais vira frase. Aqui se fixa o que o resto do app depende:

  1. leitura  — ref quebrado é classificado e expõe "" (nunca o sufixo do
     texto anterior);
  2. export   — o ref quebrado não contribui linha de texto no idioma
     primário (sem linha não há diálogo, sem diálogo não há edição);
  3. base — só entra quando o arquivo TEM ref quebrado. Com ref quebrado
     e sem edição o save é BYTE-IDÊNTICO ao original; sem ref quebrado a
     base é rejeitada de propósito e o rebuild dedup de sempre roda.
*/

// legacyPointer é o contrato observável de um ref lido (KeyedString).
type legacyPointer interface {
	PointerStatus() objectsfile.PointerStatus
	GetOffset() models.Offset
	GetString() string
}

// refState é o estado de um campo textual de um objeto no idioma primário.
type refState struct {
	key    string
	status objectsfile.PointerStatus
	offset models.Offset
	text   string
}

// refStates percorre TODOS os campos do objeto — não só os exportáveis —,
// que é onde a guarda precisa enxergar: o ref quebrado não está no export.
func refStates(obj datastore.IGlobalLocalizedTextObject, languageCode string) []refState {
	var out []refState
	for _, key := range objectsfile.FieldKeys(obj) {
		st := refState{key: key}
		seg := obj.GetKeyedString(key)
		if seg != nil {
			if p, ok := seg.GetLocalizedContent(languageCode).(legacyPointer); ok {
				st.status = p.PointerStatus()
				st.offset = p.GetOffset()
				st.text = p.GetString()
			}
		}
		out = append(out, st)
	}
	return out
}

// exportedPrimary devolve o texto do idioma primário exportado por campo:
// o export é o único caminho que vira linha de texto no DTO.
func exportedPrimary(obj datastore.IGlobalLocalizedTextObject, languageCode string) map[string]string {
	out := map[string]string{}
	for _, f := range objectsfile.ExportFieldTexts(obj) {
		out[f.Key] = f.Texts[languageCode]
	}
	return out
}

var _ = Describe("Pointer guard (fluxo legado)", Ordered, func() {
	var (
		tmpRoot           string
		originalGameFiles string
		originalResources string
	)

	// setup espelha o do BinaryFile Integrity: cópia de testData num dir
	// temporário, interaction service apontando pra ele e charsets prontos.
	setup := func(srcTree string, version common.GameVersion) {
		common.SetCurrentGameVersion(version)

		root, err := os.MkdirTemp("", "pointer_guard")
		Expect(err).ToNot(HaveOccurred())
		tmpRoot = root

		gameDir := filepath.Join(root, "game")
		Expect(os.CopyFS(gameDir, os.DirFS(srcTree))).To(Succeed())

		config := interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(version)
		config.SetLocation("GameFilesLocation", gameDir)
		config.SetLocation("ExtractLocation", filepath.Join(root, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(root, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(root, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
		Expect(common.GameFilesRoot).To(Equal(gameDir))
		Expect(reader.InitializeInternals(version)).To(Succeed())
	}

	originalBytes := func(pattern string) []byte {
		gameDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
		rel := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), pattern)
		data, err := os.ReadFile(filepath.Join(gameDir, rel))
		Expect(err).ToNot(HaveOccurred())
		return data
	}

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
		originalGameFiles = common.GameFilesRoot
		originalResources = common.ResourcesRoot
		setup(filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary"), common.GameVersionFFX)
	})

	AfterAll(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	// ffxIntegrityCases é a MESMA tabela de leitura por arquivo que o
	// serviço usa em produção (FileLayouts).
	for _, c := range ffxIntegrityCases {
		c := c
		lang := common.DefaultLocalization

		It("guarda os refs de "+c.pattern, func() {
			binFile := c.read(c.pattern, common.GameVersionFFX)
			Expect(binFile).NotTo(BeNil())
			objects := binFile.GetObjects().Items()
			Expect(objects).NotTo(BeEmpty())

			broken, visited := 0, 0
			hist := map[objectsfile.PointerStatus]int{}
			var kinds, keysSeen []string
			seenKind, seenKey := map[string]bool{}, map[string]bool{}
			for i, obj := range objects {
				if kind := fmt.Sprintf("%T", obj); !seenKind[kind] {
					seenKind[kind] = true
					kinds = append(kinds, kind)
				}
				for _, k := range objectsfile.FieldKeys(obj) {
					if !seenKey[k] {
						seenKey[k] = true
						keysSeen = append(keysSeen, k)
					}
				}
				exported := exportedPrimary(obj, lang)
				for _, st := range refStates(obj, lang) {
					visited++
					hist[st.status]++
					if !st.status.IsBroken() {
						continue
					}
					broken++
					// A leitura não pode devolver o sufixo de outro texto.
					Expect(st.text).To(BeEmpty(),
						"obj %d campo %q [%s @ %d] leu texto %q", i, st.key, st.status, st.offset, st.text)
					// E o export não pode virar linha de texto.
					Expect(exported[st.key]).To(BeEmpty(),
						"obj %d campo %q [%s @ %d] virou linha de texto no export", i, st.key, st.status, st.offset)
				}
			}

			fmt.Printf("POINTER %-42s refs=%-5d quebrados=%-4d tipos=%v chaves=%v hist=%v\n",
				c.pattern, visited, broken, kinds, keysSeen, hist)

			// Sem ref quebrado a string table original é REJEITADA de
			// propósito: a guarda fica desligada e o rebuild dedup de
			// sempre roda, como antes da mudança.
			if broken == 0 {
				return
			}

			// Com ref quebrado a base é aceita: sem edição não há texto
			// novo, então o arquivo salvo tem que sair byte-idêntico.
			outPath := filepath.Join(tmpRoot, "reimported", filepath.Base(c.pattern))
			Expect(binFile.SaveToBinary(outPath)).To(Succeed())
			outBytes, err := os.ReadFile(outPath)
			Expect(err).ToNot(HaveOccurred())
			origBytes := originalBytes(c.pattern)
			Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)),
				"save sem edição não reproduziu o original (%d vs %d bytes) — a base não foi aceita",
				len(outBytes), len(origBytes))
		})
	}

	// O CAMINHO DE PRODUÇÃO: o serviço lê por FileLayoutFor, e os Fields
	// desse mapa podem ser mais largos que os layouts dos leitores de cima
	// (mesmo arquivo, mais campos). A guarda tem que valer onde o serviço
	// lê — medir só pelo leitor de teste enxergaria meia tabela.
	for _, c := range ffxIntegrityCases {
		c := c
		lang := common.DefaultLocalization

		It("guarda os refs que o serviço lê de "+c.pattern, func() {
			layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, c.pattern)
			if !ok {
				fmt.Printf("PROD    %-42s sem layout em FileLayouts\n", c.pattern)
				return
			}
			binFile, err := objectsfile.LoadObjectFileFrom(layout, common.SourcePreferred)
			if err != nil || binFile == nil {
				fmt.Printf("PROD    %-42s ilegível: %v\n", c.pattern, err)
				return
			}
			objects := binFile.GetObjects().Items()

			broken, visited := 0, 0
			hist := map[objectsfile.PointerStatus]int{}
			var keysSeen []string
			seenKey := map[string]bool{}
			for i, obj := range objects {
				for _, k := range objectsfile.FieldKeys(obj) {
					if !seenKey[k] {
						seenKey[k] = true
						keysSeen = append(keysSeen, k)
					}
				}
				exported := exportedPrimary(obj, lang)
				for _, st := range refStates(obj, lang) {
					visited++
					hist[st.status]++
					if !st.status.IsBroken() {
						continue
					}
					broken++
					Expect(st.text).To(BeEmpty(),
						"obj %d campo %q [%s @ %d] leu texto %q", i, st.key, st.status, st.offset, st.text)
					Expect(exported[st.key]).To(BeEmpty(),
						"obj %d campo %q [%s @ %d] virou linha de texto no export", i, st.key, st.status, st.offset)
				}
			}

			fmt.Printf("PROD    %-42s refs=%-5d quebrados=%-4d chaves=%v hist=%v\n",
				c.pattern, visited, broken, keysSeen, hist)

			// Sem ref quebrado não há problema para ligar: a base é
			// rejeitada de propósito e o rebuild dedup de sempre roda.
			if broken == 0 {
				return
			}

			// Com ref quebrado a string table ORIGINAL entra como base —
			// sem edição não há texto novo, então o arquivo tem que sair
			// byte-idêntico. É a prova de que a base foi aceita no caminho
			// que o serviço usa.
			outPath := filepath.Join(tmpRoot, "reimported", "prod_"+filepath.Base(c.pattern))
			Expect(binFile.SaveToBinary(outPath)).To(Succeed())
			outBytes, err := os.ReadFile(outPath)
			Expect(err).ToNot(HaveOccurred())
			origBytes := originalBytes(c.pattern)
			Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)),
				"save sem edição não reproduziu o original (%d vs %d bytes) — a base não foi aceita",
				len(outBytes), len(origBytes))
		})
	}
})
