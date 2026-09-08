package objectsfile_test

import (
	"encoding/json"
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
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
		ffxencoding.ByteToCharMaps = make(map[string]map[uint]rune)
		ffxencoding.CharToByteMaps = make(map[string]map[rune]uint)
	})

	Context("WriteAllKeyItemsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadKeyItemsWithAllLocalizations()
		})

		It("should write key items data to binary and verify persistence", func() {
			Expect(datastore.KeyItems.Len()).To(Equal(64))

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			jsonPath := filepath.Join(tempEditDir, common.WithVersionSuffix("key_items_all_localizations.json"))
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(objectsfile.ProcessKeyItemsJsonFile()).To(Succeed())

			// Verify modification in memory
			modifiedItem := datastore.KeyItems.Get(0)
			Expect(modifiedItem).ToNot(BeNil())
			usContent := modifiedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			usContentStr := usContent.GetString()
			Expect(usContentStr).To(Equal("Test Write Key Item"))

			// Write to binary files
			Expect(objectsfile.WriteAllKeyItemsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.KeyItems.Clear()
			objectsfile.ReadKeyItemsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItem := datastore.KeyItems.Get(0)
			Expect(reloadedItem).ToNot(BeNil())
			reloadedUsContent := reloadedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Key Item"))
		})

		It("should return error when no key items are loaded", func() {
			datastore.KeyItems.Clear()
			err := objectsfile.WriteAllKeyItemsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no key items loaded"))
		})
	})

	Context("WriteAllCommandsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadCommandLocalizations()
		})

		It("should write commands data to binary and verify persistence", func() {
			Expect(datastore.Commands.Len()).To(Equal(320))

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			jsonPath := filepath.Join(tempEditDir, common.WithVersionSuffix("commands_all_localizations.json"))
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(objectsfile.ProcessCommandsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllCommandsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.Commands.Clear()
			objectsfile.ReadCommandLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedCommand := datastore.Commands.Get(0)
			Expect(reloadedCommand).ToNot(BeNil())
			reloadedUsContent := reloadedCommand.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Command"))
		})

		It("should return error when no commands are loaded", func() {
			datastore.Commands.Clear()
			err := objectsfile.WriteAllCommandsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no commands loaded"))
		})
	})

	Context("WriteAllItemsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadItemsWithAllLocalizations()
		})

		It("should write items data to binary and verify persistence", func() {
			Expect(datastore.Items.Len()).To(Equal(0))

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			jsonPath := filepath.Join(tempEditDir, common.WithVersionSuffix("items_all_localizations.json"))
			jsonData, err := json.MarshalIndent(testData, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.WriteFile(jsonPath, jsonData, 0644)).To(Succeed())
			Expect(objectsfile.ProcessItemsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllItemsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.Items.Clear()
			objectsfile.ReadItemsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItem := datastore.Items.Get(0)
			Expect(reloadedItem).ToNot(BeNil())
			reloadedUsContent := reloadedItem.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Item"))
		})

		It("should return error when no items are loaded", func() {
			datastore.Items.Clear()
			err := objectsfile.WriteAllItemsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no items loaded"))
		})
	})

	Context("WriteAllArmsTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadArmsTextWithAllLocalizations()
		})

		It("should write arms text data to binary and verify persistence", func() {
			Expect(datastore.ArmsTxt.Len()).To(Equal(0))

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessArmsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllArmsTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.ArmsTxt.Clear()
			objectsfile.ReadArmsTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedArms := datastore.ArmsTxt.Get(0)
			Expect(reloadedArms).ToNot(BeNil())
			reloadedUsContent := reloadedArms.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Arms"))
		})

		It("should return error when no arms text is loaded", func() {
			datastore.ArmsTxt.Clear()
			err := objectsfile.WriteAllArmsTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no arms text loaded"))
		})
	})

	Context("WriteAllBattleTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBattleTextWithAllLocalizations()
		})

		It("should write battle text data to binary and verify persistence", func() {
			Expect(datastore.BattleTxt.Len()).To(Equal(0))

			// Create test JSON with modified content (name-only for battle text)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessBattleTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllBattleTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.BattleTxt.Clear()
			objectsfile.ReadBattleTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBattle := datastore.BattleTxt.Get(0)
			Expect(reloadedBattle).ToNot(BeNil())
			reloadedUsContent := reloadedBattle.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Battle"))
		})

		It("should return error when no battle text is loaded", func() {
			datastore.BattleTxt.Clear()
			err := objectsfile.WriteAllBattleTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no battle text loaded"))
		})
	})

	Context("WriteAllBattleEndTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBattleEndTextWithAllLocalizations()
		})

		It("should write battle end text data to binary and verify persistence", func() {
			Expect(datastore.BattleEndTxt.Len()).To(Equal(0))

			// Create test JSON with modified content (name-only due to bug mentioned in data_writers.go)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessBattleEndTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllBattleEndTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			datastore.BattleEndTxt.Clear()
			objectsfile.ReadBattleEndTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBattleEnd := datastore.BattleEndTxt.Get(0)
			Expect(reloadedBattleEnd).ToNot(BeNil())
			reloadedUsContent := reloadedBattleEnd.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Battle End"))
		})

		It("should return error when no battle end text is loaded", func() {
			datastore.BattleEndTxt.Clear()
			err := objectsfile.WriteAllBattleEndTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no battle end text loaded"))
		})
	})

	Context("WriteAllMonsterMagic1Data", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMonsterMagic1WithAllLocalizations()
		})

		It("should write monster magic 1 data to binary and verify persistence", func() {
			Expect(objectsfile.MONMAGIC1.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for monster magic)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessMonsterMagic1JsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllMonsterMagic1Data()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.MONMAGIC1 = nil
			objectsfile.ReadMonsterMagic1WithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMonMagic := objectsfile.MONMAGIC1.Get(0)
			Expect(reloadedMonMagic).ToNot(BeNil())
			reloadedUsContent := reloadedMonMagic.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write MonMagic1"))
		})

		It("should return error when no monster magic 1 is loaded", func() {
			objectsfile.MONMAGIC1 = nil
			err := objectsfile.WriteAllMonsterMagic1Data()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no monster magic 1 loaded"))
		})
	})

	Context("WriteAllMonsterMagic2Data", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMonsterMagic2WithAllLocalizations()
		})

		It("should write monster magic 2 data to binary and verify persistence", func() {
			Expect(objectsfile.MONMAGIC2.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for monster magic)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessMonsterMagic2JsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllMonsterMagic2Data()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.MONMAGIC2 = nil
			objectsfile.ReadMonsterMagic2WithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMonMagic := objectsfile.MONMAGIC2.Get(0)
			Expect(reloadedMonMagic).ToNot(BeNil())
			reloadedUsContent := reloadedMonMagic.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write MonMagic2"))
		})

		It("should return error when no monster magic 2 is loaded", func() {
			objectsfile.MONMAGIC2 = nil
			err := objectsfile.WriteAllMonsterMagic2Data()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no monster magic 2 loaded"))
		})
	})

	Context("WriteAllBuildTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadBuildTextWithAllLocalizations()
		})

		It("should write build text data to binary and verify persistence", func() {
			Expect(objectsfile.BUILD_TEXT.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only due to bug mentioned in data_writers.go)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessBuildTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllBuildTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.BUILD_TEXT = nil
			objectsfile.ReadBuildTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedBuild := objectsfile.BUILD_TEXT.Get(0)
			Expect(reloadedBuild).ToNot(BeNil())
			reloadedUsContent := reloadedBuild.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Build"))
		})

		It("should return error when no build text is loaded", func() {
			objectsfile.BUILD_TEXT = nil
			err := objectsfile.WriteAllBuildTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no build text loaded"))
		})
	})

	Context("WriteAllConfigTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadConfigTextWithAllLocalizations()
		})

		It("should write config text data to binary and verify persistence", func() {
			Expect(objectsfile.CONFIG_TEXT.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessConfigTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllConfigTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.CONFIG_TEXT = nil
			objectsfile.ReadConfigTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedConfig := objectsfile.CONFIG_TEXT.Get(0)
			Expect(reloadedConfig).ToNot(BeNil())
			reloadedUsContent := reloadedConfig.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Config"))
		})

		It("should return error when no config text is loaded", func() {
			objectsfile.CONFIG_TEXT = nil
			err := objectsfile.WriteAllConfigTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no config text loaded"))
		})
	})

	Context("WriteAllMainMenuTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadMainMenuTextWithAllLocalizations()
		})

		It("should write main menu text data to binary and verify persistence", func() {
			Expect(objectsfile.MMAIN_TEXT.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessMainMenuTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllMainMenuTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.MMAIN_TEXT = nil
			objectsfile.ReadMainMenuTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedMainMenu := objectsfile.MMAIN_TEXT.Get(0)
			Expect(reloadedMainMenu).ToNot(BeNil())
			reloadedUsContent := reloadedMainMenu.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Main Menu"))
		})

		It("should return error when no main menu text is loaded", func() {
			objectsfile.MMAIN_TEXT = nil
			err := objectsfile.WriteAllMainMenuTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no main menu text loaded"))
		})
	})

	Context("WriteAllItemCommandsData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadItemCommandsWithAllLocalizations()
		})

		It("should write item commands data to binary and verify persistence", func() {
			Expect(objectsfile.ITEM_TEXT.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessItemCommandsJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllItemCommandsData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.ITEM_TEXT = nil
			objectsfile.ReadItemCommandsWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedItemText := objectsfile.ITEM_TEXT.Get(0)
			Expect(reloadedItemText).ToNot(BeNil())
			reloadedUsContent := reloadedItemText.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Item Commands"))
		})

		It("should return error when no item text is loaded", func() {
			objectsfile.ITEM_TEXT = nil
			err := objectsfile.WriteAllItemCommandsData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no item text loaded"))
		})
	})

	Context("WriteAllPlayerRoomTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadPlayerRomTextWithAllLocalizations()
		})

		It("should write player room text data to binary and verify persistence", func() {
			Expect(objectsfile.PLAYER_ROOM.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content
			testData := []objectsfile.NameDescriptionData{
				{
					NameOnlyData: objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessPlayerRoomTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllPlayerRoomTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.PLAYER_ROOM = nil
			objectsfile.ReadPlayerRomTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedPlayerRoom := objectsfile.PLAYER_ROOM.Get(0)
			Expect(reloadedPlayerRoom).ToNot(BeNil())
			reloadedUsContent := reloadedPlayerRoom.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Player Room"))
		})

		It("should return error when no player room text is loaded", func() {
			objectsfile.PLAYER_ROOM = nil
			err := objectsfile.WriteAllPlayerRoomTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no player room text loaded"))
		})
	})

	Context("WriteAllNameTextData", func() {
		BeforeEach(func() {
			Expect(reader.InitializeInternals()).To(Succeed())
			objectsfile.ReadNameTextWithAllLocalizations()
		})

		It("should write name text data to binary and verify persistence", func() {
			Expect(objectsfile.NAME_TEXT.Items()).ToNot(BeEmpty())

			// Create test JSON with modified content (name-only for name text)
			testData := []objectsfile.NameOnlyData{
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
			Expect(objectsfile.ProcessNameTextJsonFile()).To(Succeed())

			// Write to binary files
			Expect(objectsfile.WriteAllNameTextData()).To(Succeed())

			// Enable mods to read from mods folder
			common.SetModsEnabled(true)

			// Clear memory and reload from binary
			objectsfile.NAME_TEXT = nil
			objectsfile.ReadNameTextWithAllLocalizations()

			// Disable mods after reading
			common.SetModsEnabled(false)

			// Verify the changes were persisted to file
			reloadedNameText := objectsfile.NAME_TEXT.Get(0)
			Expect(reloadedNameText).ToNot(BeNil())
			reloadedUsContent := reloadedNameText.GetTextObject().GetKeyedString("name").GetLocalizedContent("us")
			Expect(reloadedUsContent.GetString()).To(Equal("Test Write Name Text"))
		})

		It("should return error when no name text is loaded", func() {
			objectsfile.NAME_TEXT = nil
			err := objectsfile.WriteAllNameTextData()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no name text loaded"))
		})
	})
})
