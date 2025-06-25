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

func TestDataWriters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Writers Suite")
}

var _ = Describe("Data Writers", Ordered, func() {
	var (
		originalResourcesRoot string
		originalGameFilesRoot string
		originalModsEnabled   bool
		tempEditDir           string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetGameVersion(1)
		common.SetVerboseMode(false)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	BeforeEach(func() {
		// Save original paths and mods state
		originalResourcesRoot = common.ResourcesRoot
		originalGameFilesRoot = common.GameFilesRoot
		originalModsEnabled = common.AreModsEnabled()

		// Create temporary edit directory for test JSON files
		tempEditDir = filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
		Expect(os.MkdirAll(tempEditDir, 0755)).To(Succeed())
	})

	AfterEach(func() {
		// Restore original paths and mods state
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot
		common.SetModsEnabled(originalModsEnabled)

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

	Context("WriteAllKeyItemsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadKeyItemsWithAllLocalizations()
		})

		It("should write key items data to binary and verify persistence", func() {
			Expect(components.KEY_ITEMS.GetItems()).To(HaveLen(64))

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Key Item",
						},
					},
					Description: map[string]string{
						"us": "Test Write Description",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "key_items_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessKeyItemsJsonFile()).To(Succeed())

			// Verify modification in memory
			modifiedItem := components.KEY_ITEMS.Get(0)
			Expect(modifiedItem).ToNot(BeNil())
			usContent := modifiedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			usContentStr := usContent.GetString()
			Expect(usContentStr).To(Equal("Test Write Key Item"))

			// Write to binary files
			Expect(converter.WriteAllKeyItemsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.KEY_ITEMS = nil
			converter.ReadKeyItemsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItem := components.KEY_ITEMS.Get(0)
			Expect(reloadedItem).ToNot(BeNil())
			reloadedUsContent := reloadedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Key Item"))
		})

		It("should return error when no key items are loaded", func() {
			components.KEY_ITEMS = nil
			err := converter.WriteAllKeyItemsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no key items loaded"))
		})
	})

	Context("WriteAllCommandsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadCommandsWithAllLocalizations()
		})

		It("should write commands data to binary and verify persistence", func() {
			Expect(components.COMMANDS.GetItems()).To(HaveLen(320))

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Command",
						},
					},
					Description: map[string]string{
						"us": "Test Write Command Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "commands_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessCommandsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllCommandsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.COMMANDS = nil
			converter.ReadCommandsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedCommand := components.COMMANDS.Get(0)
			Expect(reloadedCommand).ToNot(BeNil())
			reloadedUsContent := reloadedCommand.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Command"))
		})

		It("should return error when no commands are loaded", func() {
			components.COMMANDS = nil
			err := converter.WriteAllCommandsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no commands loaded"))
		})
	})

	Context("WriteAllItemsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadItemsWithAllLocalizations()
		})

		It("should write items data to binary and verify persistence", func() {
			Expect(components.ITEMS.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Item",
						},
					},
					Description: map[string]string{
						"us": "Test Write Item Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "items_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessItemsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllItemsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.ITEMS = nil
			converter.ReadItemsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItem := components.ITEMS.Get(0)
			Expect(reloadedItem).ToNot(BeNil())
			reloadedUsContent := reloadedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Item"))
		})

		It("should return error when no items are loaded", func() {
			components.ITEMS = nil
			err := converter.WriteAllItemsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no items loaded"))
		})
	})

	Context("WriteAllArmsTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadArmsTextWithAllLocalizations()
		})

		It("should write arms text data to binary and verify persistence", func() {
			Expect(components.ARMS_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Arms",
						},
					},
					Description: map[string]string{
						"us": "Test Write Arms Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "arms_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessArmsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllArmsTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.ARMS_TEXT = nil
			converter.ReadArmsTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedArms := components.ARMS_TEXT.Get(0)
			Expect(reloadedArms).ToNot(BeNil())
			reloadedUsContent := reloadedArms.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Arms"))
		})

		It("should return error when no arms text is loaded", func() {
			components.ARMS_TEXT = nil
			err := converter.WriteAllArmsTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no arms text loaded"))
		})
	})

	Context("WriteAllBattleTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBattleTextWithAllLocalizations()
		})

		It("should write battle text data to binary and verify persistence", func() {
			Expect(components.BTL_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for battle text)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write Battle",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "battle_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessBattleTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllBattleTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.BTL_TEXT = nil
			converter.ReadBattleTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBattle := components.BTL_TEXT.Get(0)
			Expect(reloadedBattle).ToNot(BeNil())
			reloadedUsContent := reloadedBattle.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Battle"))
		})

		It("should return error when no battle text is loaded", func() {
			components.BTL_TEXT = nil
			err := converter.WriteAllBattleTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no battle text loaded"))
		})
	})

	Context("WriteAllBattleEndTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBattleEndTextWithAllLocalizations()
		})

		It("should write battle end text data to binary and verify persistence", func() {
			Expect(components.BTLEND_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only due to bug mentioned in data_writers.go)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write Battle End",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "battle_end_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessBattleEndTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllBattleEndTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.BTLEND_TEXT = nil
			converter.ReadBattleEndTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBattleEnd := components.BTLEND_TEXT.Get(0)
			Expect(reloadedBattleEnd).ToNot(BeNil())
			reloadedUsContent := reloadedBattleEnd.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Battle End"))
		})

		It("should return error when no battle end text is loaded", func() {
			components.BTLEND_TEXT = nil
			err := converter.WriteAllBattleEndTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no battle end text loaded"))
		})
	})

	Context("WriteAllMonsterMagic1Data", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMonsterMagic1WithAllLocalizations()
		})

		It("should write monster magic 1 data to binary and verify persistence", func() {
			Expect(components.MONMAGIC1.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for monster magic)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write MonMagic1",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "monster_magic1_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessMonsterMagic1JsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllMonsterMagic1Data()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.MONMAGIC1 = nil
			converter.ReadMonsterMagic1WithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMonMagic := components.MONMAGIC1.Get(0)
			Expect(reloadedMonMagic).ToNot(BeNil())
			reloadedUsContent := reloadedMonMagic.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write MonMagic1"))
		})

		It("should return error when no monster magic 1 is loaded", func() {
			components.MONMAGIC1 = nil
			err := converter.WriteAllMonsterMagic1Data()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no monster magic 1 loaded"))
		})
	})

	Context("WriteAllMonsterMagic2Data", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMonsterMagic2WithAllLocalizations()
		})

		It("should write monster magic 2 data to binary and verify persistence", func() {
			Expect(components.MONMAGIC2.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for monster magic)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write MonMagic2",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "monster_magic2_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessMonsterMagic2JsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllMonsterMagic2Data()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.MONMAGIC2 = nil
			converter.ReadMonsterMagic2WithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMonMagic := components.MONMAGIC2.Get(0)
			Expect(reloadedMonMagic).ToNot(BeNil())
			reloadedUsContent := reloadedMonMagic.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write MonMagic2"))
		})

		It("should return error when no monster magic 2 is loaded", func() {
			components.MONMAGIC2 = nil
			err := converter.WriteAllMonsterMagic2Data()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no monster magic 2 loaded"))
		})
	})

	Context("WriteAllBuildTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadBuildTextWithAllLocalizations()
		})

		It("should write build text data to binary and verify persistence", func() {
			Expect(components.BUILD_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only due to bug mentioned in data_writers.go)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write Build",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "build_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessBuildTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllBuildTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.BUILD_TEXT = nil
			converter.ReadBuildTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBuild := components.BUILD_TEXT.Get(0)
			Expect(reloadedBuild).ToNot(BeNil())
			reloadedUsContent := reloadedBuild.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Build"))
		})

		It("should return error when no build text is loaded", func() {
			components.BUILD_TEXT = nil
			err := converter.WriteAllBuildTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no build text loaded"))
		})
	})

	Context("WriteAllConfigTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadConfigTextWithAllLocalizations()
		})

		It("should write config text data to binary and verify persistence", func() {
			Expect(components.CONFIG_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Config",
						},
					},
					Description: map[string]string{
						"us": "Test Write Config Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "config_text_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessConfigTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllConfigTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.CONFIG_TEXT = nil
			converter.ReadConfigTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedConfig := components.CONFIG_TEXT.Get(0)
			Expect(reloadedConfig).ToNot(BeNil())
			reloadedUsContent := reloadedConfig.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Config"))
		})

		It("should return error when no config text is loaded", func() {
			components.CONFIG_TEXT = nil
			err := converter.WriteAllConfigTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no config text loaded"))
		})
	})

	Context("WriteAllMainMenuTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadMainMenuTextWithAllLocalizations()
		})

		It("should write main menu text data to binary and verify persistence", func() {
			Expect(components.MMAIN_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Main Menu",
						},
					},
					Description: map[string]string{
						"us": "Test Write Main Menu Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "main_menu_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessMainMenuTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllMainMenuTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.MMAIN_TEXT = nil
			converter.ReadMainMenuTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMainMenu := components.MMAIN_TEXT.Get(0)
			Expect(reloadedMainMenu).ToNot(BeNil())
			reloadedUsContent := reloadedMainMenu.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Main Menu"))
		})

		It("should return error when no main menu text is loaded", func() {
			components.MMAIN_TEXT = nil
			err := converter.WriteAllMainMenuTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no main menu text loaded"))
		})
	})

	Context("WriteAllItemCommandsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadItemCommandsWithAllLocalizations()
		})

		It("should write item commands data to binary and verify persistence", func() {
			Expect(components.ITEM_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Item Commands",
						},
					},
					Description: map[string]string{
						"us": "Test Write Item Commands Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "item_commands_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessItemCommandsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllItemCommandsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.ITEM_TEXT = nil
			converter.ReadItemCommandsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItemText := components.ITEM_TEXT.Get(0)
			Expect(reloadedItemText).ToNot(BeNil())
			reloadedUsContent := reloadedItemText.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Item Commands"))
		})

		It("should return error when no item text is loaded", func() {
			components.ITEM_TEXT = nil
			err := converter.WriteAllItemCommandsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no item text loaded"))
		})
	})

	Context("WriteAllPlayerRoomTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadPlayerRomTextWithAllLocalizations()
		})

		It("should write player room text data to binary and verify persistence", func() {
			Expect(components.PLAYER_ROOM.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []converter.NameDescriptionData{
				{
					NameOnlyData: converter.NameOnlyData{
						ID: 0,
						Name: map[string]string{
							"us": "Test Write Player Room",
						},
					},
					Description: map[string]string{
						"us": "Test Write Player Room Desc",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "player_room_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessPlayerRoomTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllPlayerRoomTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.PLAYER_ROOM = nil
			converter.ReadPlayerRomTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedPlayerRoom := components.PLAYER_ROOM.Get(0)
			Expect(reloadedPlayerRoom).ToNot(BeNil())
			reloadedUsContent := reloadedPlayerRoom.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Player Room"))
		})

		It("should return error when no player room text is loaded", func() {
			components.PLAYER_ROOM = nil
			err := converter.WriteAllPlayerRoomTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no player room text loaded"))
		})
	})

	Context("WriteAllNameTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			converter.ReadNameTextWithAllLocalizations()
		})

		It("should write name text data to binary and verify persistence", func() {
			Expect(components.NAME_TEXT.GetItems()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for name text)
			testData := []converter.NameOnlyData{
				{
					ID: 0,
					Name: map[string]string{
						"us": "Test Write Name Text",
					},
				},
			}

			// Import JSON to modify memory
			jsonPath := filepath.Join(tempEditDir, "names_all_localizations.json")
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(converter.ProcessNameTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(converter.WriteAllNameTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			components.NAME_TEXT = nil
			converter.ReadNameTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedNameText := components.NAME_TEXT.Get(0)
			Expect(reloadedNameText).ToNot(BeNil())
			reloadedUsContent := reloadedNameText.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Name Text"))
		})

		It("should return error when no name text is loaded", func() {
			components.NAME_TEXT = nil
			err := converter.WriteAllNameTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no name text loaded"))
		})
	})
})
