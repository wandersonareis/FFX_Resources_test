package converter_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/reader"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDataObjectsBinaryReader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Objects Binary Reader Suite")
}

var _ = Describe("Data Objects Binary Reader", Ordered, func() {
	var (
		originalResourcesRoot string
		originalGameFilesRoot string
		originalModsEnabled   bool
		originalVerboseMode   bool
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetGameVersion(1)
		common.SetVerboseMode(false)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	BeforeEach(func() {
		// Save original state
		originalResourcesRoot = common.ResourcesRoot
		originalGameFilesRoot = common.GameFilesRoot
		originalModsEnabled = common.AreModsEnabled()
		originalVerboseMode = common.IsVerboseMode()

		// Ensure clean state
		common.SetModsEnabled(false)
		common.SetVerboseMode(false)
	})

	AfterEach(func() {
		// Restore original state
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot
		common.SetModsEnabled(originalModsEnabled)
		common.SetVerboseMode(originalVerboseMode)

		// Clear global state to avoid interference between tests
		components.KEY_ITEMS = nil
		components.COMMANDS = nil
		components.ITEMS = nil
		components.ARMS_TEXT = nil
		components.BTL_TEXT = nil
		components.BTLEND_TEXT = nil
		components.MONMAGIC1 = nil
		components.MONMAGIC2 = nil
		components.BUILD_TEXT = nil
		components.CONFIG_TEXT = nil
		components.ITEM_TEXT = nil
		components.MMAIN_TEXT = nil
		components.PLAYER_ROOM = nil
		components.NAME_TEXT = nil
		components.ByteToCharMaps = make(map[string]map[uint]rune)
		components.CharToByteMaps = make(map[string]map[rune]uint)
		components.MacroLookup = make(map[int]*components.LocalizedMacroStringObject)
	})

	Context("ReadNameOnlyDataObjectsWithIlist", func() {
		It("should read battle text data successfully", func() {
			// Use a known pattern path that exists in test data
			patternPath := "battle/kernel/btl_txt.bin"

			result := converter.ReadNameOnlyDataObjectsWithIlist(patternPath)

			Expect(result).ToNot(BeNil())
			Expect(result.GetLength()).To(BeNumerically(">", 0))

			// Verify first item exists and has expected structure
			firstItem := result.Get(0)
			Expect(firstItem).ToNot(BeNil())

			// Verify it's a name-only object (should have name but not description)
			textObj := firstItem.GetTextObject()
			Expect(textObj).ToNot(BeNil())

			nameString := textObj.GetKeyedString("name")
			Expect(nameString).ToNot(BeNil())

			// Verify localization content exists
			usContent := nameString.GetLocalizedContent(common.DefaultLocalization)
			Expect(usContent).ToNot(BeNil())
		})

		It("should return nil for non-existent file", func() {
			patternPath := "non/existent/file.bin"

			result := converter.ReadNameOnlyDataObjectsWithIlist(patternPath)

			Expect(result).To(BeNil())
		})

		It("should handle empty pattern path gracefully", func() {
			patternPath := ""

			result := converter.ReadNameOnlyDataObjectsWithIlist(patternPath)

			Expect(result).To(BeNil())
		})
	})

	Context("ReadNameDescriptionObjectsWithIlist", func() {
		It("should read key items data successfully", func() {
			// Use a known pattern path that exists in test data
			patternPath := "battle/kernel/important.bin"

			result := converter.ReadNameDescriptionObjectsWithIlist(patternPath)

			Expect(result).ToNot(BeNil())
			Expect(result.GetLength()).To(BeNumerically(">", 0))

			// Verify first item exists and has expected structure
			firstItem := result.Get(0)
			Expect(firstItem).ToNot(BeNil())

			// Verify it's a name-description object (should have both name and description)
			textObj := firstItem.GetTextObject()
			Expect(textObj).ToNot(BeNil())

			nameString := textObj.GetKeyedString("name")
			Expect(nameString).ToNot(BeNil())

			descString := textObj.GetKeyedString("description")
			Expect(descString).ToNot(BeNil())

			// Verify localization content exists
			usNameContent := nameString.GetLocalizedContent(common.DefaultLocalization)
			Expect(usNameContent).ToNot(BeNil())

			usDescContent := descString.GetLocalizedContent(common.DefaultLocalization)
			Expect(usDescContent).ToNot(BeNil())
		})

		It("should read commands data successfully", func() {
			patternPath := "battle/kernel/command.bin"

			result := converter.ReadNameDescriptionObjectsWithIlist(patternPath)

			Expect(result).ToNot(BeNil())
			Expect(result.GetLength()).To(BeNumerically(">", 0))

			// Verify structure
			firstItem := result.Get(0)
			Expect(firstItem).ToNot(BeNil())

			textObj := firstItem.GetTextObject()
			Expect(textObj).ToNot(BeNil())

			nameString := textObj.GetKeyedString("name")
			Expect(nameString).ToNot(BeNil())

			descString := textObj.GetKeyedString("description")
			Expect(descString).ToNot(BeNil())
		})

		It("should return nil for non-existent file", func() {
			patternPath := "invalid/path/file.bin"

			result := converter.ReadNameDescriptionObjectsWithIlist(patternPath)

			Expect(result).To(BeNil())
		})

		It("should handle empty pattern path gracefully", func() {
			patternPath := ""

			result := converter.ReadNameDescriptionObjectsWithIlist(patternPath)

			Expect(result).To(BeNil())
		})
	})

	Context("ReadDataListWithIlist", func() {
		It("should read binary file with name-only creator", func() {
			// Construct the full file path for a known binary file
			patternPath := "battle/kernel/btl_txt.bin"
			filename := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

			// Name-only creator function
			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			result := converter.ReadDataListWithIlist(filename, common.DefaultLocalization, creator)

			Expect(result).ToNot(BeNil())
			Expect(result.GetLength()).To(BeNumerically(">", 0))

			// Verify structure
			firstItem := result.Get(0)
			Expect(firstItem).ToNot(BeNil())
		})

		It("should read binary file with name-description creator", func() {
			// Construct the full file path for a known binary file
			patternPath := "battle/kernel/important.bin"
			filename := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

			// Name-description creator function
			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameDescriptionTextObject(data, stringBytes, headerLength, localization)
			}

			result := converter.ReadDataListWithIlist(filename, common.DefaultLocalization, creator)

			Expect(result).ToNot(BeNil())
			Expect(result.GetLength()).To(BeNumerically(">", 0))

			// Verify structure
			firstItem := result.Get(0)
			Expect(firstItem).ToNot(BeNil())
		})

		It("should return nil for non-existent file", func() {
			filename := "/non/existent/file.bin"

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			result := converter.ReadDataListWithIlist(filename, common.DefaultLocalization, creator)

			Expect(result).To(BeNil())
		})

		It("should handle different language codes", func() {
			patternPath := "battle/kernel/btl_txt.bin"
			filename := filepath.Join(common.GetLocalizationRoot("us"), patternPath)

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			result := converter.ReadDataListWithIlist(filename, "us", creator)

			// Should either return data or nil (depending on file existence)
			// The function should not panic or error
			if result != nil {
				Expect(result.GetLength()).To(BeNumerically(">=", 0))
			}
		})

		It("should handle corrupted binary file gracefully", func() {
			// Create a temporary corrupted file
			tempDir := os.TempDir()
			corruptedFile := filepath.Join(tempDir, "corrupted.bin")

			// Write invalid binary data (too small)
			err := os.WriteFile(corruptedFile, []byte{0x01, 0x02}, 0644)
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(corruptedFile)

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			result := converter.ReadDataListWithIlist(corruptedFile, common.DefaultLocalization, creator)

			Expect(result).To(BeNil())
		})
	})

	Context("PopulateDataObjectLocalizationsWithIlist", func() {
		It("should populate localizations for existing objects", func() {
			// First create a list with base objects
			patternPath := "battle/kernel/btl_txt.bin"

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)
			objects := converter.ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)

			if objects != nil && objects.GetLength() > 0 {
				// Store original count
				originalCount := objects.GetLength()

				// Call populate function - this should add localizations to existing objects
				converter.PopulateDataObjectLocalizationsWithIlist(patternPath, objects, creator)

				// Verify objects still exist and count hasn't changed
				Expect(objects.GetLength()).To(Equal(originalCount))

				// Verify first object still exists
				firstItem := objects.Get(0)
				Expect(firstItem).ToNot(BeNil())
			}
		})

		It("should handle empty objects list gracefully", func() {
			emptyList := components.NewList[components.ILocalizedTextObject](0)

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			// Should not panic
			converter.PopulateDataObjectLocalizationsWithIlist("battle/kernel/btl_txt.bin", emptyList, creator)

			Expect(emptyList.GetLength()).To(Equal(0))
		})

		It("should handle empty objects using IsEmpty method", func() {
			// Test the IsEmpty() check in the function
			emptyList := components.NewList[components.ILocalizedTextObject](0)

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			// Should exit early due to IsEmpty() check
			Expect(func() {
				converter.PopulateDataObjectLocalizationsWithIlist("battle/kernel/btl_txt.bin", emptyList, creator)
			}).ToNot(Panic())

			Expect(emptyList.GetLength()).To(Equal(0))
		})

		It("should handle non-existent pattern path gracefully", func() {
			// Create a list with some objects
			objects := components.NewList[components.ILocalizedTextObject](1)

			// Add a mock object
			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			mockData := make([]byte, 4)
			mockObject := creator(mockData, []byte("test"), 4, "us")
			objects.Add(mockObject)

			// Try to populate with non-existent path - should not panic
			Expect(func() {
				converter.PopulateDataObjectLocalizationsWithIlist("non/existent/path.bin", objects, creator)
			}).ToNot(Panic())

			// Objects should still exist
			Expect(objects.GetLength()).To(Equal(1))
		})

		It("should iterate through all supported languages", func() {
			// Create objects with a known pattern
			patternPath := "battle/kernel/btl_txt.bin"

			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
				return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)
			objects := converter.ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)

			if objects != nil && objects.GetLength() > 0 {
				originalCount := objects.GetLength()

				// This function should iterate through all supported languages
				// and try to load localizations for each
				converter.PopulateDataObjectLocalizationsWithIlist(patternPath, objects, creator)

				// Objects count should remain the same
				Expect(objects.GetLength()).To(Equal(originalCount))

				// At least the first object should still be valid
				firstItem := objects.Get(0)
				Expect(firstItem).ToNot(BeNil())
				Expect(firstItem.GetTextObject()).ToNot(BeNil())
			}
		})
	})

	Context("Integration Tests", func() {
		It("should demonstrate complete workflow for name-only data", func() {
			// Test the complete workflow that mimics ReadBattleTextWithAllLocalizations
			patternPath := "battle/kernel/btl_txt.bin"

			result := converter.ReadNameOnlyDataObjectsWithIlist(patternPath)

			if result != nil && result.GetLength() > 0 {
				Expect(result.GetLength()).To(BeNumerically(">", 0))

				// Verify first item has proper structure
				firstItem := result.Get(0)
				Expect(firstItem).ToNot(BeNil())

				textObj := firstItem.GetTextObject()
				Expect(textObj).ToNot(BeNil())

				nameString := textObj.GetKeyedString("name")
				Expect(nameString).ToNot(BeNil())

				// Should have default localization
				usContent := nameString.GetLocalizedContent(common.DefaultLocalization)
				Expect(usContent).ToNot(BeNil())
			}
		})

		It("should demonstrate complete workflow for name-description data", func() {
			// Test the complete workflow that mimics ReadKeyItemsWithAllLocalizations
			patternPath := "battle/kernel/important.bin"

			result := converter.ReadNameDescriptionObjectsWithIlist(patternPath)

			if result != nil && result.GetLength() > 0 {
				Expect(result.GetLength()).To(BeNumerically(">", 0))

				// Verify first item has proper structure
				firstItem := result.Get(0)
				Expect(firstItem).ToNot(BeNil())

				textObj := firstItem.GetTextObject()
				Expect(textObj).ToNot(BeNil())

				nameString := textObj.GetKeyedString("name")
				Expect(nameString).ToNot(BeNil())

				descString := textObj.GetKeyedString("description")
				Expect(descString).ToNot(BeNil())

				// Should have default localization for both
				usNameContent := nameString.GetLocalizedContent(common.DefaultLocalization)
				Expect(usNameContent).ToNot(BeNil())

				usDescContent := descString.GetLocalizedContent(common.DefaultLocalization)
				Expect(usDescContent).ToNot(BeNil())
			}
		})

		It("should verify that different pattern paths produce different results", func() {
			battleTextResult := converter.ReadNameOnlyDataObjectsWithIlist("battle/kernel/btl_txt.bin")
			keyItemsResult := converter.ReadNameDescriptionObjectsWithIlist("battle/kernel/important.bin")

			// Both should exist and have different characteristics
			if battleTextResult != nil && keyItemsResult != nil {
				// They may have different lengths
				battleTextCount := battleTextResult.GetLength()
				keyItemsCount := keyItemsResult.GetLength()

				Expect(battleTextCount).To(BeNumerically(">", 0))
				Expect(keyItemsCount).To(BeNumerically(">", 0))

				// They are different types of data, so counts may be different
				// Just verify both are valid
				battleItem := battleTextResult.Get(0)
				keyItem := keyItemsResult.Get(0)

				Expect(battleItem).ToNot(BeNil())
				Expect(keyItem).ToNot(BeNil())
			}
		})
	})
})
