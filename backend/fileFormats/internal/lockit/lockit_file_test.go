package lockit_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/lockit"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LockitFile", Ordered, func() {
	var (
		interactionService   *interactions.InteractionService
		extractTempPath      string
		reimportTempPath     string
		translatePath        string
		testDataPath         string
		config               *interactions.FFXAppConfig
		formatter            interfaces.ITextFormatter
		fileOptions          core.ILockitFileOptions
		source               interfaces.ISource
		destination          locations.IDestination
		testLockitExtractor  *lockit.LockitFileExtractor
		testLockitCompressor *lockit.LockitFileCompressor
		lockitEncoding       ffxencoding.IFFXTextLockitEncoding
		temp                 *common.TempProvider
		gameVersionDir       string
		loggerHandler        logger.ILoggerHandler
		err                  error
	)

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

		// Initialize file options
		fileOptions = core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())

		// Initialize logger
		loggerHandler = logger.NewLoggerHandler("lockit_file_testing")

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
			TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
			ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
		}
	})

	AfterAll(func() {
		if lockitEncoding != nil {
			lockitEncoding.Dispose()
		}
	})

	BeforeEach(func() {
		testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"

		// Setup source and destination
		source, err = locations.NewSource(testPath)
		Expect(err).To(BeNil())

		destination.InitializeLocations(source, formatter)

		// Initialize lockit encoding
		lockitEncoding = ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()

		// Initialize lockit extractor
		testLockitExtractor = lockit.NewLockitFileExtractor(source, destination, lockitEncoding, fileOptions, loggerHandler)

		testLockitCompressor = lockit.NewLockitFileCompressor(source, destination, lockitEncoding, fileOptions, loggerHandler)
	})

	AfterEach(func() {
		common.RemoveDir(temp.TempFilePath)
	})

	Describe("Miscellaneous Functionality", func() {
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

		It("should have valid source", func() {
			Expect(source).NotTo(BeNil())
		})

		It("should have valid lockit encoding", func() {
			Expect(lockitEncoding).NotTo(BeNil())
		})

		It("should have valid file options", func() {
			Expect(fileOptions).NotTo(BeNil())
		})

		It("should have valid lockit extractor", func() {
			Expect(testLockitExtractor).NotTo(BeNil())
		})

		It("should have valid lockit compressor", func() {
			Expect(testLockitCompressor).NotTo(BeNil())
		})
	})

	Describe("Extract Functionality", func() {
		It("should have correct extract location path", func() {
			expected := filepath.Join(extractTempPath, gameVersionDir, "lockit_text")
			expected = filepath.ToSlash(expected)

			actual := destination.Extract().GetTargetPath()
			actual = filepath.ToSlash(actual)

			Expect(actual).To(Equal(expected))
		})

		It("should extract the lockit file successfully", func() {
			Expect(testLockitExtractor).NotTo(BeNil())
			Expect(testLockitExtractor.Extract()).To(Succeed())
		})

		It("should verify file integrity successfully", func() {
			Expect(testLockitExtractor).NotTo(BeNil())
			Expect(testLockitExtractor.Extract()).To(Succeed())

			lockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			targetPath := destination.Extract().GetTargetPath()
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
		})
	})

	Describe("Compress Functionality", func() {
		It("should compress the lockit file successfully", func() {
			extractPath := filepath.Join(testDataPath, "extracted", "lockit_text")

			destination.Extract().SetTargetPath(extractPath)
			Expect(destination.Extract().GetTargetPath()).To(Equal(extractPath))

			Expect(testLockitCompressor).NotTo(BeNil())
			Expect(testLockitCompressor.Compress()).To(Succeed())
		})

		It("should verify file integrity successfully", func() {
			Expect(testLockitCompressor).NotTo(BeNil())
			Expect(testLockitCompressor.Compress()).To(Succeed())

			lockitIntegrity := integrity.NewLockitFileIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
			Expect(lockitIntegrity).NotTo(BeNil())

			Expect(lockitIntegrity.Verify(destination, lockitEncoding, fileOptions)).To(Succeed())
		})
	})
})
