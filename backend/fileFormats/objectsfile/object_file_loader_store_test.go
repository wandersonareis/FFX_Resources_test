package objectsfile_test

import (
	"crypto/sha256"
	"os"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
	testcommon "ffxresources/testData"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// storeVsReadCase pareia um arquivo com seu oráculo Read* (caminho congelado).
// O layout do FileLayout deve ser idêntico ao usado pelo Read correspondente.
type storeVsReadCase struct {
	pattern string
	version common.GameVersion
	read    func(string, common.GameVersion) datastore.IBinaryFile
}

var storeVsReadCases = []storeVsReadCase{
	{pattern: "battle/kernel/command.bin", version: common.GameVersionFFX, read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/important.bin", version: common.GameVersionFFX, read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/item.bin", version: common.GameVersionFFX, read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/monster1.bin", version: common.GameVersionFFX, read: objectsfile.ReadMonsterLocalizations},
	{pattern: "battle/kernel/btl_txt.bin", version: common.GameVersionFFX, read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/name_txt.bin", version: common.GameVersionFFX, read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/ply_save.bin", version: common.GameVersionFFX, read: objectsfile.ReadNameOnlyV2Localizations},
}

var _ = Describe("LoadObjectFile differential vs Read oracle", Ordered, func() {
	var (
		tmpRoot           string
		originalGameFiles string
		originalResources string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)

		originalGameFiles = common.GameFilesRoot
		originalResources = common.ResourcesRoot

		srcTree := filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary")
		common.SetCurrentGameVersion(common.GameVersionFFX)

		var err error
		tmpRoot, err = os.MkdirTemp("", "objectstore_differential")
		Expect(err).ToNot(HaveOccurred())

		gameDir := filepath.Join(tmpRoot, "game")
		Expect(os.CopyFS(gameDir, os.DirFS(srcTree))).To(Succeed())

		config := interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(common.GameVersionFFX)
		config.SetLocation("GameFilesLocation", gameDir)
		config.SetLocation("ExtractLocation", filepath.Join(tmpRoot, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(tmpRoot, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(tmpRoot, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	AfterAll(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	for _, tc := range storeVsReadCases {
		tc := tc
		It("should map identical texts to Read for "+tc.pattern, func() {
			oracle := tc.read(tc.pattern, tc.version)
			Expect(oracle).NotTo(BeNil())
			Expect(oracle.GetObjects()).NotTo(BeNil())

			objectsfile.ResetObjectFileStoreForTest()
			stored, err := objectsfile.LoadObjectFileFromStore(tc.version, tc.pattern)
			Expect(err).To(BeNil())
			Expect(stored).NotTo(BeNil())
			Expect(stored.GetObjects()).NotTo(BeNil())

			Expect(stored.GetObjects().Len()).To(Equal(oracle.GetObjects().Len()))

			oracleItems := oracle.GetObjects().Items()
			storedItems := stored.GetObjects().Items()
			for i := range oracleItems {
				Expect(storedItems[i].ToString(common.DefaultLocalization)).
					To(Equal(oracleItems[i].ToString(common.DefaultLocalization)),
						"text mismatch at index %d", i)
			}
		})

		It("should round-trip "+tc.pattern+" with matching hash", func() {
			objectsfile.ResetObjectFileStoreForTest()
			stored, err := objectsfile.LoadObjectFileFromStore(tc.version, tc.pattern)
			Expect(err).To(BeNil())

			outPath := filepath.Join(tmpRoot, "reimported", tc.pattern)
			Expect(os.MkdirAll(filepath.Dir(outPath), 0755)).To(Succeed())
			Expect(stored.SaveToBinary(outPath)).To(Succeed())

			gameDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			rel := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), tc.pattern)
			origBytes, err := os.ReadFile(filepath.Join(gameDir, rel))
			Expect(err).ToNot(HaveOccurred())
			outBytes, err := os.ReadFile(outPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)))
		})
	}
})

var _ = Describe("Layout version in file paths", Ordered, func() {
	It("should map versions to ffx/ffx2 path names", func() {
		Expect(common.VersionPathName(common.GameVersionFFX)).To(Equal("ffx"))
		Expect(common.VersionPathName(common.GameVersionFFX2)).To(Equal("ffx2"))
		Expect(common.VersionPathName(common.GameVersionLastMiss)).To(Equal("ffx2"))
	})

	It("should build FilePath with the layout version root", func() {
		ffx, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())
		Expect(ffx.FilePath()).To(Equal("ffx/battle/kernel/command.bin"))

		lm, ok := objectsfile.FileLayoutFor(common.GameVersionLastMiss, "lastmiss/kernel/lm_command.bin")
		Expect(ok).To(BeTrue())
		Expect(lm.FilePath()).To(Equal("ffx2/lastmiss/kernel/lm_command.bin"))
	})
})

var _ = Describe("LastMiss load without global version", Ordered, func() {
	var (
		tmpRoot           string
		originalGameFiles string
		originalResources string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)

		originalGameFiles = common.GameFilesRoot
		originalResources = common.ResourcesRoot

		srcTree := filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX-2", "binary")

		var err error
		tmpRoot, err = os.MkdirTemp("", "objectstore_lastmiss")
		Expect(err).ToNot(HaveOccurred())

		gameDir := filepath.Join(tmpRoot, "game")
		Expect(os.CopyFS(gameDir, os.DirFS(srcTree))).To(Succeed())

		config := interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(common.GameVersionLastMiss)
		config.SetLocation("GameFilesLocation", gameDir)
		config.SetLocation("ExtractLocation", filepath.Join(tmpRoot, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(tmpRoot, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(tmpRoot, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	AfterAll(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	loadTexts := func() []string {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionLastMiss, "lastmiss/kernel/lm_accesary.bin")
		Expect(ok).To(BeTrue())

		objectsfile.ResetObjectFileStoreForTest()
		stored, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
		Expect(err).To(BeNil())
		Expect(stored.GetObjects()).NotTo(BeNil())
		Expect(stored.GetObjects().Len()).To(BeNumerically(">", 0))

		var texts []string
		for _, obj := range stored.GetObjects().Items() {
			texts = append(texts, obj.ToString(common.DefaultLocalization))
		}
		return texts
	}

	It("should load lastmiss layout with global forced to FFX", func() {
		common.SetCurrentGameVersion(common.GameVersionFFX)
		Expect(loadTexts()).NotTo(BeEmpty())
	})

	It("should map identical texts regardless of global version", func() {
		common.SetCurrentGameVersion(common.GameVersionFFX)
		asFFX := loadTexts()

		common.SetCurrentGameVersion(common.GameVersionLastMiss)
		asLastMiss := loadTexts()

		Expect(asFFX).To(Equal(asLastMiss))
	})
})
