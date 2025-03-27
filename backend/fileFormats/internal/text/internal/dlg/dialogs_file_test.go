package dlg_test

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
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
	testDlgCompressor  dlg.IDlgCompressor
	temp               *common.TempProvider
	gameVersionDir     string
	log                logger.ILoggerHandler
	err                error
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
		testDlgCompressor = dlg.NewDlgCompressor(log)
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

	Describe("Extract Functionality", func() {
		It("should extract the bika07_236.bin successfully", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_236\bika07_236.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err = locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			gameVersion := interactionService.FFXGameVersion().GetGameVersion()
			source.PopulateDuplicatesFiles(gameVersion)

			Expect(destination).NotTo(BeNil())

			destination.InitializeLocations(source, formatter)

			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		It("should extract the dialog file successfully", func() {
			Expect(testDlgExtractor).NotTo(BeNil())
			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())
		})

		/* It("should verify file integrity successfully", func() {
			Expect(testDlgExtractor).NotTo(BeNil())
			Expect(testDlgExtractor.Extract(source, destination)).To(Succeed())

			lockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			targetPath := destination.Extract().Get().GetTargetPath()
			Expect(targetPath).NotTo(BeEmpty())

			Expect(lockitIntegrity.Verify(targetPath, fileOptions)).To(Succeed())
		})

		It("should verify file integrity binary fail", func() {
			lockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			targetPath, err := filepath.Abs(filepath.Join("testdata", gameVersionDir, "binary_missing_linebreak"))
			Expect(err).To(BeNil())

			err = lockitIntegrity.Verify(targetPath, fileOptions)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("error when counting line breaks: the file has 159 line breaks, expected 162"))
		})

		It("should verify file integrity text fail", func() {
			lockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			targetPath, err := filepath.Abs(filepath.Join("testdata", gameVersionDir, "text_missing_linebreak"))
			Expect(err).To(BeNil())

			err = lockitIntegrity.Verify(targetPath, fileOptions)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("error when counting line breaks: the file has 1117 line breaks, expected 1121"))
		})*/
	})

	Describe("Compress Functionality", func() {
		Describe("Compress bika07_235 file", func() {
			BeforeEach(func() {
				testFile := `binary\ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`

				// Setup source and destination
				sourceFile := filepath.Join(testDataPath, testFile)
				source, err = locations.NewSource(sourceFile)
				Expect(err).To(BeNil())

				destination.InitializeLocations(source, formatter)
			})

			It("should compress the lockit file successfully", func() {
				extractPath := filepath.Join(testDataPath, "extracted", "lockit_text")

				destination.Extract().Get().SetTargetPath(extractPath)
				Expect(destination.Extract().Get().GetTargetPath()).To(Equal(extractPath))

				Expect(testDlgCompressor).NotTo(BeNil())
				Expect(testDlgCompressor.Compress(source, destination)).To(Succeed())
			})
		})
		/* It("should compress the lockit file successfully", func() {
			extractPath := filepath.Join(testDataPath, "extracted", "lockit_text")

			destination.Extract().Get().SetTargetPath(extractPath)
			Expect(destination.Extract().Get().GetTargetPath()).To(Equal(extractPath))

			Expect(testDlgCompressor).NotTo(BeNil())
			Expect(testDlgCompressor.Compress(source, destination)).To(Succeed())
		}) */

		/* It("should verify file integrity successfully", func() {
			Expect(testDlgCompressor).NotTo(BeNil())
			Expect(testDlgCompressor.Compress()).To(Succeed())

			lockitIntegrity := integrity.NewLockitFileIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			Expect(lockitIntegrity.Verify(destination, lockitEncoding, fileOptions)).To(Succeed())
		}) */
	})

	Describe("Extract and Compress Functionality by DialogFile", func() {
		It("should extract and compress the DialogFile file successfully", func() {
			lockitFile := text.NewDialogs(source, destination)

			Expect(lockitFile).NotTo(BeNil())
			Expect(lockitFile.Extract()).To(Succeed())
			Expect(lockitFile.Compress()).To(Succeed())
		})
	})

	It("should be dialogs count", func() {
		encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
		defer encoding.Dispose()

		file := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\tutorial.msb`

		source, err := locations.NewSource(file)
		Expect(err).To(BeNil())

		sourceFile := source.Get().Path
		Expect(sourceFile).NotTo(BeEmpty())

		count, err := lib.TextSegmentsCounter(sourceFile, source.Get().Type)
		Expect(err).To(BeNil())
		Expect(count).To(Equal(119))
	})
})
