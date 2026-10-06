package services_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/lockit"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
	"ffxresources/backend/interactions"
	"ffxresources/backend/services"
	"ffxresources/testData"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metadata Service Suite")
}

var _ = Describe("MetadataService", Ordered, func() {
	var (
		rootDir         string
		gameLocation    string
		tmpRoot         string
		config          *interactions.AppConfig
		mockNotifier    *testcommon.MockNotifier
		metadataService *services.MetadataService
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetCurrentGameVersion(common.GameVersionFFX2)

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

		srcTree := filepath.Join(rootDir, "FFX-2", "binary")

		var err error
		tmpRoot, err = os.MkdirTemp("", "metadata_service")
		Expect(err).NotTo(HaveOccurred())

		gameLocation = filepath.Join(tmpRoot, "game")
		Expect(os.CopyFS(gameLocation, os.DirFS(srcTree))).To(Succeed())

		// Semeia um evento no grupo "lm" (conteúdo da Last Mission) para
		// cobrir a régua de events: ffx2 esconde "lm", lastmiss só o mostra.
		objRoot := filepath.Join(gameLocation, "ffx_ps2", "ffx2", "master", "new_uspc", "event", "obj_ps3")
		lmData, err := os.ReadFile(filepath.Join(objRoot, "hi", "hiku2800", "hiku2800.bin"))
		Expect(err).NotTo(HaveOccurred())
		lmDir := filepath.Join(objRoot, "lm", "lmhiku0000")
		Expect(os.MkdirAll(lmDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(lmDir, "lmhiku0000.bin"), lmData, 0o644)).To(Succeed())

		// Copia a árvore do FFX (jogo-pai do eternalcalm) e semeia um
		// evento no grupo "sc" (sc/scene*.bin — Eternal Calm) para cobrir
		// a régua espelhada: ffx esconde "sc", eternalcalm só o mostra.
		ffxSrc := filepath.Join(rootDir, "FFX", "binary", "ffx_ps2")
		Expect(os.CopyFS(filepath.Join(gameLocation, "ffx_ps2"), os.DirFS(ffxSrc))).To(Succeed())
		ffxObjRoot := filepath.Join(gameLocation, "ffx_ps2", "ffx", "master", "new_uspc", "event", "obj_ps3")
		sceneData, err := os.ReadFile(filepath.Join(ffxObjRoot, "az", "azit0000", "azit0000.bin"))
		Expect(err).NotTo(HaveOccurred())
		scDir := filepath.Join(ffxObjRoot, "sc", "scene1000")
		Expect(os.MkdirAll(scDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(scDir, "scene1000.bin"), sceneData, 0o644)).To(Succeed())

		config = interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(common.GameVersionFFX2)
		config.SetLocation("GameFilesLocation", gameLocation)
		config.SetLocation("ExtractLocation", filepath.Join(tmpRoot, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(tmpRoot, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(tmpRoot, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
	})

	AfterAll(func() {
		if tmpRoot != "" {
			_ = os.RemoveAll(tmpRoot)
		}
	})

	BeforeEach(func() {
		mockNotifier = testcommon.NewMockNotifier()
		Expect(mockNotifier).NotTo(BeNil())
		metadataService = services.NewMetadataService(mockNotifier)
		Expect(metadataService).NotTo(BeNil())
	})

	It("lists events without charset errors", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty(), "events should be loaded from the datastore")
	})

	It("lists objects entries", func() {
		entries, err := metadataService.ListEntries(services.KindObjects, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())
	})

	It("lists macro chunks without charset errors", func() {
		entries, err := metadataService.ListEntries(services.KindMacro, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty(), "macro chunks should be found")
	})

	It("gets an event entry with rows", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())

		entry, err := metadataService.GetEntry(services.KindEvents, entries[0].ID, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entry.Metadata.Key).NotTo(BeEmpty())
	})

	It("exports an event to json and strings", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())

		paths, err := metadataService.ExportEntry(services.KindEvents, common.GameVersionFFX2, entries[0].ID, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(HaveLen(2))
		for _, p := range paths {
			_, statErr := os.Stat(p)
			Expect(statErr).NotTo(HaveOccurred(), "exported artifact should exist: %s", p)
		}
	})

	It("hides the lm group in ffx2 events (it belongs to lastmiss)", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())
		for _, e := range entries {
			Expect(strings.HasPrefix(e.ID, "lm")).To(BeFalse(),
				"lm is Last Mission content, must not be listed for ffx2: %s", e.ID)
		}
	})

	It("lists only lm events for lastmiss", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionLastMiss)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty(), "seeded lm event should be visible in lastmiss")
		for _, e := range entries {
			Expect(strings.HasPrefix(e.ID, "lm")).To(BeTrue(),
				"lastmiss must list only the lm group: %s", e.ID)
		}
	})

	It("has no dictionary (macro) for lastmiss", func() {
		entries, err := metadataService.ListEntries(services.KindMacro, common.GameVersionLastMiss)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty(), "lastmiss has no macro dictionary of its own")
	})

	It("hides the sc group in ffx events (it belongs to eternalcalm)", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())
		for _, e := range entries {
			Expect(strings.HasPrefix(e.ID, "sc")).To(BeFalse(),
				"sc is Eternal Calm content, must not be listed for ffx: %s", e.ID)
		}
	})

	It("lists only sc events for eternalcalm", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionEternalCalm)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty(), "seeded sc event should be visible in eternalcalm")
		for _, e := range entries {
			Expect(strings.HasPrefix(e.ID, "sc")).To(BeTrue(),
				"eternalcalm must list only the sc group: %s", e.ID)
		}

		// A entrada carrega rows do binário semeado (mesma árvore do ffx).
		entry, err := metadataService.GetEntry(services.KindEvents, entries[0].ID, common.GameVersionEternalCalm)
		Expect(err).NotTo(HaveOccurred())
		Expect(entry.Rows).NotTo(BeEmpty(), "eternalcalm event must carry text rows")
		Expect(entry.Metadata.Key).To(HavePrefix("ffx/event/obj_ps3/sc/"),
			"eternalcalm key resolves under the ffx tree: %s", entry.Metadata.Key)
	})

	It("serves no own tree kinds for eternalcalm (macro, lockit, help, images)", func() {
		for _, kind := range []string{services.KindMacro, services.KindLockit, services.KindHelp} {
			entries, err := metadataService.ListEntries(kind, common.GameVersionEternalCalm)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(BeEmpty(), "eternalcalm has no %s tree of its own", kind)

			c, err := metadataService.GetCollection(kind, common.GameVersionEternalCalm, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c).To(BeEmpty(), "eternalcalm must not serve the ffx %s", kind)
		}

		// Imagens também são do jogo-pai (images não tem Collection —
		// só o índice da árvore).
		imgs, err := metadataService.ListEntries(services.KindImages, common.GameVersionEternalCalm)
		Expect(err).NotTo(HaveOccurred())
		Expect(imgs).To(BeEmpty(), "eternalcalm has no textures of its own")
	})

	It("imports strings, warns about binary save and persists it in mods", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())

		// Events crcr* são um caso especial conhecido, ainda sem correção no
		// formato (o rebuild da string table estoura o limite uint16 dos
		// ponteiros de texto; objectfile não sofre disso porque separa bloco
		// de texto de bloco de dados). Qualquer erro de crcr é ignorado: o
		// spec usa o primeiro evento comum.
		id := ""
		for _, e := range entries {
			if !strings.HasPrefix(e.ID, "crcr") {
				id = e.ID
				break
			}
		}
		Expect(id).NotTo(BeEmpty(), "deve existir um evento não-crcr para o import")
		entry, err := metadataService.GetEntry(services.KindEvents, id, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entry.Metadata.Key).NotTo(BeEmpty())

		// Encurta um texto 'us' (capacity-safe) para gerar uma alteração.
		// O Hash NÃO é re-derivado: o artefato carrega o hash do texto no
		// momento do export, e é ele que casa com a store. Um tradutor mexe
		// só no valor — re-derivar o hash faria a row casar com nada.
		changed := false
		for i := range entry.Rows {
			if t := entry.Rows[i].Text[common.DefaultLocalization]; len(t) > 1 {
				newText := t[:len(t)-1]
				entry.Rows[i].Text[common.DefaultLocalization] = newText
				changed = true
				break
			}
		}
		Expect(changed).To(BeTrue(), "a entrada deve ter um texto 'us' encurtável")

		// .strings da entrada em arquivo de importação.
		c := dto.Collection{id: entry}
		raw, err := strfmt.Marshal(c)
		Expect(err).NotTo(HaveOccurred())
		importPath := filepath.Join(tmpRoot, "reimported", "test_import.strings")
		Expect(os.MkdirAll(filepath.Dir(importPath), 0o755)).To(Succeed())
		Expect(os.WriteFile(importPath, raw, 0o644)).To(Succeed())

		// O modal deve avisar que a confirmação salva em binário.
		summary, err := metadataService.PreviewImport(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.SavesBinary).To(BeTrue(),
			"o resumo deve avisar que o import será salvo em binário")

		changedCount, err := metadataService.ImportFile(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(changedCount).To(Equal(1))

		// O binário reconstruído deve existir em <game>/mods/... — de onde
		// a carga mods-first o lê na sessão seguinte.
		p, ok := dto.ParseKey(entry.Metadata.Key)
		Expect(ok).To(BeTrue())
		Expect(p.LocalizationPattern).NotTo(BeEmpty())
		modBinary := filepath.Join(gameLocation, "mods", "ffx_ps2", "ffx2", "master",
			"new_uspc", filepath.FromSlash(p.LocalizationPattern))
		_, statErr := os.Stat(modBinary)
		Expect(statErr).NotTo(HaveOccurred(),
			"binário importado deve existir em mods: %s", modBinary)
	})

	It("lists lockit entries for ffx2 and none for lastmiss", func() {
		entries, err := metadataService.ListEntries(services.KindLockit, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1), "ffx2 tem um único kit de localização")
		Expect(entries[0].ID).To(Equal("ffx2_loc_kit_ps3"))
		Expect(entries[0].Key).To(Equal("ffx2/gamedata/ps3data/lockit/ffx2_loc_kit_ps3.bin"))

		// Last Mission é expansão do ffx2 e não tem lockit próprio.
		lm, err := metadataService.ListEntries(services.KindLockit, common.GameVersionLastMiss)
		Expect(err).NotTo(HaveOccurred())
		Expect(lm).To(BeEmpty(), "lastmiss não tem lockit próprio")

		c, err := metadataService.GetCollection(services.KindLockit, common.GameVersionLastMiss, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(c).To(BeEmpty(), "lastmiss não serve o lockit do ffx2")
	})

	It("gets a lockit entry grouped in game and utf8 rows", func() {
		entry, err := metadataService.GetEntry(services.KindLockit, "ffx2_loc_kit_ps3", common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entry.Metadata.Key).To(Equal("ffx2/gamedata/ps3data/lockit/ffx2_loc_kit_ps3.bin"))
		Expect(entry.Metadata.RowCount).To(Equal(len(entry.Rows)))
		Expect(len(entry.Rows)).To(BeNumerically(">", 1000))

		// game ocupa 0..G-1 e utf8 G..G+U-1, contíguos e nessa ordem.
		game, utf8 := 0, 0
		for _, row := range entry.Rows {
			Expect(row.Name).To(BeElementOf("game", "utf8"), "linha com name inesperado")
			if row.Name == "game" {
				Expect(row.Index).To(Equal(game), "índices game contíguos a partir de 0")
				game++
				continue
			}
			Expect(row.Index).To(Equal(game+utf8), "índices utf8 contíguos após o grupo game")
			utf8++
		}
		Expect(game).To(BeNumerically(">", 1000))
		Expect(utf8).To(BeNumerically(">", 400))
		Expect(game + utf8).To(Equal(len(entry.Rows)))
	})

	It("exports lockit to json and strings", func() {
		paths, err := metadataService.ExportEntry(services.KindLockit, common.GameVersionFFX2, "ffx2_loc_kit_ps3", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(HaveLen(2))
		for _, p := range paths {
			_, statErr := os.Stat(p)
			Expect(statErr).NotTo(HaveOccurred(), "exported artifact should exist: %s", p)
		}
	})

	It("imports edited lockit json, touching only the edited line", func() {
		const id = "ffx2_loc_kit_ps3"
		entry, err := metadataService.GetEntry(services.KindLockit, id, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())

		// Edita uma row do grupo game com texto 'us' ÚNICO no arquivo.
		// Row 2 deste arquivo compartilha o texto com outra linha — dedup
		// colapsa as duas em $hash no artefato e editar o dono da ref
		// propagaria a edição para a gêmea (2 changes). Para validar a
		// intenção "mexer só na linha editada", o alvo não pode ter gêmeo.
		ti := -1
		for i, r := range entry.Rows {
			if r.Name != "game" || len(r.Text[common.DefaultLocalization]) <= 1 {
				continue
			}
			orig := r.Text[common.DefaultLocalization]
			n := 0
			for _, o := range entry.Rows {
				if o.Text[common.DefaultLocalization] == orig {
					n++
				}
			}
			if n == 1 {
				ti = i
				break
			}
		}
		Expect(ti).To(BeNumerically(">=", 0), "deve haver row game com texto 'us' único")
		target := entry.Rows[ti]
		const newText = "LOCKIT EDITADO"
		Expect(target.Text[common.DefaultLocalization]).NotTo(Equal(newText))
		// Hash fica como no export — é a chave de casamento e o sinal de
		// staleness, não um derivado do texto editado.
		target.Text[common.DefaultLocalization] = newText

		raw, err := jsonfmt.NewJSONObjectFormatter().Marshal(dto.Collection{id: entry})
		Expect(err).NotTo(HaveOccurred())
		importPath := filepath.Join(tmpRoot, "reimported", "lockit_import.json")
		Expect(os.MkdirAll(filepath.Dir(importPath), 0o755)).To(Succeed())
		Expect(os.WriteFile(importPath, raw, 0o644)).To(Succeed())

		summary, err := metadataService.PreviewImport(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.Kind).To(Equal(services.KindLockit))
		Expect(summary.SavesBinary).To(BeTrue())
		Expect(summary.Errors).To(BeEmpty())
		Expect(summary.ChangedTexts).To(Equal(1))

		changed, err := metadataService.ImportFile(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(Equal(1))

		// Só a linha editada difere do original: rows intocadas devolvem
		// byte-a-byte a origem (re-encode apenas do que foi editado).
		rel := filepath.FromSlash("ffx-2_data/gamedata/ps3data/lockit/ffx2_loc_kit_ps3_us.bin")
		orig, err := os.ReadFile(filepath.Join(gameLocation, rel))
		Expect(err).NotTo(HaveOccurred())
		mod, err := os.ReadFile(filepath.Join(gameLocation, "mods", rel))
		Expect(err).NotTo(HaveOccurred())
		origLines := bytes.Split(orig, []byte("\n"))
		modLines := bytes.Split(mod, []byte("\n"))
		Expect(modLines).To(HaveLen(len(origLines)), "o nº de linhas não pode mudar")
		diff := 0
		for i := range origLines {
			if !bytes.Equal(origLines[i], modLines[i]) {
				diff++
			}
		}
		Expect(diff).To(Equal(1), "apenas a linha editada pode diferir da origem")

		// A edição volta na leitura mods-first, no registro certo do grupo.
		lockit.DataStore.Clear()
		l, ok := lockit.LayoutForID(common.GameVersionFFX2, id)
		Expect(ok).To(BeTrue())
		f, err := lockit.Load(l)
		Expect(err).NotTo(HaveOccurred())
		game, _ := f.IndexesByKind()
		Expect(f.Records()[game[target.Index]].Text(common.DefaultLocalization)).To(Equal(newText))
	})

	It("imports edited lockit strings back to the binary", func() {
		const id = "ffx2_loc_kit_ps3"
		entry, err := metadataService.GetEntry(services.KindLockit, id, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())

		// Edita a primeira row do grupo utf8 com texto 'us' ÚNICO no
		// arquivo — row com texto compartilhado vira ref no dedup do
		// artefato e a edição resolve para o texto do dono anterior.
		ti := -1
		for i := range entry.Rows {
			if entry.Rows[i].Name != "utf8" || entry.Rows[i].Text[common.DefaultLocalization] == "" {
				continue
			}
			orig := entry.Rows[i].Text[common.DefaultLocalization]
			n := 0
			for _, o := range entry.Rows {
				if o.Text[common.DefaultLocalization] == orig {
					n++
				}
			}
			if n == 1 {
				ti = i
				break
			}
		}
		Expect(ti).To(BeNumerically(">=", 0), "deve haver row utf8 com texto 'us' único")
		target := ti
		const newText = "LOCKIT UTF8 EDITADO"
		entry.Rows[target].Text[common.DefaultLocalization] = newText

		raw, err := strfmt.Marshal(dto.Collection{id: entry})
		Expect(err).NotTo(HaveOccurred())
		importPath := filepath.Join(tmpRoot, "reimported", "lockit_import.strings")
		Expect(os.WriteFile(importPath, raw, 0o644)).To(Succeed())

		summary, err := metadataService.PreviewImport(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.Kind).To(Equal(services.KindLockit))
		Expect(summary.Errors).To(BeEmpty(), "o .strings deve reimportar sem erros de validação")
		Expect(summary.ChangedTexts).To(Equal(1))

		changed, err := metadataService.ImportFile(importPath, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(Equal(1))

		// Conferência da posição física do grupo utf8.
		lockit.DataStore.Clear()
		l, ok := lockit.LayoutForID(common.GameVersionFFX2, id)
		Expect(ok).To(BeTrue())
		f, err := lockit.Load(l)
		Expect(err).NotTo(HaveOccurred())
		game, utf8 := f.IndexesByKind()
		phys := utf8[entry.Rows[target].Index-len(game)]
		Expect(f.Records()[phys].Text(common.DefaultLocalization)).To(Equal(newText))
	})
})
