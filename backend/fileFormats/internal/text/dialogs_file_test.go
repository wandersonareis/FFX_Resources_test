package text_test

import (
	"bufio"
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"ffxresources/testData"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
		testDlgVerifyer     textVerifier.ITextVerifier
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
		log                 logger.ILoggerHandler
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

		log = logger.NewLoggerHandler("dlg_file_test")

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocationInfo(locationsBase.WithDirectoryName("extracted"), locationsBase.WithTargetDirectory(extractTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
			TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
			ImportLocation:    locations.NewImportLocationInfo(locationsBase.WithDirectoryName("reimported"), locationsBase.WithTargetDirectory(reimportTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
		}
	})

	BeforeEach(func() {
		testDlgExtractor = dlg.NewDlgExtractor(log)
		Expect(testDlgExtractor).NotTo(BeNil())

		testDlgCompressor = dlg.NewDlgCompressor(log)
		Expect(testDlgCompressor).NotTo(BeNil())

		testDlgVerifyer = textVerifier.NewTextsVerify(log)
		Expect(testDlgVerifyer).NotTo(BeNil())

		mockNotifierService = testcommon.NewMockNotifier()
		Expect(mockNotifierService).NotTo(BeNil())
	})

	AfterEach(func() {
		Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
	})

	Describe("Miscellaneous Functionality", func() {
		It("should have valid game version for FFX-2", func() {
			Expect(os.Getenv("FFX_GAME_VERSION")).To(Equal("2"))
		})

		It("should have valid configuration", func() {
			Expect(config).NotTo(BeNil())
		})

		It("should have valid formatter", func() {
			Expect(formatter).NotTo(BeNil())
		})

		It("should have interaction service with text formatter", func() {
			Expect(interactions.NewInteractionService().TextFormatter()).NotTo(BeNil())
		})

		It("should have valid destination", func() {
			Expect(destination).NotTo(BeNil())
		})

		It("should have valid logger", func() {
			Expect(log).NotTo(BeNil())
		})
	})

	Describe("Clones functionality", func() {
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
		var extractFile = func(fileRelPath string, fileType models.NodeType) {
			testFilePath := filepath.Join(gameLocationPath, fileRelPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())
			Expect(source.Get().Type).To(Equal(fileType))

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())

			// Verify the extracted file
			Expect(testDlgVerifyer).NotTo(BeNil())
			Expect(testDlgVerifyer.Verify(source, destination, textVerifier.ExtractIntegrityCheck)).To(Succeed())
			Expect(common.CheckPathExists(destination.Extract().Get().GetTargetFile())).To(Succeed())
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
	})

	Context("Dialogs Compress Functionality", func() {
		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		var checkClones = func(fileClones []string) {
			count := 0
			for _, clone := range fileClones {
				destinationPath := destination.Import().Get().GetTargetDirectory()
				cloneFilePath := filepath.Join(destinationPath, clone)
				Expect(common.CheckPathExists(cloneFilePath)).To(Succeed(), "Clone file should exist: %s", cloneFilePath)

				count++
				mockNotifierService.NotifySuccess(fmt.Sprintf("Clones: %d/%d", count, len(fileClones)))
			}

			Expect(count).To(Equal(len(fileClones)), "All clones should be processed")
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
			Expect(common.CheckPathExists(destination.Import().Get().GetTargetFile())).To(Succeed())

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

	Describe("Verifyer Functionality", func() {
		removeFirstNLines := func(filePath string, n int) error {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			var remainingLines []string
			scanner := bufio.NewScanner(file)

			lineNumber := 0
			for scanner.Scan() {
				if lineNumber >= n {
					remainingLines = append(remainingLines, scanner.Text())
				}
				lineNumber++
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			file, err = os.Create(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			writer := bufio.NewWriter(file)
			for _, line := range remainingLines {
				_, err := writer.WriteString(line + "\n")
				if err != nil {
					return err
				}
			}

			return writer.Flush()
		}

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

			dlgIntegrity := textVerifier.NewTextsVerify(log)
			Expect(dlgIntegrity).NotTo(BeNil())

			Expect(removeFirstNLines(destination.Extract().Get().GetTargetFile(), 4)).To(Succeed())

			err = dlgIntegrity.Verify(source, destination, textVerifier.ExtractIntegrityCheck)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source and target segments count mismatch"))

			err = common.CheckPathExists(destination.Extract().Get().GetTargetFile())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("path does not exist:"))
		})
	})

	It("should be dialogs count", func() {
		testPath := `binary\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
		sourceFile := filepath.Join(testDataPath, testPath)
		Expect(common.CheckPathExists(sourceFile)).To(Succeed())

		source, err := locations.NewSource(sourceFile)
		Expect(err).To(BeNil())
		Expect(source).NotTo(BeNil())

		encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
		defer encoding.Dispose()

		count, err := lib.TextSegmentsCounter(sourceFile, source.Get().Type)
		Expect(err).To(BeNil())
		Expect(count).To(Equal(119))
	})
})
