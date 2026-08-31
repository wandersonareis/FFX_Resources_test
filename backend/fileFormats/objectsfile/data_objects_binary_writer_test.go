package objectsfile_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/sharedutils"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDataObjectsBinaryWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Objects Binary Writer Suite")
}

var _ = Describe("Data Objects Binary Writer", Ordered, func() {
	var (
		originalResourcesRoot string
		originalGameFilesRoot string
		originalModsEnabled   bool
		originalVerboseMode   bool
		testModsDir           string
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

		// Enable mods for writing tests
		common.SetModsEnabled(true)
		common.SetVerboseMode(false)

		// Ensure internals are initialized before each test
		Expect(reader.InitializeInternals()).To(Succeed())

		// Create test mods directory
		testModsDir = filepath.Join(common.GameFilesRoot, common.ModsFolder)
		Expect(common.EnsurePathExists(testModsDir)).To(Succeed())
	})

	AfterEach(func() {
		// Restore original state
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot
		common.SetModsEnabled(originalModsEnabled)
		common.SetVerboseMode(originalVerboseMode)

		// Clear global state to avoid interference between tests
		datastore.KeyItems.Clear()
		datastore.Commands.Clear()
		datastore.Items.Clear()
		datastore.ArmsTxt.Clear()
		datastore.BattleTxt.Clear()
		datastore.BattleEndTxt.Clear()
		objectsfile.MONMAGIC1 = nil
		objectsfile.MONMAGIC2 = nil
		objectsfile.BUILD_TEXT = nil
		objectsfile.CONFIG_TEXT = nil
		objectsfile.ITEM_TEXT = nil
		objectsfile.MMAIN_TEXT = nil
		objectsfile.PLAYER_ROOM = nil
		objectsfile.NAME_TEXT = nil
	})

	// Helper function to clean up test files and empty directories
	cleanupTestFiles := func(patternPath string) {
		for localizationKey := range common.SupportedLanguages {
			localizationRoot := common.GetLocalizationRoot(localizationKey)
			filePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, patternPath)
			filePath = filepath.FromSlash(filePath)

			if common.IsPathExists(filePath) {
				os.Remove(filePath)
			}

			// Remove empty directories
			dir := filepath.Dir(filePath)
			for dir != testModsDir && dir != "." {
				if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
					os.Remove(dir)
					dir = filepath.Dir(dir)
				} else {
					break
				}
			}
		}
	}

	Context("ExportLocalizedTextData - Name Only Objects", func() {
		var patternPath string

		BeforeEach(func() {
			patternPath = "battle/kernel/btl_txt.bin"
		})

		AfterEach(func() {
			cleanupTestFiles(patternPath)
		})

		It("should export and read back unchanged name-only data", func() {
			// Read original data
			originalData := objectsfile.ReadNameOnlyDataObjectsWithIlist(patternPath)
			Expect(originalData).ToNot(BeNil())
			Expect(originalData.Len()).To(BeNumerically(">", 0))

			// Get original first object's name
			originalFirstObj := originalData.Get(0)
			originalNameString := originalFirstObj.GetTextObject().GetKeyedString("name")
			originalNameContent := originalNameString.GetLocalizedContent(common.DefaultLocalization)
			originalName := originalNameContent.GetString()

			// Export the data
			err := objectsfile.ExportLocalizedTextData(originalData, patternPath)
			Expect(err).To(BeNil())

			// Read back the exported data
			exportedData := objectsfile.ReadNameOnlyDataObjectsWithIlist(patternPath)
			Expect(exportedData).ToNot(BeNil())
			Expect(exportedData.Len()).To(Equal(originalData.Len()))

			// Compare first object's name
			exportedFirstObj := exportedData.Get(0)
			exportedNameString := exportedFirstObj.GetTextObject().GetKeyedString("name")
			exportedNameContent := exportedNameString.GetLocalizedContent(common.DefaultLocalization)
			exportedName := exportedNameContent.GetString()

			Expect(exportedName).To(Equal(originalName))
		})

		It("should export modified name-only data correctly", func() {
			// Read original data
			originalData := objectsfile.ReadNameOnlyDataObjectsWithIlist(patternPath)
			Expect(originalData).ToNot(BeNil())
			Expect(originalData.Len()).To(BeNumerically(">", 0))

			// Create JSON file for modification
			jsonFileName := "test_btl_txt.json"
			jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", common.WithVersionSuffix(jsonFileName))
			editsDir := filepath.Dir(jsonFilePath)
			Expect(common.EnsurePathExists(editsDir)).To(Succeed())

			// Create test JSON data to modify index 0
			testName := "Modified Test Name"
			jsonContent := `[
				{
					"index": 0,
					"name": {
						"us": "` + testName + `"
					}
				}
			]`

			err := os.WriteFile(jsonFilePath, []byte(jsonContent), 0644)
			Expect(err).To(BeNil())

			// Import modifications
			err = objectsfile.ImportLocalizedDataFromJsonFile(jsonFileName, originalData)
			Expect(err).To(BeNil())

			// Export the modified data
			err = objectsfile.ExportLocalizedTextData(originalData, patternPath)
			Expect(err).To(BeNil())

			// Read back the exported data
			exportedData := objectsfile.ReadNameOnlyDataObjectsWithIlist(patternPath)
			Expect(exportedData).ToNot(BeNil())

			// Verify the modification was applied and persisted
			exportedFirstObj := exportedData.Get(0)
			exportedNameString := exportedFirstObj.GetTextObject().GetKeyedString("name")
			exportedNameContent := exportedNameString.GetLocalizedContent(common.DefaultLocalization)
			exportedName := exportedNameContent.GetString()

			Expect(exportedName).To(Equal(testName))

			// Cleanup JSON file
			os.Remove(jsonFilePath)
			if entries, err := os.ReadDir(editsDir); err == nil && len(entries) == 0 {
				os.Remove(editsDir)
			}
		})
	})

	Context("ExportLocalizedTextData - Name Description Objects", func() {
		var patternPath string

		BeforeEach(func() {
			patternPath = "battle/kernel/important.bin"
		})

		AfterEach(func() {
			cleanupTestFiles(patternPath)
		})

		It("should export and read back unchanged name-description data", func() {
			// Read original data
			originalData := objectsfile.ReadNameDescriptionObjectsWithIlist(patternPath)
			Expect(originalData).ToNot(BeNil())
			Expect(originalData.Len()).To(BeNumerically(">", 0))

			// Get original first object's name and description
			originalFirstObj := originalData.Get(0)
			originalNameString := originalFirstObj.GetTextObject().GetKeyedString("name")
			originalNameContent := originalNameString.GetLocalizedContent(common.DefaultLocalization)
			originalName := originalNameContent.GetString()

			originalDescString := originalFirstObj.GetTextObject().GetKeyedString("description")
			originalDescContent := originalDescString.GetLocalizedContent(common.DefaultLocalization)
			originalDesc := originalDescContent.GetString()

			// Export the data
			err := objectsfile.ExportLocalizedTextData(originalData, patternPath)
			Expect(err).To(BeNil())

			// Read back the exported data
			exportedData := objectsfile.ReadNameDescriptionObjectsWithIlist(patternPath)
			Expect(exportedData).ToNot(BeNil())
			Expect(exportedData.Len()).To(Equal(originalData.Len()))

			// Compare first object's name and description
			exportedFirstObj := exportedData.Get(0)
			exportedNameString := exportedFirstObj.GetTextObject().GetKeyedString("name")
			exportedNameContent := exportedNameString.GetLocalizedContent(common.DefaultLocalization)
			exportedName := exportedNameContent.GetString()

			exportedDescString := exportedFirstObj.GetTextObject().GetKeyedString("description")
			exportedDescContent := exportedDescString.GetLocalizedContent(common.DefaultLocalization)
			exportedDesc := exportedDescContent.GetString()

			Expect(exportedName).To(Equal(originalName))
			Expect(exportedDesc).To(Equal(originalDesc))
		})

		It("should export modified name-description data correctly", func() {
			// Read original data
			originalData := objectsfile.ReadNameDescriptionObjectsWithIlist(patternPath)
			Expect(originalData).ToNot(BeNil())
			Expect(originalData.Len()).To(BeNumerically(">", 0))

			// Create JSON file for modification
			jsonFileName := "test_important.json"
			jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", jsonFileName)
			editsDir := filepath.Dir(jsonFilePath)
			Expect(common.EnsurePathExists(editsDir)).To(Succeed())

			// Create test JSON data to modify index 0
			testName := "Modified Test Item"
			testDesc := "Modified test description for item"
			jsonContent := `[
				{
					"index": 0,
					"name": {
						"us": "` + testName + `"
					},
					"description": {
						"us": "` + testDesc + `"
					}
				}
			]`

			err := os.WriteFile(jsonFilePath, []byte(jsonContent), 0644)
			Expect(err).To(BeNil())

			// Import modifications
			err = objectsfile.ImportLocalizedDataFromJsonFile(jsonFileName, originalData)
			Expect(err).To(BeNil())

			// Export the modified data
			err = objectsfile.ExportLocalizedTextData(originalData, patternPath)
			Expect(err).To(BeNil())

			// Read back the exported data
			exportedData := objectsfile.ReadNameDescriptionObjectsWithIlist(patternPath)
			Expect(exportedData).ToNot(BeNil())

			// Verify the modifications were applied and persisted
			exportedFirstObj := exportedData.Get(0)
			exportedNameString := exportedFirstObj.GetTextObject().GetKeyedString("name")
			exportedNameContent := exportedNameString.GetLocalizedContent(common.DefaultLocalization)
			exportedName := exportedNameContent.GetString()

			exportedDescString := exportedFirstObj.GetTextObject().GetKeyedString("description")
			exportedDescContent := exportedDescString.GetLocalizedContent(common.DefaultLocalization)
			exportedDesc := exportedDescContent.GetString()

			Expect(exportedName).To(Equal(testName))
			Expect(exportedDesc).To(Equal(testDesc))

			// Cleanup JSON file
			os.Remove(jsonFilePath)
			if entries, err := os.ReadDir(editsDir); err == nil && len(entries) == 0 {
				os.Remove(editsDir)
			}
		})
	})

	Context("ConvertFFXLocalizedDataToBytes - Memory Operations", func() {
		It("should convert and parse back unchanged name-only data in memory", func() {
			// Read original binary data from file
			patternPath := "battle/kernel/btl_txt.bin"
			filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

			fileAccessor, err := common.NewFileAccessor(filePath)
			Expect(err).To(BeNil())
			Expect(fileAccessor.Exists).To(BeTrue())

			originalBinaryData, err := os.ReadFile(fileAccessor.ResolvedPath)
			Expect(err).To(BeNil())

			// Parse binary data into objects
			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
				return objectsfile.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
			}

			originalObjects := objectsfile.ParseDataListWithIlist(originalBinaryData, common.DefaultLocalization, creator)
			Expect(originalObjects).ToNot(BeNil())
			Expect(originalObjects.Len()).To(BeNumerically(">", 0))

			// Get original first object's name for comparison
			originalFirstObj := originalObjects.Get(0)
			originalNameString := originalFirstObj.GetTextObject().GetKeyedString("name")
			originalNameContent := originalNameString.GetLocalizedContent(common.DefaultLocalization)
			originalName := originalNameContent.GetString()

			// Convert objects back to binary data
			objectsSlice := make([]datastore.IGlobalLocalizedTextObject, originalObjects.Len())
			for i, obj := range originalObjects.Items() {
				objectsSlice[i] = obj
			}

			convertedBinaryData, err := objectsfile.ConvertFFXLocalizedDataToBytes(objectsSlice, 0, len(objectsSlice), common.DefaultLocalization)
			Expect(err).To(BeNil())
			Expect(len(convertedBinaryData)).To(BeNumerically(">", 0))

			// Parse the converted binary data back into objects
			parsedObjects := objectsfile.ParseDataListWithIlist(convertedBinaryData, common.DefaultLocalization, creator)
			Expect(parsedObjects).ToNot(BeNil())
			Expect(parsedObjects.Len()).To(Equal(originalObjects.Len()))

			// Compare first object's name to ensure round-trip conversion worked
			parsedFirstObj := parsedObjects.Get(0)
			parsedNameString := parsedFirstObj.GetTextObject().GetKeyedString("name")
			parsedNameContent := parsedNameString.GetLocalizedContent(common.DefaultLocalization)
			parsedName := parsedNameContent.GetString()

			Expect(parsedName).To(Equal(originalName))
		})

		It("should convert and parse back modified name-description data in memory", func() {
			// Read original binary data from file
			patternPath := "battle/kernel/important.bin"
			filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

			fileAccessor, err := common.NewFileAccessor(filePath)
			Expect(err).To(BeNil())
			Expect(fileAccessor.Exists).To(BeTrue())

			originalBinaryData, err := os.ReadFile(fileAccessor.ResolvedPath)
			Expect(err).To(BeNil())

			// Parse binary data into objects
			creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
				return objectsfile.NewNameDescriptionTextObject(data, stringBytes, headerLength, localization)
			}

			originalObjects := objectsfile.ParseDataListWithIlist(originalBinaryData, common.DefaultLocalization, creator)
			Expect(originalObjects).ToNot(BeNil())
			Expect(originalObjects.Len()).To(BeNumerically(">", 0)) // Modify the first object's name and description in memory
			testName := "Memory Test Item"
			testDesc := "Memory test description for item"

			firstObj := originalObjects.Get(0)
			nameString := firstObj.GetTextObject().GetKeyedString("name")
			nameContent := nameString.GetLocalizedContent(common.DefaultLocalization)
			charset := sharedutils.GetCharsetForLanguage(common.DefaultLocalization)
			nameContent.SetString(testName, charset)

			descString := firstObj.GetTextObject().GetKeyedString("description")
			descContent := descString.GetLocalizedContent(common.DefaultLocalization)
			descContent.SetString(testDesc, charset)

			// Convert modified objects to binary data
			objectsSlice := make([]datastore.IGlobalLocalizedTextObject, originalObjects.Len())
			for i, obj := range originalObjects.Items() {
				objectsSlice[i] = obj
			}

			convertedBinaryData, err := objectsfile.ConvertFFXLocalizedDataToBytes(objectsSlice, 0, len(objectsSlice), common.DefaultLocalization)
			Expect(err).To(BeNil())
			Expect(len(convertedBinaryData)).To(BeNumerically(">", 0))

			// Parse the converted binary data back into objects
			parsedObjects := objectsfile.ParseDataListWithIlist(convertedBinaryData, common.DefaultLocalization, creator)
			Expect(parsedObjects).ToNot(BeNil())
			Expect(parsedObjects.Len()).To(Equal(originalObjects.Len()))

			// Verify the modifications were preserved in the round-trip conversion
			parsedFirstObj := parsedObjects.Get(0)
			parsedNameString := parsedFirstObj.GetTextObject().GetKeyedString("name")
			parsedNameContent := parsedNameString.GetLocalizedContent(common.DefaultLocalization)
			parsedName := parsedNameContent.GetString()

			parsedDescString := parsedFirstObj.GetTextObject().GetKeyedString("description")
			parsedDescContent := parsedDescString.GetLocalizedContent(common.DefaultLocalization)
			parsedDesc := parsedDescContent.GetString()

			Expect(parsedName).To(Equal(testName))
			Expect(parsedDesc).To(Equal(testDesc))
		})
	})
})
