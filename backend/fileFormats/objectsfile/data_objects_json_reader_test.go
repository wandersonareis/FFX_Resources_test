package objectsfile_test

import (
	"encoding/json"
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

func TestDataObjectsJSONReaders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Objects JSON Readers Suite")
}

// ObjectsData represents the JSON structure for test data
type ObjectsData struct {
	ID          int               `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
}

var _ = Describe("Data Objects JSON Readers", Ordered, func() {
	var (
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
		sharedutils.ByteToCharMaps = make(map[string]map[uint]rune)
		sharedutils.CharToByteMaps = make(map[string]map[rune]uint)
		//components.MacroLookup = make(map[int]*macrodic.LocalizedMacroStringObject)
	})

	Context("ProcessKeyItemsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadKeyItemsWithAllLocalizations()
		})

		It("should process key items JSON file without error", func() {
			Expect(datastore.KeyItems.Len()).To(Equal(64))

			// Create test JSON file
			testData := []ObjectsData{
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
			Expect(objectsfile.ProcessKeyItemsJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(datastore.KeyItems.Len()).To(Equal(64))

			testData := []ObjectsData{
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
			originalItem := datastore.KeyItems.Get(0)
			Expect(originalItem).ToNot(BeNil())

			originalNameUs := ""

			// Process the JSON file
			Expect(objectsfile.ProcessKeyItemsJsonFile()).To(Succeed())

			// Validate the item was modified
			modifiedItem := datastore.KeyItems.Get(0)
			Expect(modifiedItem).ToNot(BeNil())
			Expect(modifiedItem.GetTextObject()).ToNot(BeNil())
			Expect(modifiedItem.GetTextObject().GetKeyedString("name")).ToNot(BeNil())

			usContent := modifiedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(usContent).ToNot(BeNil())
			newNameUs := usContent.GetString()

			// Verify the name was actually changed
			Expect(newNameUs).To(Equal("Modified Key Item Name"))
			Expect(newNameUs).ToNot(Equal(originalNameUs))
		})

		It("should return error when JSON file does not exist", func() {
			// Don't create the JSON file
			err := objectsfile.ProcessKeyItemsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessCommandsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadCommandsWithAllLocalizations()
		})

		It("should process commands JSON file without error", func() {
			Expect(datastore.Commands.Len()).To(Equal(320))

			testData := []ObjectsData{
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

			Expect(objectsfile.ProcessCommandsJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(datastore.Commands.Len()).To(Equal(320))

			testData := []ObjectsData{
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
			originalCommand := datastore.Commands.Get(0)
			Expect(originalCommand).ToNot(BeNil())

			originalNameUs := ""

			// Process the JSON file
			Expect(objectsfile.ProcessCommandsJsonFile()).To(Succeed())

			// Validate the command was modified
			modifiedCommand := datastore.Commands.Get(0)
			Expect(modifiedCommand).ToNot(BeNil())
			Expect(modifiedCommand.GetTextObject()).ToNot(BeNil())
			Expect(modifiedCommand.GetTextObject().GetKeyedString("name")).ToNot(BeNil())

			usContent := modifiedCommand.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(usContent).ToNot(BeNil())
			newNameUs := usContent.GetString()

			// Verify the name was actually changed
			Expect(newNameUs).To(Equal("Modified Command Name"))
			Expect(newNameUs).ToNot(Equal(originalNameUs))
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessCommandsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessItemsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadItemsWithAllLocalizations()
		})

		It("should process items JSON file without error", func() {
			Expect(datastore.Items.Len()).To(Equal(0))

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Item",
					},
					Description: map[string]string{
						"us": "Test Item Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "items_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessItemsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessItemsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessArmsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadArmsTextWithAllLocalizations()
		})

		It("should process arms JSON file without error", func() {
			Expect(datastore.ArmsTxt.Len()).To(Equal(0))

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Arms",
					},
					Description: map[string]string{
						"us": "Test Arms Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "arms_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessArmsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessArmsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBattleTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBattleTextWithAllLocalizations()
		})

		It("should process battle text JSON file without error", func() {
			Expect(datastore.BattleTxt.Len()).To(Equal(0))

			// Battle text is name-only data
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Battle Text",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "battle_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessBattleTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessBattleTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBattleEndTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBattleEndTextWithAllLocalizations()
		})

		It("should process battle end text JSON file without error", func() {
			Expect(datastore.BattleEndTxt.Len()).To(Equal(0))

			// Battle end text is name-only data (due to bug mentioned in data_writers.go)
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Battle End Text",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "battle_end_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessBattleEndTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessBattleEndTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMonsterMagic1JsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMonsterMagic1WithAllLocalizations()
		})

		It("should process monster magic 1 JSON file without error", func() {
			Expect(objectsfile.MONMAGIC1.Items()).ToNot(BeEmpty())

			// Monster magic is name-only data
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test MonMagic1",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "monster_magic1_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessMonsterMagic1JsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessMonsterMagic1JsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMonsterMagic2JsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMonsterMagic2WithAllLocalizations()
		})

		It("should process monster magic 2 JSON file without error", func() {
			Expect(objectsfile.MONMAGIC2.Items()).ToNot(BeEmpty())

			// Monster magic is name-only data
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test MonMagic2",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "monster_magic2_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessMonsterMagic2JsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessMonsterMagic2JsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBuildTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBuildTextWithAllLocalizations()
		})

		It("should process build text JSON file without error", func() {
			Expect(objectsfile.BUILD_TEXT.Items()).ToNot(BeEmpty())

			// Build text is name-only data (due to bug mentioned in data_writers.go)
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Build Text",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "build_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessBuildTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessBuildTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessConfigTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadConfigTextWithAllLocalizations()
		})

		It("should process config text JSON file without error", func() {
			Expect(objectsfile.CONFIG_TEXT.Items()).ToNot(BeEmpty())

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Config Text",
					},
					Description: map[string]string{
						"us": "Test Config Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "config_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessConfigTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessConfigTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessItemCommandsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadItemCommandsWithAllLocalizations()
		})

		It("should process item commands JSON file without error", func() {
			Expect(objectsfile.ITEM_TEXT.Items()).ToNot(BeEmpty())

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Item Text",
					},
					Description: map[string]string{
						"us": "Test Item Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "item_commands_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessItemCommandsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessItemCommandsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMainMenuTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMainMenuTextWithAllLocalizations()
		})

		It("should process main menu text JSON file without error", func() {
			Expect(objectsfile.MMAIN_TEXT.Items()).ToNot(BeEmpty())

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Main Menu Text",
					},
					Description: map[string]string{
						"us": "Test Main Menu Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "main_menu_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessMainMenuTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessMainMenuTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessPlayerRoomTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadPlayerRomTextWithAllLocalizations()
		})

		It("should process player room text JSON file without error", func() {
			Expect(objectsfile.PLAYER_ROOM.Items()).ToNot(BeEmpty())

			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Player Room Text",
					},
					Description: map[string]string{
						"us": "Test Player Room Description",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "player_room_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessPlayerRoomTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessPlayerRoomTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessNameTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadNameTextWithAllLocalizations()
		})

		It("should process name text JSON file without error", func() {
			Expect(objectsfile.NAME_TEXT.Items()).ToNot(BeEmpty())

			// Name text is name-only data
			testData := []ObjectsData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Name Text",
					},
				},
			}

			jsonPath := filepath.Join(tempEditDir, "names_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())

			Expect(objectsfile.ProcessNameTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := objectsfile.ProcessNameTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})
})
