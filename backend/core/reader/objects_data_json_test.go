package reader_test

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestObjectsDataJSON(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ObjectsDataJSON Suite")
}

var _ = Describe("ObjectsDataJSON", Ordered, func() {
	var (
		//rootDir               string
		originalResourcesRoot string
		originalGameFilesRoot string
		tempEditDir           string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetGameVersion(1)
		common.SetVerboseMode(false)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	BeforeEach(func() {
		// Save original paths
		originalResourcesRoot = common.ResourcesRoot
		originalGameFilesRoot = common.GameFilesRoot

		//rootDir = testcommon.GetTestDataRootDirectory()
		//Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

		//common.SetGameFilesRoot(filepath.Join(rootDir, "FFX", "binary"))

		// Create temporary edit directory for test JSON files
		tempEditDir = filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
		Expect(os.MkdirAll(tempEditDir, 0755)).To(Succeed())
	})

	AfterEach(func() {
		// Restore original paths
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot

		// Clean up test files
		if tempEditDir != "" {
			os.RemoveAll(tempEditDir)
		}

		// Clear global state
		components.KEY_ITEMS = nil
		components.COMMANDS = nil
		components.ByteToCharMaps = make(map[string]map[uint]rune)
		components.CharToByteMaps = make(map[string]map[rune]uint)
		components.MacroLookup = make(map[int]*components.LocalizedMacroStringObject)
	})

	Context("EditAndSaveKeyItemsJSONFiles", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())

			components.KEY_ITEMS = reader.ReadKeyItemsWithAllLocalizations()
		})
		It("should process key items JSON file without error and have 64 items", func() {
			Expect(components.KEY_ITEMS).To(HaveLen(64))

			// Create test JSON file
			testData := []reader.ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Key Item",
					},
					Description: map[string]string{
						"us": "Test Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "key_items_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			// Test the function
			Expect(reader.ProcessKeyItemsJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(components.KEY_ITEMS).To(HaveLen(64))

			testData := []reader.ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Modified Key Item Name",
					},
					Description: map[string]string{
						"us": "Modified Key Item Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "key_items_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			// Get original name for comparison
			originalItem := components.KEY_ITEMS[0]
			Expect(originalItem).ToNot(BeNil())

			originalNameUs := ""
			if originalItem.GetNameDescriptionTextObject() != nil &&
				originalItem.GetNameDescriptionTextObject().Name != nil {
				if content := originalItem.GetNameDescriptionTextObject().Name.GetLocalizedContent("us"); content != nil {
					originalNameUs = content.GetString()
				}
			}

			// Process the JSON file
			Expect(reader.ProcessKeyItemsJsonFile()).To(Succeed())

			// Validate the item was modified
			modifiedItem := components.KEY_ITEMS[0]
			Expect(modifiedItem).ToNot(BeNil())
			Expect(modifiedItem.GetNameDescriptionTextObject()).ToNot(BeNil())
			Expect(modifiedItem.GetNameDescriptionTextObject().Name).ToNot(BeNil())

			usContent := modifiedItem.GetNameDescriptionTextObject().Name.GetLocalizedContent("us")
			Expect(usContent).ToNot(BeNil())
			newNameUs := usContent.GetString()

			// Verify the name was actually changed
			Expect(newNameUs).To(Equal("Modified Key Item Name"))
			Expect(newNameUs).ToNot(Equal(originalNameUs))
		})

		It("should return error when JSON file does not exist", func() {
			// Don't create the JSON file
			err := reader.ProcessKeyItemsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("EditAndSaveCommandJSONFiles", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())

			components.COMMANDS = reader.ReadCommandsWithAllLocalizations()
		})

		It("should process commands JSON file without error and have 320 commands", func() {
			Expect(components.COMMANDS).To(HaveLen(320))

			// Create test JSON file
			testData := []reader.ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Command",
					},
					Description: map[string]string{
						"us": "Test Command Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "commands_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			// Test the function
			Expect(reader.ProcessCommandJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(components.COMMANDS).To(HaveLen(320))

			// Create test JSON with modified content
			testData := []reader.ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Modified Command Name",
					},
					Description: map[string]string{
						"us": "Modified Command Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "commands_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			// Get original name for comparison
			originalCommand := components.COMMANDS[0]
			Expect(originalCommand).ToNot(BeNil())

			originalNameUs := ""
			if originalCommand.GetNameDescriptionTextObject() != nil &&
				originalCommand.GetNameDescriptionTextObject().Name != nil {
				if content := originalCommand.GetNameDescriptionTextObject().Name.GetLocalizedContent("us"); content != nil {
					originalNameUs = content.GetString()
				}
			}

			// Process the JSON file
			Expect(reader.ProcessCommandJsonFile()).To(Succeed())

			// Validate the command was modified
			modifiedCommand := components.COMMANDS[0]
			Expect(modifiedCommand).ToNot(BeNil())
			Expect(modifiedCommand.GetNameDescriptionTextObject()).ToNot(BeNil())
			Expect(modifiedCommand.GetNameDescriptionTextObject().Name).ToNot(BeNil())

			usContent := modifiedCommand.GetNameDescriptionTextObject().Name.GetLocalizedContent("us")
			Expect(usContent).ToNot(BeNil())
			newNameUs := usContent.GetString()

			// Verify the name was actually changed
			Expect(newNameUs).To(Equal("Modified Command Name"))
			Expect(newNameUs).ToNot(Equal(originalNameUs))
		})

		It("should return error when JSON file does not exist", func() {
			// Don't create the JSON file
			err := reader.ProcessCommandJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})
})
