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

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("LockitFile", func() {
	var (
		interactionService  *interactions.InteractionService
		extractTempPath     string
		reimportTempPath    string
		translatePath       string
		config              *interactions.FFXAppConfig
		formatter           interfaces.ITextFormatter
		fileOptions         core.ILockitFileOptions
		source              interfaces.ISource
		destination         locations.IDestination
		testLockitExtractor *lockit.LockitFileExtractor
		lockitEncoding      ffxencoding.IFFXTextLockitEncoding
		temp                *common.TempProvider
		gameVersionDir      string
		err                 error
	)

	ginkgo.BeforeEach(func() {
		gomega.Expect(os.Setenv("APP_BASE_PATH", `F:\ffxWails\FFX_Resources\build\bin`)).To(gomega.Succeed())

		testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"
		currentDir, err := filepath.Abs(".")
		gomega.Expect(err).To(gomega.BeNil())

		temp = common.NewTempProvider("", "")

		extractTempPath = filepath.Join(temp.TempFilePath, "extract")
		reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
		translatePath = filepath.Join(currentDir, "/testData/")

		gameVersionDir = "FFX-2"

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

		// Setup source and destination
		source, err = locations.NewSource(testPath)
		gomega.Expect(err).To(gomega.BeNil())

		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocationInfo(locationsBase.WithDirectoryName("extracted"), locationsBase.WithTargetDirectory(extractTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
			TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
			ImportLocation:    locations.NewImportLocationInfo(locationsBase.WithDirectoryName("reimported"), locationsBase.WithTargetDirectory(reimportTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
		}

		destination.InitializeLocations(source, formatter)

		// Initialize encoding and file options
		lockitEncoding = ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
		fileOptions = core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())

		// Initialize lockit extractor
		testLockitExtractor = lockit.NewLockitFileExtractor(source, destination, lockitEncoding, logger.NewLoggerHandler("lockit_file_testing"))
	})

	ginkgo.AfterEach(func() {
		if lockitEncoding != nil {
			lockitEncoding.Dispose()
		}
		err = common.RemoveDir(temp.TempFilePath)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("should have valid configuration", func() {
		gomega.Expect(config).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have initialized interaction service", func() {
		gomega.Expect(interactionService).NotTo(gomega.BeNil())
		gomega.Expect(interactionService.FFXAppConfig()).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have valid formatter", func() {
		gomega.Expect(formatter).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have interaction service with text formatter", func() {
		gomega.Expect(interactionService.TextFormatter()).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have valid destination", func() {
		gomega.Expect(destination).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have correct extract location path", func() {
		expected := filepath.Join(extractTempPath, gameVersionDir, "lockit_text")
		expected = filepath.ToSlash(expected)

		actual := destination.Extract().Get().GetTargetPath()
		actual = filepath.ToSlash(actual)

		gomega.Expect(actual).To(gomega.Equal(expected))
	})

	ginkgo.It("should extract the lockit file successfully", func() {
		gomega.Expect(testLockitExtractor).NotTo(gomega.BeNil())
		gomega.Expect(testLockitExtractor.Extract()).To(gomega.Succeed())
	})

	ginkgo.It("should verify file integrity successfully", func() {
		gomega.Expect(testLockitExtractor).NotTo(gomega.BeNil())
		gomega.Expect(testLockitExtractor.Extract()).To(gomega.Succeed())

		lockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))
		gomega.Expect(lockitIntegrity).NotTo(gomega.BeNil())

		gomega.Expect(lockitIntegrity.VerifyFileIntegrity(destination, fileOptions)).To(gomega.Succeed())
	})
})
