package services_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
	strfmt "ffxresources/backend/formatters/strings"
	"ffxresources/backend/interactions"
	"ffxresources/backend/services"
	"ffxresources/testData"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	It("imports strings, warns about binary save and persists it in mods", func() {
		entries, err := metadataService.ListEntries(services.KindEvents, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).NotTo(BeEmpty())

		id := entries[0].ID
		entry, err := metadataService.GetEntry(services.KindEvents, id, common.GameVersionFFX2)
		Expect(err).NotTo(HaveOccurred())
		Expect(entry.Metadata.Key).NotTo(BeEmpty())

		// Encurta um texto 'us' (capacity-safe) para gerar uma alteração.
		changed := false
		for i := range entry.Rows {
			if t := entry.Rows[i].Text[common.DefaultLocalization]; len(t) > 1 {
				newText := t[:len(t)-1]
				entry.Rows[i].Text[common.DefaultLocalization] = newText
				entry.Rows[i].Hash[common.DefaultLocalization] = hash.Sum64Hex(newText)
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
})
