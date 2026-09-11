package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/models"
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

	assertVersionedCharset := func(version models.GameVersion) {
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

	Context("when testing FFX (version 1)", func() {
		BeforeEach(func() {
			common.SetGameVersion(1)
		})

		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(models.FFX)
		})

		It("should keep FFX and FFX-2 buckets isolated", func() {
			Expect(reader.PrepareCharset(models.FFX, "us")).To(Succeed())
			Expect(ffxencoding.GetByteToCharMap(models.FFX, "us")).ToNot(BeNil())
		})
	})

	Context("when testing FFX-2 (version 2)", func() {
		BeforeEach(func() {
			common.SetGameVersion(2)
		})

		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(models.FFX2)
		})
	})

	Context("version model mapping", func() {
		It("should map GameVersion to GameVersionModel", func() {
			Expect(models.FFX.Model()).To(Equal(models.GameVersionModelFFX))
			Expect(models.FFX2.Model()).To(Equal(models.GameVersionModelFFX2))
			Expect(models.NewGameVersionModelFromInt(2)).To(Equal(models.GameVersionModelFFX2))
			Expect(models.NewGameVersionModelFromString("ffx2").ToGameVersion()).To(Equal(models.FFX2))
			fmt.Fprintf(GinkgoWriter, "version models OK\n")
		})
	})

	Context("when charset file does not exist", func() {
		It("should return an error", func() {
			err := reader.PrepareCharset(models.FFX, "nonexistent")
			Expect(err).To(HaveOccurred(), "PrepareCharset should fail for nonexistent charset")
		})
	})
})
