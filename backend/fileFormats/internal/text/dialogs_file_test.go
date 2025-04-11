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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	interactionService *interactions.InteractionService
	extractTempPath    string
	reimportTempPath   string
	translatePath      string
	testDataPath       string
	config             *interactions.FFXAppConfig
	formatter          interfaces.ITextFormatter
	source             interfaces.ISource
	destination        locations.IDestination
	testDlgExtractor   dlg.IDlgExtractor
	//testDlgCompressor  dlg.IDlgCompressor
	temp           *common.TempProvider
	gameVersionDir string
	log            logger.ILoggerHandler
	err            error
)

var _ = Describe("DlgFile", Ordered, func() {

	BeforeAll(func() {
		Expect(os.Setenv("APP_BASE_PATH", `F:\ffxWails\FFX_Resources\build\bin`)).To(Succeed())
		Expect(os.Setenv("FFX_GAME_VERSION", "2")).To(Succeed())

		currentDir, err := filepath.Abs(".")
		Expect(err).To(BeNil())

		temp = common.NewTempProvider("", "")

		gameVersionDir = "FFX-2"

		testDataPath = filepath.Join(currentDir, "testdata", gameVersionDir)

		extractTempPath = filepath.Join(temp.TempFilePath, "extract")
		reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
		translatePath = filepath.Join(testDataPath, "translated")

		// Setup config
		config = &interactions.FFXAppConfig{
			FFXGameVersion:    2,
			GameFilesLocation: translatePath,
			ExtractLocation:   extractTempPath,
			TranslateLocation: translatePath,
			ImportLocation:    reimportTempPath,
		}

		// Initialize formatter
		formatter = &formatters.TxtFormatter{
			TargetExtension: ".txt",
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}

		// Setup interaction service
		interactionService = interactions.NewInteractionServiceWithConfig(config)
		interactionService = interactions.NewInteractionWithTextFormatter(formatter)

		// Setup logger
		log = logger.NewLoggerHandler("dlg_file_test")

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocationInfo(locationsBase.WithDirectoryName("extracted"), locationsBase.WithTargetDirectory(extractTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
			TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
			ImportLocation:    locations.NewImportLocationInfo(locationsBase.WithDirectoryName("reimported"), locationsBase.WithTargetDirectory(reimportTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
		}
	})

	BeforeEach(func() {
		// Initialize dialog extractor
		testDlgExtractor = dlg.NewDlgExtractor(log)

		// Initialize dialog compressor
		//testDlgCompressor = dlg.NewDlgCompressor()
	})

	AfterEach(func() {
		err = common.RemoveDir(temp.TempFilePath)
		Expect(err).To(BeNil())
	})

	Describe("Miscellaneous Functionality", func() {
		It("should have valid game version for FFX-2", func() {
			Expect(os.Getenv("FFX_GAME_VERSION")).To(Equal("2"))
		})

		It("should have valid configuration", func() {
			Expect(config).NotTo(BeNil())
		})

		It("should have initialized interaction service", func() {
			Expect(interactionService).NotTo(BeNil())
			Expect(interactionService.FFXAppConfig()).NotTo(BeNil())
		})

		It("should have valid formatter", func() {
			Expect(formatter).NotTo(BeNil())
		})

		It("should have interaction service with text formatter", func() {
			Expect(interactionService.TextFormatter()).NotTo(BeNil())
		})

		It("should have valid destination", func() {
			Expect(destination).NotTo(BeNil())
		})

		It("should have valid logger", func() {
			Expect(log).NotTo(BeNil())
		})
	})

	Describe("Clones functionality", func() {
		It("should not have clones in tutorial.msb", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			gameVersion := interactionService.FFXGameVersion().GetGameVersion()

			source.PopulateDuplicatesFiles(gameVersion)
			Expect(source.Get().ClonedItems).To(HaveLen(0))
		})

		It("should have files clones in bika07_235.bin", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			gameVersion := interactionService.FFXGameVersion().GetGameVersion()

			source.PopulateDuplicatesFiles(gameVersion)
			Expect(source.Get().ClonedItems).To(HaveLen(283))
		})

		It("should have files clones in hiku2800.bin", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\hi\hiku2800\hiku2800.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			gameVersion := interactionService.FFXGameVersion().GetGameVersion()

			source.PopulateDuplicatesFiles(gameVersion)
			Expect(source.Get().ClonedItems).To(HaveLen(1))
		})
	})

	Describe("Extract Functionality", func() {
		/* It("should have correct extract location path", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.txt`
			expected := filepath.Join(extractTempPath, gameVersionDir, testPath)
			expected = filepath.ToSlash(expected)

			actual := destination.Extract().Get().GetTargetFile()
			actual = filepath.ToSlash(actual)

			Expect(actual).To(Equal(expected))
		}) */

		It("should extract the bika07_235.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())
			Expect(err).To(BeNil())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should extract the tutorial.msb successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())
			
			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should extract the cloud.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should extract variant event file crcr0000.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\cr\crcr0000\crcr0000.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should extract variant event file credits.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\event\obj_ps3\cr\credits\credits.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should verify file integrity binary fail", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
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

	Describe("Compress Functionality", func() {
		It("should compress the cloud.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\cloudsave\cloud.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Compress()).To(Succeed())
		})

		It("should compress the tutorial.msb successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Compress()).To(Succeed())
		})

		It("should compress the bika07_235.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			dlgFile := text.NewDialogs(source, destination)
			Expect(dlgFile).NotTo(BeNil())
			Expect(dlgFile.Compress()).To(Succeed())
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

func removeFirstNLines(filePath string, n int) error {
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
