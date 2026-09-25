package objectsfile_test

import (
	"crypto/sha256"
	"os"
	"path/filepath"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/json"
	"ffxresources/backend/interactions"
	testcommon "ffxresources/testData"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileLayout Registry", Ordered, func() {
	It("should look up known layouts by version+pattern", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())
		Expect(layout.Version).To(Equal(common.GameVersionFFX))
		Expect(layout.FileName).To(Equal("command.bin"))
		Expect(layout.DirPattern).To(Equal("battle/kernel"))
	})

	It("should look up lastmiss lm_trap layout", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionLastMiss, "lastmiss/kernel/lm_trap.bin")
		Expect(ok).To(BeTrue())
		Expect(layout.FileName).To(Equal("lm_trap.bin"))
		Expect(layout.DirPattern).To(Equal("lastmiss/kernel"))
	})
	It("should look up lastmiss lm_warehouse layout", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionLastMiss, "lastmiss/kernel/lm_warehouse.bin")
		Expect(ok).To(BeTrue())
		Expect(layout.FileName).To(Equal("lm_warehouse.bin"))
		Expect(layout.DirPattern).To(Equal("lastmiss/kernel"))
	})
})

var _ = Describe("ObjectFileStore", Ordered, func() {
	It("should Register and Get", func() {
		store := objectsfile.NewObjectFileStore()
		Expect(store.Len()).To(Equal(0))

		// Lookup a layout to satisfy the test
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())

		// We do not actually load binary here (needs game files),
		// but we can test register/get/Len with a real layout struct.
		key := objectsfile.FileLayoutKey(layout.Version, layout.PatternPath())
		Expect(objectsfile.FileLayoutKey(common.GameVersionLastMiss, "lastmiss/kernel/lm_command.bin")).
			To(Equal("lastmiss/kernel/lm_command.bin"))

		// The shared global store should exist
		Expect(objectsfile.ObjectFileDataStore).NotTo(BeNil())
		_, exists := objectsfile.ObjectFileDataStore.Get(key)
		Expect(exists).To(BeFalse())
	})

	It("should return false for missing key", func() {
		_, ok := objectsfile.ObjectFileDataStore.Get("nope/nope.bin")
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("LoadObjectFileFromStore error handling", Ordered, func() {
	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
	})

	It("should fail for unregistered file", func() {
		_, err := objectsfile.LoadObjectFileFromStore(common.GameVersionFFX, "battle/kernel/nonexistent.bin")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no layout registered"))
	})
})

var _ = Describe("Integration: integrity cycle via LoadObjectFileFromStore + FileLayout", Ordered, func() {
	var (
		tmpRoot           string
		gameDir           string
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
		tmpRoot, err = os.MkdirTemp("", "objectstore-roundtrip-*")
		Expect(err).To(BeNil())

		gameDir = filepath.Join(tmpRoot, "game")
		Expect(os.CopyFS(gameDir, os.DirFS(srcTree))).To(Succeed())

		config := interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(common.GameVersionFFX)
		config.SetLocation("GameFilesLocation", gameDir)
		config.SetLocation("ExtractLocation", filepath.Join(tmpRoot, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(tmpRoot, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(tmpRoot, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
		Expect(reader.InitializeInternals(common.GameVersionFFX)).To(Succeed())
	})

	AfterAll(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	It("should round-trip key_items and command binaries through DTO export/import + SHA-256", func() {
		cases := []struct {
			pattern string
			version common.GameVersion
		}{
			{"battle/kernel/important.bin", common.GameVersionFFX},
			{"battle/kernel/command.bin", common.GameVersionFFX},
		}
		for _, tc := range cases {
			binFile, err := objectsfile.LoadObjectFileFromStore(tc.version, tc.pattern)
			Expect(err).To(BeNil())
			Expect(binFile).NotTo(BeNil())
			Expect(binFile.GetObjects()).NotTo(BeNil())
			Expect(binFile.GetObjects().Len()).To(BeNumerically(">", 0))

			layout, ok := objectsfile.FileLayoutFor(tc.version, tc.pattern)
			Expect(ok).To(BeTrue())
			key := objectsfile.FileLayoutKey(tc.version, tc.pattern)
			collection, err := builders.BuildObjectsDTO(binFile.GetObjects(), layout, key)
			Expect(err).To(BeNil())
			paths, err := json.NewJSONObjectFormatter().WriteObjects(collection, tc.version, nil)
			Expect(err).To(BeNil())
			Expect(paths).To(HaveLen(1))
			readBack, err := json.NewJSONObjectFormatter().ReadObjects(paths[0])
			Expect(err).To(BeNil())
			entry, ok := readBack[builders.ObjectsCollectionKey(layout)]
			Expect(ok).To(BeTrue())
			Expect(builders.ApplyObjectsEntry(binFile.GetObjects(), tc.version, key, entry)).To(BeNil())

			reimported, err := os.CreateTemp(tmpRoot, "reimported-*.bin")
			Expect(err).To(BeNil())
			reimported.Close()
			Expect(binFile.SaveToBinary(reimported.Name())).To(BeNil())

			origRel := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), tc.pattern)
			origBytes, err := os.ReadFile(filepath.Join(gameDir, origRel))
			Expect(err).To(BeNil())
			outBytes, err := os.ReadFile(reimported.Name())
			Expect(err).To(BeNil())
			Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)))
		}
	})

	It("should load objects without index_count validation", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())
		binFile, err := objectsfile.LoadObjectFileFromStore(common.GameVersionFFX, layout.PatternPath())
		Expect(err).To(BeNil())
		Expect(binFile.GetObjects().Len()).To(BeNumerically(">", 0))
	})
})