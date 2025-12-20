package objectsfile_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	testcommon "ffxresources/testData"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDataObjectsReaders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Objects Readers Suite")
}

var _ = Describe("Data Objects Readers", Ordered, func() {
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

	AfterEach(func() {
		// Restore original state
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot
		common.SetModsEnabled(originalModsEnabled)
		common.SetVerboseMode(originalVerboseMode)

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
	})

	Context("ReadCommandsWithAllLocalizations", func() {
		It("should populate COMMANDS global variable when file exists", func() {
			// Arrange
			Expect(datastore.Commands.Len()).To(Equal(0))

			// Act
			objectsfile.ReadCommandsWithAllLocalizations()

			// Assert
			Expect(datastore.Commands.Len()).To(BeNumerically(">", 0))
		})

		It("should leave COMMANDS as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.Commands.Len()).To(Equal(0))

			// Act
			objectsfile.ReadCommandsWithAllLocalizations()

			// Assert
			Expect(datastore.Commands.Len()).To(Equal(0))
		})
	})

	Context("ReadKeyItemsWithAllLocalizations", func() {
		It("should populate KEY_ITEMS global variable when file exists", func() {
			// Arrange
			Expect(datastore.KeyItems.Len()).To(Equal(0))

			// Act
			objectsfile.ReadKeyItemsWithAllLocalizations()

			// Assert
			Expect(datastore.KeyItems.Len()).To(BeNumerically(">", 0))
		})

		It("should leave KEY_ITEMS as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.KeyItems.Len()).To(BeNumerically(">", 0))

			// Act
			objectsfile.ReadKeyItemsWithAllLocalizations()

			// Assert
			Expect(datastore.KeyItems.Len()).To(Equal(0))
		})
	})

	Context("ReadItemsWithAllLocalizations", func() {
		It("should populate ITEMS global variable when file exists", func() {
			// Arrange
			Expect(datastore.Items.Len()).To(Equal(0))

			// Act
			objectsfile.ReadItemsWithAllLocalizations()

			// Assert
			Expect(datastore.Items.Len()).To(BeNumerically(">", 0))
		})

		It("should leave ITEMS as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.Items.Len()).To(Equal(0))

			// Act
			objectsfile.ReadItemsWithAllLocalizations()

			// Assert
			Expect(datastore.Items.Len()).To(Equal(0))
		})
	})

	Context("ReadArmsTextWithAllLocalizations", func() {
		It("should populate ARMS_TEXT global variable when file exists", func() {
			// Arrange
			Expect(datastore.ArmsTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadArmsTextWithAllLocalizations()

			// Assert
			Expect(datastore.ArmsTxt.Len()).To(BeNumerically(">", 0))
		})

		It("should leave ARMS_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.ArmsTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadArmsTextWithAllLocalizations()

			// Assert
			Expect(datastore.ArmsTxt.Len()).To(Equal(0))
		})
	})

	Context("ReadConfigTextWithAllLocalizations", func() {
		It("should populate CONFIG_TEXT global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.CONFIG_TEXT).To(BeNil())

			// Act
			objectsfile.ReadConfigTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.CONFIG_TEXT).ToNot(BeNil())
			Expect(objectsfile.CONFIG_TEXT.Len()).To(BeNumerically(">", 0))
		})

		It("should leave CONFIG_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.CONFIG_TEXT).To(BeNil())

			// Act
			objectsfile.ReadConfigTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.CONFIG_TEXT).To(BeNil())
		})
	})

	Context("ReadItemCommandsWithAllLocalizations", func() {
		It("should populate ITEM_TEXT global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.ITEM_TEXT).To(BeNil())

			// Act
			objectsfile.ReadItemCommandsWithAllLocalizations()

			// Assert
			Expect(objectsfile.ITEM_TEXT).ToNot(BeNil())
			Expect(objectsfile.ITEM_TEXT.Len()).To(BeNumerically(">", 0))
		})

		It("should leave ITEM_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.ITEM_TEXT).To(BeNil())

			// Act
			objectsfile.ReadItemCommandsWithAllLocalizations()

			// Assert
			Expect(objectsfile.ITEM_TEXT).To(BeNil())
		})
	})

	Context("ReadMainMenuTextWithAllLocalizations", func() {
		It("should populate MMAIN_TEXT global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.MMAIN_TEXT).To(BeNil())

			// Act
			objectsfile.ReadMainMenuTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.MMAIN_TEXT).ToNot(BeNil())
			Expect(objectsfile.MMAIN_TEXT.Len()).To(BeNumerically(">", 0))
		})

		It("should leave MMAIN_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.MMAIN_TEXT).To(BeNil())

			// Act
			objectsfile.ReadMainMenuTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.MMAIN_TEXT).To(BeNil())
		})
	})

	Context("ReadPlayerRomTextWithAllLocalizations", func() {
		It("should populate PLAYER_ROOM global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.PLAYER_ROOM).To(BeNil())

			// Act
			objectsfile.ReadPlayerRomTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.PLAYER_ROOM).ToNot(BeNil())
			Expect(objectsfile.PLAYER_ROOM.Len()).To(BeNumerically(">", 0))
		})

		It("should leave PLAYER_ROOM as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.PLAYER_ROOM).To(BeNil())

			// Act
			objectsfile.ReadPlayerRomTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.PLAYER_ROOM).To(BeNil())
		})
	})

	Context("ReadBattleTextWithAllLocalizations", func() {
		It("should populate BTL_TEXT global variable when file exists", func() {
			// Arrange
			Expect(datastore.BattleTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadBattleTextWithAllLocalizations()

			// Assert
			Expect(datastore.BattleTxt.Len()).To(BeNumerically(">", 0))
		})

		It("should leave BTL_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.BattleTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadBattleTextWithAllLocalizations()

			// Assert
			Expect(datastore.BattleTxt.Len()).To(Equal(0))
		})
	})

	Context("ReadBattleEndTextWithAllLocalizations", func() {
		It("should populate BTLEND_TEXT global variable when file exists", func() {
			// Arrange
			Expect(datastore.BattleEndTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadBattleEndTextWithAllLocalizations()

			// Assert
			Expect(datastore.BattleEndTxt.Len()).To(BeNumerically(">", 0))
		})

		It("should leave BTLEND_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(datastore.BattleEndTxt.Len()).To(Equal(0))

			// Act
			objectsfile.ReadBattleEndTextWithAllLocalizations()

			// Assert
			Expect(datastore.BattleEndTxt.Len()).To(Equal(0))
		})
	})

	Context("ReadMonsterMagic1WithAllLocalizations", func() {
		It("should populate MONMAGIC1 global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.MONMAGIC1).To(BeNil())

			// Act
			objectsfile.ReadMonsterMagic1WithAllLocalizations()

			// Assert
			Expect(objectsfile.MONMAGIC1).ToNot(BeNil())
			Expect(objectsfile.MONMAGIC1.Len()).To(BeNumerically(">", 0))
		})

		It("should leave MONMAGIC1 as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.MONMAGIC1).To(BeNil())

			// Act
			objectsfile.ReadMonsterMagic1WithAllLocalizations()

			// Assert
			Expect(objectsfile.MONMAGIC1).To(BeNil())
		})
	})

	Context("ReadMonsterMagic2WithAllLocalizations", func() {
		It("should populate MONMAGIC2 global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.MONMAGIC2).To(BeNil())

			// Act
			objectsfile.ReadMonsterMagic2WithAllLocalizations()

			// Assert
			Expect(objectsfile.MONMAGIC2).ToNot(BeNil())
			Expect(objectsfile.MONMAGIC2.Len()).To(BeNumerically(">", 0))
		})

		It("should leave MONMAGIC2 as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.MONMAGIC2).To(BeNil())

			// Act
			objectsfile.ReadMonsterMagic2WithAllLocalizations()

			// Assert
			Expect(objectsfile.MONMAGIC2).To(BeNil())
		})
	})

	Context("ReadBuildTextWithAllLocalizations", func() {
		It("should populate BUILD_TEXT global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.BUILD_TEXT).To(BeNil())

			// Act
			objectsfile.ReadBuildTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.BUILD_TEXT).ToNot(BeNil())
			Expect(objectsfile.BUILD_TEXT.Len()).To(BeNumerically(">", 0))
		})

		It("should leave BUILD_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.BUILD_TEXT).To(BeNil())

			// Act
			objectsfile.ReadBuildTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.BUILD_TEXT).To(BeNil())
		})
	})

	Context("ReadNameTextWithAllLocalizations", func() {
		It("should populate NAME_TEXT global variable when file exists", func() {
			// Arrange
			Expect(objectsfile.NAME_TEXT).To(BeNil())

			// Act
			objectsfile.ReadNameTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.NAME_TEXT).ToNot(BeNil())
			Expect(objectsfile.NAME_TEXT.Len()).To(BeNumerically(">", 0))
		})

		It("should leave NAME_TEXT as nil when file does not exist", func() {
			// Arrange
			tempDir, err := os.MkdirTemp("", "test_converter")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tempDir)

			// Set to non-existent path
			common.GameFilesRoot = tempDir
			Expect(objectsfile.NAME_TEXT).To(BeNil())

			// Act
			objectsfile.ReadNameTextWithAllLocalizations()

			// Assert
			Expect(objectsfile.NAME_TEXT).To(BeNil())
		})
	})

	Context("Integration Tests", func() {
		It("should load multiple data types successfully", func() {
			// Act - Load multiple data types
			objectsfile.ReadCommandsWithAllLocalizations()
			objectsfile.ReadKeyItemsWithAllLocalizations()
			objectsfile.ReadItemsWithAllLocalizations()
			objectsfile.ReadBattleTextWithAllLocalizations()
			objectsfile.ReadBattleEndTextWithAllLocalizations()

			// Verify they have different data
			Expect(datastore.Commands.Len()).To(BeNumerically(">", 0))
			Expect(datastore.KeyItems.Len()).To(BeNumerically(">", 0))
			Expect(datastore.Items.Len()).To(BeNumerically(">", 0))
			Expect(datastore.BattleTxt.Len()).To(BeNumerically(">", 0))
			Expect(datastore.BattleEndTxt.Len()).To(BeNumerically(">", 0))
		})
	})
})
