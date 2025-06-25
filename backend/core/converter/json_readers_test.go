package converter_test

import (
	"encoding/json"
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

func TestJSONReaders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JSON Readers Suite")
}

// ObjectsData represents the JSON structure for test data
type ObjectsData struct {
	ID          int               `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
}

var _ = Describe("JSON Readers", Ordered, func() {
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

	Context("ProcessKeyItemsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadKeyItemsWithAllLocalizations()
		})

		It("should process key items JSON file without error", func() {
			Expect(components.KEY_ITEMS.GetItems()).To(HaveLen(64))

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
			Expect(converter.ProcessKeyItemsJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(components.KEY_ITEMS.GetItems()).To(HaveLen(64))

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
			originalItem := components.KEY_ITEMS.Get(0)
			Expect(originalItem).ToNot(BeNil())

			originalNameUs := ""
			if original := originalItem.GetTextObject(); original != nil {
				if name := original.GetKeyedString("name"); name != nil {
					if content := name.GetLocalizedContent("us"); content != nil {
						originalNameUs = content.GetString()
					}
				}
			}

			// Process the JSON file
			Expect(converter.ProcessKeyItemsJsonFile()).To(Succeed())

			// Validate the item was modified
			modifiedItem := components.KEY_ITEMS.Get(0)
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
			err := converter.ProcessKeyItemsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessCommandsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadCommandsWithAllLocalizations()
		})

		It("should process commands JSON file without error", func() {
			Expect(components.COMMANDS.GetItems()).To(HaveLen(320))

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

			Expect(converter.ProcessCommandsJsonFile()).To(Succeed())
		})

		It("should edit and validate array has edited item", func() {
			Expect(components.COMMANDS.GetItems()).To(HaveLen(320))

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
			originalCommand := components.COMMANDS.Get(0)
			Expect(originalCommand).ToNot(BeNil())

			originalNameUs := ""
			if original := originalCommand.GetTextObject(); original != nil {
				if name := original.GetKeyedString("name"); name != nil {
					if content := name.GetLocalizedContent("us"); content != nil {
						originalNameUs = content.GetString()
					}
				}
			}

			// Process the JSON file
			Expect(converter.ProcessCommandsJsonFile()).To(Succeed())

			// Validate the command was modified
			modifiedCommand := components.COMMANDS.Get(0)
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
			err := converter.ProcessCommandsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessItemsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadItemsWithAllLocalizations()
		})

		It("should process items JSON file without error", func() {
			Expect(components.ITEMS.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessItemsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessItemsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessArmsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadArmsTextWithAllLocalizations()
		})

		It("should process arms JSON file without error", func() {
			Expect(components.ARMS_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessArmsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessArmsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBattleTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBattleTextWithAllLocalizations()
		})

		It("should process battle text JSON file without error", func() {
			Expect(components.BTL_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessBattleTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessBattleTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBattleEndTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBattleEndTextWithAllLocalizations()
		})

		It("should process battle end text JSON file without error", func() {
			Expect(components.BTLEND_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessBattleEndTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessBattleEndTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMonsterMagic1JsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMonsterMagic1WithAllLocalizations()
		})

		It("should process monster magic 1 JSON file without error", func() {
			Expect(components.MONMAGIC1.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessMonsterMagic1JsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessMonsterMagic1JsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMonsterMagic2JsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMonsterMagic2WithAllLocalizations()
		})

		It("should process monster magic 2 JSON file without error", func() {
			Expect(components.MONMAGIC2.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessMonsterMagic2JsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessMonsterMagic2JsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessBuildTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBuildTextWithAllLocalizations()
		})

		It("should process build text JSON file without error", func() {
			Expect(components.BUILD_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessBuildTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessBuildTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessConfigTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadConfigTextWithAllLocalizations()
		})

		It("should process config text JSON file without error", func() {
			Expect(components.CONFIG_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessConfigTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessConfigTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessItemCommandsJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadItemCommandsWithAllLocalizations()
		})

		It("should process item commands JSON file without error", func() {
			Expect(components.ITEM_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessItemCommandsJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessItemCommandsJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessMainMenuTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMainMenuTextWithAllLocalizations()
		})

		It("should process main menu text JSON file without error", func() {
			Expect(components.MMAIN_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessMainMenuTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessMainMenuTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessPlayerRoomTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadPlayerRomTextWithAllLocalizations()
		})

		It("should process player room text JSON file without error", func() {
			Expect(components.PLAYER_ROOM.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessPlayerRoomTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessPlayerRoomTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessNameTextJsonFile", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadNameTextWithAllLocalizations()
		})

		It("should process name text JSON file without error", func() {
			Expect(components.NAME_TEXT.GetItems()).ToNot(BeEmpty())

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

			Expect(converter.ProcessNameTextJsonFile()).To(Succeed())
		})

		It("should return error when JSON file does not exist", func() {
			err := converter.ProcessNameTextJsonFile()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})
})
