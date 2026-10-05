package objectsfile_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"ffxresources/backend/builders"
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
Deslocamento global ao crescer um texto.

Regra que o tradutor depende: se o PRIMEIRO texto ganha 1 caractere, o
ponteiro do SEGUNDO tem que andar 1 byte — e todos os demais junto. Não
vale a solução "joga o texto editado pro fim do arquivo e atualiza só o
ponteiro dele": lá o segundo ponteiro fica PARADO em 0x06 quando deveria
ir para 0x07.

Este teste mede o deslocamento real do rebuild passando pelo caminho de
produção completo: BuildObjectsDTO -> ApplyObjectsEntry -> SaveToBinary
-> releitura.
*/

// shiftRef é um ref lido, identificado por (índice do objeto, campo).
type shiftRef struct {
	index  int
	name   string
	status objectsfile.PointerStatus
	offset models.Offset
	text   string
}

// shiftSnapshot varre todos os campos de todos os objetos e devolve os refs
// com texto, ordenados por offset (é a ordem em que a string table os
// organiza no arquivo).
func shiftSnapshot(objects []datastore.IGlobalLocalizedTextObject, lang string) []shiftRef {
	out := []shiftRef{}
	for i, obj := range objects {
		if obj == nil {
			continue
		}
		for _, key := range objectsfile.FieldKeys(obj) {
			seg := obj.GetKeyedString(key)
			if seg == nil {
				continue
			}
			p, ok := seg.GetLocalizedContent(lang).(legacyPointer)
			if !ok {
				continue
			}
			out = append(out, shiftRef{
				index:  i,
				name:   key,
				status: p.PointerStatus(),
				offset: p.GetOffset(),
				text:   p.GetString(),
			})
		}
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].offset != out[b].offset {
			return out[a].offset < out[b].offset
		}
		if out[a].index != out[b].index {
			return out[a].index < out[b].index
		}
		return out[a].name < out[b].name
	})
	return out
}

// findShiftRef localiza o mesmo (índice, campo) após a releitura.
func findShiftRef(refs []shiftRef, index int, name string) (shiftRef, bool) {
	for _, r := range refs {
		if r.index == index && r.name == name {
			return r, true
		}
	}
	return shiftRef{}, false
}

// Sem Ordered: a falha de um caso não pode abortar o outro — a comparação
// entre "arquivo com ref quebrado" e "arquivo limpo" É o resultado.
var _ = Describe("Deslocamento de offsets ao crescer um texto", func() {
	var (
		tmpRoot           string
		originalGameFiles string
		originalResources string
	)

	setup := func(srcTree string, version common.GameVersion) {
		common.SetCurrentGameVersion(version)

		root, err := os.MkdirTemp("", "pointer_shift")
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

	BeforeEach(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
		originalGameFiles = common.GameFilesRoot
		originalResources = common.ResourcesRoot
		setup(filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary"), common.GameVersionFFX)
	})

	AfterEach(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	// name_txt.bin tem refs MID/OOB (base aceita -> base+append);
	// command.bin é limpo (base rejeitada -> rebuild do zero).
	for _, pattern := range []string{
		"battle/kernel/name_txt.bin",
		"battle/kernel/command.bin",
	} {
		pattern := pattern
		lang := common.DefaultLocalization

		It("desloca o ponteiro seguinte em +1 em "+pattern, func() {
			version := common.GameVersionFFX
			layout, ok := objectsfile.FileLayoutFor(version, pattern)
			Expect(ok).To(BeTrue(), "sem layout para %s", pattern)

			binFile, err := objectsfile.LoadObjectFileFrom(layout, common.SourcePreferred)
			Expect(err).ToNot(HaveOccurred())
			Expect(binFile).NotTo(BeNil())

			objects := binFile.GetObjects().Items()
			Expect(objects).NotTo(BeEmpty())

			before := shiftSnapshot(objects, lang)

			// Escolhe o PRIMEIRO ref OK com texto e o ref do offset
			// imediatamente seguinte, desde que o texto seja DIFERENTE
			// (senão o dedup do propagate movia os dois juntos e a
			// medição ficaria ambígua).
			var owner, neighbour shiftRef
			found := false
			for i, r := range before {
				if r.status != objectsfile.PointerOK || r.text == "" {
					continue
				}
				for _, n := range before[i+1:] {
					if n.status == objectsfile.PointerOK && n.text != "" && n.text != r.text && n.offset > r.offset {
						owner, neighbour, found = r, n, true
						break
					}
				}
				if found {
					break
				}
			}
			Expect(found).To(BeTrue(), "não achei par de refs OK com textos distintos em %s", pattern)

			// Caminho de produção: DTO -> merge -> save.
			key := objectsfile.FileLayoutKey(version, pattern)
			collection, err := builders.BuildObjectsDTO(binFile.GetObjects(), layout, key)
			Expect(err).ToNot(HaveOccurred())
			entry, okEntry := collection[builders.ObjectsCollectionKey(layout)]
			Expect(okEntry).To(BeTrue())

			edited := false
			for i := range entry.Rows {
				if entry.Rows[i].Index == owner.index && entry.Rows[i].Name == owner.name {
					entry.Rows[i].Text[lang] = owner.text + "X" // +1 caractere
					edited = true
					break
				}
			}
			Expect(edited).To(BeTrue(), "row do ref alvo não está no DTO (%d/%s)", owner.index, owner.name)

			Expect(builders.ApplyObjectsEntry(binFile.GetObjects(), version, key, entry)).To(Succeed())

			gameDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			rel := filepath.Join(common.GetLocalizationRoot(lang), pattern)
			origBytes, err := os.ReadFile(filepath.Join(gameDir, rel))
			Expect(err).ToNot(HaveOccurred())

			outPath := filepath.Join(tmpRoot, "shifted_"+filepath.Base(pattern))
			Expect(binFile.SaveToBinary(outPath)).To(Succeed())
			outBytes, err := os.ReadFile(outPath)
			Expect(err).ToNot(HaveOccurred())

			// Releitura do binário gerado: sobrescreve a cópia do fixture
			// e carrega de novo pelo caminho de produção.
			Expect(os.WriteFile(filepath.Join(gameDir, rel), outBytes, 0o644)).To(Succeed())

			reloaded, err := objectsfile.LoadObjectFileFrom(layout, common.SourcePreferred)
			Expect(err).ToNot(HaveOccurred())
			Expect(reloaded).NotTo(BeNil())
			after := shiftSnapshot(reloaded.GetObjects().Items(), lang)

			newOwner, okOwner := findShiftRef(after, owner.index, owner.name)
			Expect(okOwner).To(BeTrue(), "ref editado sumiu da releitura")
			newNeighbour, okNeighbour := findShiftRef(after, neighbour.index, neighbour.name)
			Expect(okNeighbour).To(BeTrue(), "vizinho sumiu da releitura")

			ownerDelta := int(newOwner.offset) - int(owner.offset)
			neighbourDelta := int(newNeighbour.offset) - int(neighbour.offset)

			// Conta quantos refs mudaram de offset no total.
			moved := 0
			for _, b := range before {
				if a, ok := findShiftRef(after, b.index, b.name); ok && a.offset != b.offset {
					moved++
				}
			}

			fmt.Printf("SHIFT  %-34s dono[%d/%s] off %d->%d (delta %+d) | "+
				"vizinho[%d/%s] off %d->%d (delta %+d) | refs movidos %d/%d | bytes %d->%d\n",
				pattern,
				owner.index, owner.name, int(owner.offset), int(newOwner.offset), ownerDelta,
				neighbour.index, neighbour.name, int(neighbour.offset), int(newNeighbour.offset), neighbourDelta,
				moved, len(before),
				len(origBytes), len(outBytes))

			// A REGRA: cresceu 1 no primeiro => o seguinte anda exatamente 1.
			Expect(neighbourDelta).To(Equal(1),
				"vizinho não deslocou: %s[%d] %d -> %d (delta %+d). "+
					"Deslocamento global quebrado — o texto editado foi só jogado pro fim?",
				neighbour.name, neighbour.index,
				int(neighbour.offset), int(newNeighbour.offset), neighbourDelta)
		})
	}
})
