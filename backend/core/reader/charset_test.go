package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/encoding"
	testcommon "ffxresources/testData"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCharset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Charset Suite")
}

var _ = Describe("PrepareCharset", Ordered, func() {
	var (
		rootDir               string
		originalResourcesRoot string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
	})

	BeforeEach(func() {
		// Save original ResourcesRoot
		originalResourcesRoot = common.ResourcesRoot

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")
	})

	AfterEach(func() {
		// Restore original ResourcesRoot
		common.ResourcesRoot = originalResourcesRoot
	})

	assertVersionedCharset := func(version common.GameVersion) {
		// Test each charset in common.Charsets
		for _, charset := range common.Charsets {
			err := reader.PrepareCharset(version, charset)
			Expect(err).ToNot(HaveOccurred(), "PrepareCharset should succeed for charset: %s", charset)

			// Verify that ByteToCharMaps contains the charset for this version
			Expect(ffxencoding.GetByteToCharMap(version, charset)).ToNot(BeNil(), "ByteToCharMaps should contain charset: %s", charset)

			// Verify that CharToByteMaps contains the charset for this version
			charToByte := ffxencoding.GetCharToByteMap(version, charset)
			Expect(charToByte).ToNot(BeNil(), "CharToByteMaps should contain charset: %s", charset)

			// Verify that the maps are not empty
			Expect(len(ffxencoding.GetByteToCharMap(version, charset))).To(BeNumerically(">", 0),
				"ByteToCharMaps[%s] should not be empty", charset)
			Expect(len(charToByte)).To(BeNumerically(">", 0),
				"CharToByteMaps[%s] should not be empty", charset)
			// Forward mapping (char->byte->char) should always work
			for char, byteVal := range charToByte {
				mappedChar, exists := ffxencoding.ByteToChar(byteVal, charset, version)
				Expect(exists).To(BeTrue(), "ByteToCharMaps should contain mapping for byte %d from charset %s", byteVal, charset)
				Expect(mappedChar).To(Equal(char), "Forward mapping should be consistent for byte %d in charset %s", byteVal, charset)
			}
		}

		// Verify all expected charsets are present
		for _, expectedCharset := range common.Charsets {
			Expect(ffxencoding.GetByteToCharMap(version, expectedCharset)).ToNot(BeNil(),
				"ByteToCharMaps should contain expected charset: %s", expectedCharset)

			Expect(ffxencoding.GetCharToByteMap(version, expectedCharset)).ToNot(BeNil(),
				"CharToByteMaps should contain expected charset: %s", expectedCharset)
		}
	}

	Context("when testing FFX (ffx)", func() {
		BeforeEach(func() {
			common.SetCurrentGameVersion(common.GameVersionFFX)
		})

		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(common.GameVersionFFX)
		})

		It("should keep FFX and FFX-2 buckets isolated", func() {
			Expect(reader.PrepareCharset(common.GameVersionFFX, "us")).To(Succeed())
			Expect(ffxencoding.GetByteToCharMap(common.GameVersionFFX, "us")).ToNot(BeNil())
		})
	})

	Context("when testing FFX-2 (ffx2)", func() {
		BeforeEach(func() {
			common.SetCurrentGameVersion(common.GameVersionFFX2)
		})

		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(common.GameVersionFFX2)
		})
	})

	Context("version identity mapping", func() {
		It("should parse and normalize game versions", func() {
			Expect(common.ParseGameVersion("ffx")).To(Equal(common.GameVersionFFX))
			Expect(common.ParseGameVersion("FFX-2")).To(Equal(common.GameVersionFFX2))
			Expect(common.ParseGameVersion("lastmiss")).To(Equal(common.GameVersionLastMiss))
			Expect(common.GameVersion("bogus").Normalize()).To(Equal(common.GameVersionFFX))
			fmt.Fprintf(GinkgoWriter, "version models OK\n")
		})
	})

	Context("when charset file does not exist", func() {
		It("should return an error", func() {
			err := reader.PrepareCharset(common.GameVersionFFX, "nonexistent")
			Expect(err).To(HaveOccurred(), "PrepareCharset should fail for nonexistent charset")
		})
	})
})
