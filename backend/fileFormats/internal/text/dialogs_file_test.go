package text_test

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDlg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dlg Suite")
}

var _ = Describe("DlgFile", Ordered, func() {
	var (
		formatter           interfaces.ITextFormatter
		testDlgVerifyer     textVerifier.ITextVerificationService
		destination         locations.IDestination
		testDlgCompressor   dlg.IDlgCompressor
		testDlgExtractor    dlg.IDlgExtractor
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
		temp                *common.TempProvider
		log                 *testcommon.MockLogHandler
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

		formatter = &formatters.TxtFormatter{
			TargetExtension: ".txt",
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}

		interactions.NewInteractionServiceWithConfig(config)
		interactions.NewInteractionWithTextFormatter(formatter)

		log = testcommon.NewLogHandlerMock()

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
			TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
			ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
		}
	})

	BeforeEach(func() {
		testDlgExtractor = dlg.NewDlgExtractor(log)
		Expect(testDlgExtractor).NotTo(BeNil())

		testDlgCompressor = dlg.NewDlgCompressor(log)
		Expect(testDlgCompressor).NotTo(BeNil())

		testDlgVerifyer = textVerifier.NewTextVerificationService(log)
		Expect(testDlgVerifyer).NotTo(BeNil())

		mockNotifierService = testcommon.NewMockNotifier()
		Expect(mockNotifierService).NotTo(BeNil())
	})

	AfterEach(func() {
		Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
	})

	Context("Clones functionality", func() {
		var checkClones = func(relPath string, expectedCount int) {
			testFilePath := filepath.Join(gameLocationPath, relPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())

			gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
			source.PopulateDuplicatesFiles(gameVersion)

			Expect(source.Get().ClonedItems).To(HaveLen(expectedCount))
		}

		It("should not have clones in tutorial.msb", func() {
			checkClones(`ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`, 0)
		})

		It("should have files clones in bika07_235.bin", func() {
			checkClones(`ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`, 283)
		})

		It("should have files clones in hiku2800.bin", func() {
			checkClones(`ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\hi\hiku2800\hiku2800.bin`, 1)
		})
	})

	Context("Dialogs Extract Functionality", func() {
		BeforeEach(func() {
			testDlgExtractor = dlg.NewDlgExtractor(log)
			Expect(testDlgExtractor).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		var extractFile = func(fileRelPath string, fileType models.NodeType) {
			testFilePath := filepath.Join(gameLocationPath, fileRelPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())
			Expect(source.GetType()).To(Equal(fileType))

			Expect(destination).NotTo(BeNil())
			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		}

		It("should extract the bika07_235.bin successfully", func() {
			extractFile(`ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`, models.Dialogs)
		})

		It("should extract the tutorial.msb successfully", func() {
			extractFile(`ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`, models.Tutorial)
		})

		It("should extract the cloud.bin successfully", func() {
			extractFile(`ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`, models.Dialogs)
		})

		It("should extract variant event file crcr0000.bin successfully", func() {
			extractFile(`ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\cr\crcr0000\crcr0000.bin`, models.DialogsSpecial)
		})

		It("should extract variant event file credits.bin successfully", func() {
			extractFile(`ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\cr\credits\credits.bin`, models.DialogsSpecial)
		})

		It("should throw error", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
			testFilePath := filepath.Join(gameLocationPath, file)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			m := new(testcommon.MockDlgDecoder)
			m.On("Decoder", source, destination, mock.Anything).Return(fmt.Errorf("mock error"))

			mockDlgDecoder := dlg.DialogExtractor{
				DialogDecoder: m,
				Logger:        log,
			}

			err = mockDlgDecoder.Extract(source, destination)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Dialogs Compress Functionality", func() {
		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		var checkClones = func(fileClones []string) {
			count := 0
			for _, clone := range fileClones {
				destinationPath := destination.Import().GetTargetDirectory()
				cloneFilePath := filepath.Join(destinationPath, clone)
				Expect(common.CheckPathExists(cloneFilePath)).To(Succeed(), "Clone file should exist: %s", cloneFilePath)

				count++
			}

			Expect(count).To(Equal(len(fileClones)), "All clones should be processed")
			mockNotifierService.NotifySuccess(fmt.Sprintf("Clones: %d/%d", count, len(fileClones)))
		}

		var compressClones = func(relPath string) {
			testFilePath := filepath.Join(gameLocationPath, relPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
			source.PopulateDuplicatesFiles(gameVersion)

			Expect(testDlgCompressor.Compress(source, destination)).To(Succeed())
			Expect(common.CheckPathExists(destination.Import().GetTargetFile())).To(Succeed())

			if source.Get().ClonedItems != nil {
				checkClones(source.Get().ClonedItems)
			}
		}

		It("should compress the cloud.bin successfully", func() {
			compressClones(`ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`)
		})

		It("should compress the tutorial.msb successfully", func() {
			compressClones(`ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`)
		})

		It("should compress the bika07_235.bin successfully", func() {
			compressClones(`ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`)
		})

		It("should compress the hiku2800.bin successfully", func() {
			compressClones(`ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\hi\hiku2800\hiku2800.bin`)
		})

		It("should compress the crcr0000.bin successfully", func() {
			compressClones(`ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\cr\crcr0000\crcr0000.bin`)
		})
	})

	Context("Verifyer Functionality", func() {
		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		It("should verify file integrity binary fail", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Extract()).To(Succeed())

			dlgIntegrity := textVerifier.NewTextVerificationService(log)
			Expect(dlgIntegrity).NotTo(BeNil())

			Expect(testcommon.RemoveFirstNLines(destination.Extract().GetTargetFile(), 4)).To(Succeed())

			err = dlgIntegrity.Verify(source, destination, textVerifier.NewTextExtractionVerificationStrategy())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source and target segments count mismatch"))

			err = common.CheckPathExists(destination.Extract().GetTargetFile())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("path does not exist:"))
		})
	})

	Context("DialogFile Functionality", func() {
		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		It("should extract dialogFile cloud.bin successfully", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(gameLocationPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed(), "File should exist: %s", sourceFile)

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())
			destination.InitializeLocations(source, formatter)

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Extract()).To(Succeed())

			err = testDlgVerifyer.Verify(source, destination, textVerifier.NewTextExtractionVerificationStrategy())
			Expect(err).To(BeNil())
		})

		It("should compress dialogFile cloud.bin successfully", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(gameLocationPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed(), "File should exist: %s", sourceFile)

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())
			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Extract()).To(Succeed())
			Expect(dlgFile.Compress()).To(Succeed())

			err = testDlgVerifyer.Verify(source, destination, textVerifier.NewTextCompressionVerificationStrategy())
			Expect(err).To(BeNil())
		})
	})

	Context("Dialogs Count Functionality", func() {
		It("should be dialogs count", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.GetType())
			defer encoding.Dispose()

			gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
			Expect(gameVersion).To(Equal(models.FFX2))

			count, err := lib.TextSegmentsCounter(sourceFile, source.GetType(), gameVersion)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(119))
		})
	})
})
