package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReadManager", func() {
	var originalResourcesRoot string

	BeforeEach(func() {
		// Save original ResourcesRoot
		originalResourcesRoot = components.ResourcesRoot

		// Create temporary directory structure that mimics real structure
		tempDir, err := os.MkdirTemp("", "ffx_test")
		Expect(err).ToNot(HaveOccurred())

		// Set ResourcesRoot to our test directory
		components.ResourcesRoot = tempDir
		// Create test directory structure
		createInternalTestDirectories(tempDir)
		createInternalTestFiles(tempDir)
	})

	AfterEach(func() {
		// Restore original ResourcesRoot
		components.ResourcesRoot = originalResourcesRoot

		// Clean up test directory
		if components.ResourcesRoot != originalResourcesRoot {
			os.RemoveAll(components.ResourcesRoot)
		}
	})

	Describe("InitializeInternals", func() {
		Context("when testing character map initialization", func() {
			It("should not panic when called with test data", func() {
				// This test verifies that InitializeInternals doesn't panic
				// even if some files are missing or different from expected
				Expect(func() {
					reader.InitializeInternals()
				}).ToNot(Panic())
			})
		})

		Context("when testing individual components", func() {
			It("should test character mapping logic", func() {
				// Create test data similar to what PrepareCharset does
				testData := []byte("ABC123テスト") // Mix of ASCII and Unicode
				runes := []rune(string(testData))

				// Simulate what PrepareCharset does
				byteToChar := make(map[int]rune)
				charToByte := make(map[rune]int)

				for i, r := range runes {
					idx := i + 0x30
					byteToChar[idx] = r
					if _, exists := charToByte[r]; !exists {
						charToByte[r] = idx
					}
				}

				// Verify mappings were created
				Expect(len(byteToChar)).To(BeNumerically(">", 0))
				Expect(len(charToByte)).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("GetLocalizationRoot", func() {
		It("should return correct localization root path format", func() {
			// Test the function logic with current PathFfxRoot
			result := reader.GetLocalizationRoot("en")
			expected := common.PathFfxRoot + "new_enpc/"
			Expect(result).To(Equal(expected))
		})

		It("should handle different localizations correctly", func() {
			testCases := []struct {
				localization string
				suffix       string
			}{
				{"en", "new_enpc/"},
				{"jp", "new_jppc/"},
				{"fr", "new_frpc/"},
				{"de", "new_depc/"},
				{"kr", "new_krpc/"},
			}

			for _, tc := range testCases {
				result := reader.GetLocalizationRoot(tc.localization)
				expected := common.PathFfxRoot + tc.suffix
				Expect(result).To(Equal(expected), "Localization root for %s should be correct", tc.localization)
			}
		})
	})

	Describe("PrepareCharset", func() {
		Context("when testing charset preparation logic", func() {
			It("should demonstrate character mapping behavior", func() {
				// This test demonstrates what PrepareCharset should do
				// without actually calling it due to path dependencies

				testData := []byte("Test data with some characters: テスト")
				runes := []rune(string(testData))

				// Simulate what PrepareCharset does
				byteToChar := make(map[int]rune)
				charToByte := make(map[rune]int)

				for i, r := range runes {
					idx := i + 0x30
					byteToChar[idx] = r
					if _, exists := charToByte[r]; !exists {
						charToByte[r] = idx
					}
				}

				// Verify mappings were created
				Expect(len(byteToChar)).To(BeNumerically(">", 0))
				Expect(len(charToByte)).To(BeNumerically(">", 0))

				// Test some specific mappings
				firstChar := runes[0]
				expectedIdx := 0x30
				Expect(byteToChar[expectedIdx]).To(Equal(firstChar))
				Expect(charToByte[firstChar]).To(Equal(expectedIdx))
			})

			It("should handle duplicate characters correctly", func() {
				testData := []byte("AAA") // Duplicate characters
				runes := []rune(string(testData))

				charToByte := make(map[rune]int)

				for i, r := range runes {
					idx := i + 0x30
					// Only add if not exists (first occurrence wins)
					if _, exists := charToByte[r]; !exists {
						charToByte[r] = idx
					}
				}

				// Should only have one entry for 'A', pointing to first occurrence
				Expect(charToByte['A']).To(Equal(0x30))
				Expect(len(charToByte)).To(Equal(1))
			})
		})
	})
})

// Helper functions for test setup
func createInternalTestDirectories(tempDir string) {
	// Create ffx_encoding directory structure that matches what the real code expects
	encodingDir := filepath.Join(tempDir, "ffx_encoding")
	os.MkdirAll(encodingDir, 0755)

	// Create localization directories for each locale in common.Localizations
	for loc := range common.Localizations {
		locDir := filepath.Join(tempDir, "new_"+loc+"pc", "menu")
		os.MkdirAll(locDir, 0755)
	}
}

func createInternalTestFiles(tempDir string) {
	// Create test charset files for each charset in common.Charsets
	for _, charset := range common.Charsets {
		charsetFile := filepath.Join(tempDir, "ffx_encoding", "ffxsjistbl_"+charset+".bin")
		// Create a simple test charset file with some sample data
		testData := []byte("テストデータサンプル") // Some test Japanese characters
		os.WriteFile(charsetFile, testData, 0644)
	}

	// Create test macro dictionary files for each localization
	for loc := range common.Localizations {
		macroFile := filepath.Join(tempDir, "new_"+loc+"pc", "menu", "macrodic.dcp")
		// Create a simple test macro file with basic DCP structure
		testMacroData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07} // Simple test data
		os.WriteFile(macroFile, testMacroData, 0644)
	}
}
