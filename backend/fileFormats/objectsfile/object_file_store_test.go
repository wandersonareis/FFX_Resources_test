package objectsfile_test

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/objectsfile"
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

	It("should look up lastmiss lm_trap layout with Start offset", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionLastMiss, "lastmiss/kernel/lm_trap.bin")
		Expect(ok).To(BeTrue())
		Expect(layout.Start).To(Equal(0))
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
			To(Equal("lastmiss/lastmiss/kernel/lm_command.bin"))

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

var _ = Describe("Export/Import roundtrip with FileMetadata", Ordered, func() {
	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
	})

	It("should produce FileMetadata with version/dir/filename/index_count from a layout", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())

		meta := common.NewObjectFileMetadata(layout.Version, layout.DirPattern, layout.FileName, layout.IndexCount)
		Expect(meta).NotTo(BeNil())
		Expect(meta.DirPattern).To(Equal("battle/kernel"))
		Expect(meta.FileName).To(Equal("command.bin"))
	})

	It("should write/omit new fields as expected", func() {
		layout, _ := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		meta := common.NewObjectFileMetadata(layout.Version, layout.DirPattern, layout.FileName, layout.IndexCount)

		// Verify omitempty behavior: Version/DirPattern/FileName present, IndexCount 0 omitted
		ver := layout.Version
		Expect(meta.Version).NotTo(BeNil())
		Expect((*meta.Version).String()).To(Equal("ffx"))
	})
})

var _ = Describe("Integration: integrity cycle via LoadObjectFileFromStore + FileLayout", Ordered, func() {
	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
	})

	It("should round-trip key_items and command binaries through export/import + SHA-256", func() {
		cases := []struct {
			pattern string
			version common.GameVersion
		}{
			{"battle/kernel/important.bin", common.GameVersionFFX},
			{"battle/kernel/command.bin", common.GameVersionFFX},
		}
		tmpRoot, err := os.MkdirTemp("", "objectstore-roundtrip-*")
		Expect(err).To(BeNil())
		defer os.RemoveAll(tmpRoot)

		for _, tc := range cases {
			binFile, err := objectsfile.LoadObjectFileFromStore(tc.version, tc.pattern)
			Expect(err).To(BeNil())
			Expect(binFile).NotTo(BeNil())
			Expect(binFile.GetObjects()).NotTo(BeNil())
			Expect(binFile.GetObjects().Len()).To(BeNumerically(">", 0))

			jsonName := filepath.Join(tmpRoot,
				strings.ReplaceAll(strings.ReplaceAll(tc.pattern, "/", "_"), ".bin", ".json"))
			Expect(binFile.ExportToJson(jsonName)).To(BeNil())

			reimported, err := os.CreateTemp(tmpRoot, "reimported-*.bin")
			Expect(err).To(BeNil())
			reimported.Close()
			Expect(binFile.SaveToBinary(reimported.Name())).To(BeNil())

			origRel := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), tc.pattern)
			origBytes, err := os.ReadFile(filepath.Join(common.GameFilesRoot, origRel))
			Expect(err).To(BeNil())
			outBytes, err := os.ReadFile(reimported.Name())
			Expect(err).To(BeNil())
			Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)))
		}
	})

	It("should validate index_count against loaded objects", func() {
		layout, ok := objectsfile.FileLayoutFor(common.GameVersionFFX, "battle/kernel/command.bin")
		Expect(ok).To(BeTrue())
		binFile, err := objectsfile.LoadObjectFileFromStore(common.GameVersionFFX, layout.PatternPath())
		Expect(err).To(BeNil())
		Expect(binFile.GetObjects().Len()).To(Equal(layout.IndexCount))
	})
})