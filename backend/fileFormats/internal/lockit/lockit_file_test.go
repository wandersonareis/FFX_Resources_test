package lockit_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/lockit"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLockit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lockit Suite")
}

var _ = Describe("LockitFile", Ordered, func() {
	var (
		interactionService   *interactions.InteractionService
		extractTempPath      string
		reimportTempPath     string
		translatePath        string
		rootDir              string
		binaryPath           string
		gameVersionDir       string
		testDataPath         string
		gameLocationPath     string
		config               *interactions.FFXAppConfig
		formatter            interfaces.ITextFormatter
		fileOptions          core.ILockitFileOptions
		source               interfaces.ISource
		destination          locations.IDestination
		testLockitExtractor  *lockit.LockitFileExtractor
		testLockitCompressor *lockit.LockitFileCompressor
		lockitEncoding       ffxencoding.IFFXTextLockitEncoding
		temp                 *common.TempProvider
		loggerHandler        *testcommon.MockLogHandler
		err                  error
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
		Expect(config).NotTo(BeNil())

		formatter = &formatters.TxtFormatter{
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}
		Expect(formatter).NotTo(BeNil())

		interactionService = interactions.NewInteractionServiceWithConfig(config)
		Expect(interactionService.FFXAppConfig()).NotTo(BeNil())

		interactionService = interactions.NewInteractionWithTextFormatter(formatter)
		Expect(interactionService.TextFormatter()).NotTo(BeNil())

		gameVersionNumber := interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber()
		Expect(gameVersionNumber).To(Equal(config.FFXGameVersion))

		// Initialize file options
		fileOptions = core.NewLockitFileOptions(gameVersionNumber)
		Expect(fileOptions).NotTo(BeNil())

		loggerHandler = testcommon.NewLogHandlerMock()
		Expect(loggerHandler).NotTo(BeNil())

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
			TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
			ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
		}
		Expect(destination).NotTo(BeNil())
	})

	AfterAll(func() {
		if lockitEncoding != nil {
			lockitEncoding.Dispose()
		}
	})

	BeforeEach(func() {
		file := "ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"
		testFilePath := filepath.Join(gameLocationPath, file)
		Expect(common.CheckPathExists(testFilePath)).To(Succeed())

		// Setup source and destination
		source, err = locations.NewSource(testFilePath)
		Expect(err).To(BeNil())
		Expect(source).NotTo(BeNil())
		Expect(source.GetType()).To(Equal(models.Lockit))

		// Setup destination targets paths using the source
		Expect(destination).NotTo(BeNil())
		Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

		lockitEncoding = ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextUTF8Encoding()

		testLockitExtractor = lockit.NewLockitFileExtractor(source, destination, lockitEncoding, fileOptions, loggerHandler)

		testLockitCompressor = lockit.NewLockitFileCompressor(source, destination, lockitEncoding, fileOptions, loggerHandler)
	})

	AfterEach(func() {
		Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
	})

	Describe("Extract Functionality", func() {
		var lockitExtractIntegrity integrity.ILockitFileExtractorIntegrity

		BeforeEach(func() {
			lockitEncoding = ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextUTF8Encoding()

			testLockitExtractor = lockit.NewLockitFileExtractor(source, destination, lockitEncoding, fileOptions, loggerHandler)
			lockitExtractIntegrity = integrity.NewLockitFileExtractorIntegrity(fileOptions, loggerHandler)
		})

		AfterEach(func() {
			Expect(lockitEncoding).NotTo(BeNil())
			lockitEncoding.Dispose()
			lockitEncoding = nil
		})

		/* It("should have correct extract location path", func() {
			expected := filepath.Join(extractTempPath, gameVersionDir, "lockit_text")
			expected = filepath.ToSlash(expected)

			actual := destination.Extract().GetTargetPath()
			actual = filepath.ToSlash(actual)

			Expect(actual).To(Equal(expected))
		}) */

		It("should extract the lockit file successfully", func() {
			Expect(testLockitExtractor).NotTo(BeNil())
			Expect(testLockitExtractor.Extract()).To(Succeed())
		})

		It("should verify file integrity successfully", func() {
			Expect(testLockitExtractor).NotTo(BeNil())
			Expect(testLockitExtractor.Extract()).To(Succeed())

			Expect(lockitExtractIntegrity).NotTo(BeNil())

			targetPath := destination.Extract().GetTargetPath()
			Expect(targetPath).NotTo(BeEmpty())

			Expect(lockitExtractIntegrity.Verify(targetPath)).To(Succeed())
		})

		It("should verify file integrity binary fail", func() {
			Expect(lockitExtractIntegrity).NotTo(BeNil())

			targetPath, err := filepath.Abs(filepath.Join("testdata", gameVersionDir, "binary_missing_linebreak"))
			Expect(err).To(BeNil())

			err = lockitExtractIntegrity.Verify(targetPath)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("error validating line breaks count for extracted lockit binary file parts: error validating line breaks count for lockit file parts: error when counting line breaks: the file has 159 line breaks, expected 162"))
		})

		It("should verify file integrity text fail", func() {
			Expect(lockitExtractIntegrity).NotTo(BeNil())

			targetPath, err := filepath.Abs(filepath.Join("testdata", gameVersionDir, "text_missing_linebreak"))
			Expect(err).To(BeNil())

			err = lockitExtractIntegrity.Verify(targetPath)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("error validating line breaks count for extracted lockit text file parts: error validating line breaks count for lockit file parts: error when counting line breaks: the file has 1117 line breaks, expected 1121"))
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

			lockitIntegrity := integrity.NewLockitFileIntegrity(loggerHandler)
			Expect(lockitIntegrity).NotTo(BeNil())

			Expect(lockitIntegrity.Verify(destination, lockitEncoding, fileOptions)).To(Succeed())
		})
	})
})
