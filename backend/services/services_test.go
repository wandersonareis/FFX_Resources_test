package services_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"ffxresources/backend/services"
	"ffxresources/testData"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extract and Compress Services Suite")
}

var _ = Describe("FFX Services", Ordered, func() {
	var (
		formatter           interfaces.ITextFormatter
		rootDir             string
		binaryPath          string
		gameVersionDir      string
		extractTempPath     string
		reimportTempPath    string
		translatePath       string
		testDataPath        string
		gameLocationPath    string
		config              *interactions.FFXAppConfig
		mockNotifierService *testcommon.MockNotifier
		mockProgressService *testcommon.MockProgressService
		temp                *common.TempProvider
		collectionService   *services.CollectionService
		compressService     *services.CompressService
		extractService      *services.ExtractService
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		Expect(os.Setenv("FFX_GAME_VERSION", "2")).To(Succeed())

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

		binaryPath = "binary"
		temp = common.NewTempProvider("", "")
		gameVersionDir = "FFX-2"

		testDataPath = filepath.Join(rootDir, gameVersionDir)
		gameLocationPath = filepath.Join(testDataPath, binaryPath)

		extractTempPath = filepath.Join(temp.TempFilePath, "extract")
		reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
		translatePath = filepath.Join(testDataPath, "translated")

		config = &interactions.FFXAppConfig{
			FFXGameVersion:    2,
			GameFilesLocation: gameLocationPath,
			ExtractLocation:   extractTempPath,
			TranslateLocation: translatePath,
			ImportLocation:    reimportTempPath,
		}
		Expect(config.UpdateConfigFile(filepath.Join(rootDir, "config.json"))).To(Succeed())

		interactions.NewInteractionServiceWithConfig(config)

		formatter = &formatters.TxtFormatter{
			TargetExtension: ".txt",
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}

		interactions.NewInteractionWithTextFormatter(formatter)
	})

	AfterAll(func() {
		mockNotifierService.NotifySuccess("Test extract and compress completed successfully")
	})

	BeforeEach(func() {
		mockNotifierService = testcommon.NewMockNotifier()
		Expect(mockNotifierService).NotTo(BeNil(), "Mock notifier service should not be nil")

		mockProgressService = testcommon.NewMockProgressService()
		Expect(mockProgressService).NotTo(BeNil(), "Mock progress service should not be nil")
	})

	AfterEach(func() {
		Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory(gameLocationPath)).To(Succeed())
	})

	Context("CollectionService", func() {
		BeforeEach(func() {
			Expect(mockNotifierService).NotTo(BeNil(), "Mock notifier service should not be nil")

			collectionService = services.NewCollectionService(mockNotifierService)
			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")
		})

		AfterEach(func() {
			collectionService = nil
			services.NodeDataStore = nil
			Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory(gameLocationPath)).To(Succeed())
		})

		It("should be equal rawMap and nodeStore", func() {
			/* rootDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			Expect(rootDir).NotTo(BeEmpty(), "Root directory should not be empty") */

			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil(), "Raw map should not be nil")
			
			nodeStore := services.NodeDataStore
			Expect(nodeStore).NotTo(BeNil(), "Node store should not be nil")
			Expect(nodeStore.Len()).To(BeNumerically(">", 0), "Node store should have elements")
			
			Expect(func() bool {
				if nodeStore.Len() != len(rawMap) {
					return false
				}

				for key, value := range rawMap {
					node, ok := nodeStore.Get(key)
					if !ok {
						return false
					}

					if node == nil && value == nil {
						return false
					}

					if node == nil || value == nil {
						return false
					}

					if !reflect.DeepEqual(node, value) {
						return false
					}
				}
				return true
			}()).To(BeTrue(), "All nodes in rawMap should exist in nodeStore and be equal")
			
		})

		It("should return nil if path is empty", func() {
			Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory("")).To(Succeed())

			gamePath := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			currentPath, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred(), "Error getting current working directory")
			Expect(gamePath).To(Equal(currentPath), "Target directory should be current working directory")

			tree := collectionService.BuildTree("")
			Expect(tree).To(BeNil(), "Tree should be nil")
		})

		It("should create file tree successfully", func() {
			tree := collectionService.BuildTree(gameLocationPath)
			Expect(tree).To(HaveLen(1), "Tree should have one element")
			Expect(tree[0].Label).To(Equal("Final Fantasy X-2"), "Tree node label should be Final Fantasy X-2")
			Expect(tree[0].Icon).To(Equal("pi pi-folder"), "Tree node icon should be pi pi-folder")
			Expect(tree[0].Children).To(HaveLen(2), "Tree node should have two child")
			Expect(tree[0].Children[0].Label).To(Equal("ffx-2_data"), "Tree node child label should be ffx-2_data")
			Expect(tree[0].Children[1].Label).To(Equal("ffx_ps2"), "Tree node child label should be ffx_ps2")

			Expect(tree[0].Data).ToNot(BeNil(), "Tree node data should not be nil")
			Expect(tree[0].Data.Source).ToNot(BeNil(), "Tree node data source should not be nil")
			Expect(tree[0].Data.Source.Type).To(Equal(models.Folder), "Tree node data source type should be Folder")
			Expect(tree[0].Data.Extract).ToNot(BeNil(), "Tree node data extract should not be nil")
			Expect(tree[0].Data.Translate).ToNot(BeNil(), "Tree node data translate should not be nil")

			Expect(mockNotifierService.Notifications).To(HaveLen(0), "Notifications should be empty")
			Expect(mockProgressService.Started).To(BeFalse(), "Progress service should not be started")
		})
	})

	Context("ExtractService", func() {
		BeforeEach(func() {
			Expect(mockNotifierService).NotTo(BeNil(), "Mock notifier service should not be nil")
			Expect(mockProgressService).NotTo(BeNil(), "Mock progress service should not be nil")

			extractService = services.NewExtractService(mockNotifierService, mockProgressService)
			Expect(extractService).NotTo(BeNil(), "Extract service should not be nil")

			collectionService = services.NewCollectionService(mockNotifierService)
			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")
		})

		AfterEach(func() {
			extractService = nil
			collectionService = nil
			services.NodeDataStore = nil
			Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory(gameLocationPath)).To(Succeed())
		})

		It("Path never be empty or cause panic", func() {
			err := extractService.Extract("")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("primitive type: argument is zero value"))
		})

		It("should be error extract if nodeStore not started", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(gameLocationPath, file)

			err := extractService.Extract(testFilePath)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("nodeStore: argument is nil"))
		})

		It("should be error on extract", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(testDataPath, file)

			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")

			/* rootDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			Expect(rootDir).NotTo(BeEmpty(), "Root directory should not be empty") */

			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			Expect(extractService).NotTo(BeNil(), "Extract service should not be nil")

			err := extractService.Extract(testFilePath)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("node not found for path: " + testFilePath))

			Expect(mockNotifierService.Notifications).To(HaveLen(0))

			Expect(mockProgressService.Started).To(BeFalse())
			Expect(mockProgressService.Max).To(Equal(0))
			Expect(mockProgressService.Steps).To(Equal(0))
			Expect(mockProgressService.Files).To(HaveLen(0))
		})

		It("should be error node not found", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(testDataPath, file)

			/* gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

			rootDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory() */

			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			extractService := services.NewExtractService(mockNotifierService, mockProgressService)
			Expect(extractService).NotTo(BeNil(), "Extract service should not be nil")

			err := extractService.Extract(testFilePath)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("node not found for path: " + testFilePath))

			Expect(mockNotifierService.Notifications).To(HaveLen(0))

			Expect(mockProgressService.Started).To(BeFalse())
			Expect(mockProgressService.Max).To(Equal(0))
			Expect(mockProgressService.Steps).To(Equal(0))
			Expect(mockProgressService.Files).To(HaveLen(0))
		})

		It("should be extract file successful", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(gameLocationPath, file)

			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			extractService := services.NewExtractService(mockNotifierService, mockProgressService)
			Expect(extractService).NotTo(BeNil(), "Extract service should not be nil")

			Expect(extractService.Extract(testFilePath)).To(Succeed())

			Expect(mockNotifierService.Notifications).To(HaveLen(1))
			Expect(mockNotifierService.Notifications[0].Severity).To(Equal(services.SeveritySuccess.String()))
			Expect(mockNotifierService.Notifications[0].Message).To(Equal("File tutorial.msb extracted successfully!"))

			Expect(mockProgressService.Started).To(BeFalse())
			Expect(mockProgressService.Max).To(Equal(0))
			Expect(mockProgressService.Steps).To(Equal(0))
			Expect(mockProgressService.Files).To(HaveLen(0))
		})

		It("should be extract directory successful", func() {
			directory := `ffx_ps2\ffx2\master\new_uspc\cloudsave`
			directoryPath := filepath.Join(gameLocationPath, directory)

			//Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory(testDataPath)).To(Succeed())

			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			Expect(extractService).NotTo(BeNil(), "Extract service should not be nil")

			Expect(extractService.Extract(directoryPath)).To(Succeed())

			Expect(mockNotifierService.Notifications).To(HaveLen(1))
			Expect(mockNotifierService.Notifications[0].Severity).To(Equal(services.SeveritySuccess.String()))
			Expect(mockNotifierService.Notifications[0].Message).To(Equal("Directory cloudsave extracted successfully!"))

			Expect(mockProgressService.Started).To(BeTrue())
			Expect(mockProgressService.Max).To(Equal(1))
			Expect(mockProgressService.Steps).To(Equal(1))
			Expect(mockProgressService.Files).To(HaveLen(1))
			Expect(mockProgressService.Files[0]).To(Equal("cloud.bin"))
		})
	})

	Context("CompressService", func() {
		BeforeEach(func() {
			Expect(mockNotifierService).NotTo(BeNil(), "Mock notifier service should not be nil")
			Expect(mockProgressService).NotTo(BeNil(), "Mock progress service should not be nil")

			collectionService = services.NewCollectionService(mockNotifierService)
			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")

			compressService = services.NewCompressService(mockNotifierService, mockProgressService)
			Expect(compressService).NotTo(BeNil(), "Compress service should not be nil")
		})
		AfterEach(func() {
			collectionService = nil
			compressService = nil
			services.NodeDataStore = nil

			Expect(interactions.NewInteractionService().GameLocation.SetTargetDirectory(gameLocationPath)).To(Succeed())
		})

		It("should be translated file not found", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(gameLocationPath, file)

			oldTranslatedPath := interactions.NewInteractionService().TranslateLocation.GetTargetDirectory()
			Expect(oldTranslatedPath).NotTo(BeEmpty(), "Old translated path should not be empty")
			Expect(interactions.NewInteractionService().TranslateLocation.SetTargetDirectory("test")).To(Succeed())
			defer func() {
				Expect(interactions.NewInteractionService().TranslateLocation.SetTargetDirectory(oldTranslatedPath)).To(Succeed())
			}()

			/* gamePath := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			Expect(gamePath).NotTo(BeEmpty(), "Game path should not be empty") */

			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")
			tree := collectionService.BuildTree(gameLocationPath)
			Expect(tree).ToNot(BeNil(), "Tree should not be nil")

			Expect(compressService).NotTo(BeNil(), "Compress service should not be nil")
			Expect(compressService.Compress(testFilePath)).To(Succeed())

			Expect(mockNotifierService.Notifications).To(HaveLen(1))
			Expect(mockNotifierService.Notifications[0].Severity).To(Equal(services.SeverityError.String()))
			Expect(mockNotifierService.Notifications[0].Message).To(HavePrefix("failed to check target file path: path does not exist"))

			Expect(mockProgressService.Started).To(BeFalse())
			Expect(mockProgressService.Max).To(Equal(0))
			Expect(mockProgressService.Steps).To(Equal(0))
			Expect(mockProgressService.Files).To(HaveLen(0))
		})

		It("should be compress file successful", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			testFilePath := filepath.Join(gameLocationPath, file)

			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")
			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			Expect(compressService).NotTo(BeNil(), "Extract service should not be nil")
			Expect(compressService.Compress(testFilePath)).To(Succeed())

			Expect(mockNotifierService.Notifications).To(HaveLen(1))
			Expect(mockNotifierService.Notifications[0].Severity).To(Equal(services.SeveritySuccess.String()))
			Expect(mockNotifierService.Notifications[0].Message).To(Equal("File tutorial.msb compressed successfully!"))

			Expect(mockProgressService.Started).To(BeFalse())
			Expect(mockProgressService.Max).To(Equal(0))
			Expect(mockProgressService.Steps).To(Equal(0))
			Expect(mockProgressService.Files).To(HaveLen(0))
		})

		It("should be compress directory successful", func() {
			directory := `ffx_ps2\ffx2\master\new_uspc\cloudsave`
			directoryPath := filepath.Join(gameLocationPath, directory)

			Expect(collectionService).NotTo(BeNil(), "Collection service should not be nil")
			rawMap := collectionService.CreateNodeDataStore(gameLocationPath, formatter)
			Expect(rawMap).NotTo(BeNil())
			Expect(len(rawMap)).To(BeNumerically(">", 0))

			Expect(compressService).NotTo(BeNil(), "Compress service should not be nil")
			Expect(compressService.Compress(directoryPath)).To(Succeed())

			Expect(mockNotifierService.Notifications).To(HaveLen(1))
			Expect(mockNotifierService.Notifications[0].Severity).To(Equal(services.SeveritySuccess.String()))
			Expect(mockNotifierService.Notifications[0].Message).To(Equal("Directory cloudsave compressed successfully!"))

			Expect(mockProgressService.Started).To(BeTrue())
			Expect(mockProgressService.Max).To(Equal(1))
			Expect(mockProgressService.Steps).To(Equal(1))
			Expect(mockProgressService.Files).To(HaveLen(1))
			Expect(mockProgressService.Files[0]).To(Equal("cloud.bin"))
		})
	})
})
