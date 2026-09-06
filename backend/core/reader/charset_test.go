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

	Context("when testing FFX (version 1)", func() {
		BeforeEach(func() {
			common.SetGameVersion(1)
		})

		It("should prepare charset maps for all expected charsets", func() {
			// Test each charset in common.Charsets
			for _, charset := range common.Charsets {
				err := reader.PrepareCharset(charset)
				Expect(err).ToNot(HaveOccurred(), "PrepareCharset should succeed for charset: %s", charset)

				// Verify that ByteToCharMaps contains the charset
				_, exists := ffxencoding.ByteToCharMaps[charset]
				Expect(exists).To(BeTrue(), "ByteToCharMaps should contain charset: %s", charset)

				// Verify that CharToByteMaps contains the charset
				_, exists = ffxencoding.CharToByteMaps[charset]
				Expect(exists).To(BeTrue(), "CharToByteMaps should contain charset: %s", charset)

				// Verify that the maps are not empty
				Expect(len(ffxencoding.ByteToCharMaps[charset])).To(BeNumerically(">", 0),
					"ByteToCharMaps[%s] should not be empty", charset)
				Expect(len(ffxencoding.CharToByteMaps[charset])).To(BeNumerically(">", 0),
					"CharToByteMaps[%s] should not be empty", charset) // Verify mapping consistency understanding duplicate characters
				// Forward mapping (char->byte->char) should always work
				for char, byteVal := range ffxencoding.CharToByteMaps[charset] {
					mappedChar, exists := ffxencoding.ByteToCharMaps[charset][byteVal]
					Expect(exists).To(BeTrue(), "ByteToCharMaps should contain mapping for byte %d from charset %s", byteVal, charset)
					Expect(mappedChar).To(Equal(char), "Forward mapping should be consistent for byte %d in charset %s", byteVal, charset)
				}

				// Note: Reverse mapping (byte->char->byte) may not work for duplicate characters
				// This is expected behavior due to the implementation in buildMappings
				// CharToByteMaps only contains the first occurrence of duplicate characters
			}

			// Verify all expected charsets are present
			for _, expectedCharset := range common.Charsets {
				_, exists := ffxencoding.ByteToCharMaps[expectedCharset]
				Expect(exists).To(BeTrue(), "ByteToCharMaps should contain expected charset: %s", expectedCharset)

				_, exists = ffxencoding.CharToByteMaps[expectedCharset]
				Expect(exists).To(BeTrue(), "CharToByteMaps should contain expected charset: %s", expectedCharset)
			}
		})
	})

	Context("when testing FFX-2 (version 2)", func() {
		BeforeEach(func() {
			common.SetGameVersion(2)
		})

		It("should prepare charset maps for all expected charsets", func() {
			// Test each charset in common.Charsets
			for _, charset := range common.Charsets {
				err := reader.PrepareCharset(charset)
				Expect(err).ToNot(HaveOccurred(), "PrepareCharset should succeed for charset: %s", charset)

				// Verify that ByteToCharMaps contains the charset
				_, exists := ffxencoding.ByteToCharMaps[charset]
				Expect(exists).To(BeTrue(), "ByteToCharMaps should contain charset: %s", charset)

				// Verify that CharToByteMaps contains the charset
				_, exists = ffxencoding.CharToByteMaps[charset]
				Expect(exists).To(BeTrue(), "CharToByteMaps should contain charset: %s", charset)

				// Verify that the maps are not empty
				Expect(len(ffxencoding.ByteToCharMaps[charset])).To(BeNumerically(">", 0),
					"ByteToCharMaps[%s] should not be empty", charset)
				Expect(len(ffxencoding.CharToByteMaps[charset])).To(BeNumerically(">", 0),
					"CharToByteMaps[%s] should not be empty", charset)

				// Verify mapping consistency understanding duplicate characters
				// Forward mapping (char->byte->char) should always work
				for char, byteVal := range ffxencoding.CharToByteMaps[charset] {
					mappedChar, exists := ffxencoding.ByteToCharMaps[charset][byteVal]
					if mappedChar != char {
						fmt.Printf("Mismatch for charset %s, char %c: byte %d maps to %c instead of %c\n",
							charset, char, byteVal, mappedChar, char)
					}
					Expect(exists).To(BeTrue(), "ByteToCharMaps should contain mapping for byte %d from charset %s", byteVal, charset)
					Expect(mappedChar).To(Equal(char), "Forward mapping should be consistent for byte %d in charset %s", byteVal, charset)
				}
				// Verify all expected charsets are present
				for _, expectedCharset := range common.Charsets {
					_, exists := ffxencoding.ByteToCharMaps[expectedCharset]
					Expect(exists).To(BeTrue(), "ByteToCharMaps should contain expected charset: %s", expectedCharset)

					_, exists = ffxencoding.CharToByteMaps[expectedCharset]
					Expect(exists).To(BeTrue(), "CharToByteMaps should contain expected charset: %s", expectedCharset)
				}
				// Note: Reverse mapping (byte->char->byte) may not work for duplicate characters
				// This is expected behavior due to the implementation in buildMappings
				// CharToByteMaps only contains the first occurrence of duplicate characters
			}

			// Verify all expected charsets are present
			for _, expectedCharset := range common.Charsets {
				_, exists := ffxencoding.ByteToCharMaps[expectedCharset]
				Expect(exists).To(BeTrue(), "ByteToCharMaps should contain expected charset: %s", expectedCharset)

				_, exists = ffxencoding.CharToByteMaps[expectedCharset]
				Expect(exists).To(BeTrue(), "CharToByteMaps should contain expected charset: %s", expectedCharset)
			}
		})
	})

	Context("when charset file does not exist", func() {
		It("should return an error", func() {
			err := reader.PrepareCharset("nonexistent")
			Expect(err).To(HaveOccurred(), "PrepareCharset should fail for nonexistent charset")
		})
	})
})
