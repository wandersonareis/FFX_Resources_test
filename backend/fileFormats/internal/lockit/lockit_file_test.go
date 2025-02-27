package lockit

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockitFile(t *testing.T) {
	assert := assert.New(t)

	var interactionService *interactions.InteractionService

	err := os.Setenv("APP_BASE_PATH", `F:\ffxWails\FFX_Resources\build\bin`)
	assert.Nil(err)

	testPath := "F:\\ffxWails\\FFX_Resources\\build\\bin\\data\\ffx-2_data\\gamedata\\ps3data\\lockit\\ffx2_loc_kit_ps3_us.bin"
	currentDir, err := filepath.Abs(".")
	assert.Nil(err)

	temp := common.NewTempProvider("", "")

	extractTempPath := filepath.Join(temp.TempFilePath, "extract")
	defer func() {
		err = common.RemoveDir(temp.TempFilePath)
		assert.Nil(err)
	}()

	reimportTempPath := filepath.Join(temp.TempFilePath, "reimport")
	translatePath := filepath.Join(currentDir, "/testData/")

	config := &interactions.FFXAppConfig{
		FFXGameVersion:    2,
		GameFilesLocation: translatePath,
		ExtractLocation:   extractTempPath,
		TranslateLocation: translatePath,
		ImportLocation:    reimportTempPath,
	}

	t.Run("Config is not nil", func(t *testing.T) {
		assert.NotNil(config)
	})

	t.Run("Initialize Interaction Service", func(t *testing.T) {
		interactionService = interactions.NewInteractionServiceWithConfig(config)
		assert.NotNil(interactionService)
		assert.NotNil(interactionService.FFXAppConfig())
	})

	gameVersionDir := "FFX-2"

	formatter := &formatters.TxtFormatter{
		TargetExtension: ".txt",
		GameVersionDir:  gameVersionDir,
		GameFilesPath:   translatePath,
	}
	t.Run("Formatter is not nil", func(t *testing.T) {
		assert.NotNil(formatter)
	})

	interactionService = interactions.NewInteractionWithTextFormatter(formatter)

	t.Run("Interaction service formatter is not nil", func(t *testing.T) {
		assert.NotNil(interactionService.TextFormatter())
	})

	source, err := locations.NewSource(testPath)
	assert.Nil(err)

	destination := &locations.Destination{
		ExtractLocation:   locations.NewExtractLocationInfo(locationsBase.WithDirectoryName("extracted"), locationsBase.WithTargetDirectory(extractTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
		TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
		ImportLocation:    locations.NewImportLocationInfo(locationsBase.WithDirectoryName("reimported"), locationsBase.WithTargetDirectory(reimportTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
	}

	t.Run("Destination is not nil", func(t *testing.T) {
		assert.NotNil(destination)
	})
	destination.InitializeLocations(source, formatter)

	t.Run("Destination extract location path is temporary path", func(t *testing.T) {
		expected := filepath.Join(extractTempPath, gameVersionDir, "lockit_text")
		expected = filepath.ToSlash(expected)

		actual := destination.Extract().Get().GetTargetPath()
		actual = filepath.ToSlash(actual)

		assert.Equal(expected, actual)
	})

	lockitEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer lockitEncoding.Dispose()
	assert.NotNil(lockitEncoding)

	fileOptions := core.NewLockitFileOptions(interactions.NewInteractionService().FFXGameVersion().GetGameVersionNumber())
	assert.NotNil(fileOptions)

	t.Run("Initialize Lockit Extractor", func(t *testing.T) {
		testLockitExtractor := &LockitFileExtractor{
			IBaseFileFormat:  baseFormats.NewFormatsBase(source, destination),
			filePartsDecoder: lockitParts.NewLockitFilePartsDecoder(),
			lockitEncoding:   lockitEncoding,
			options:          fileOptions,
			log:              logger.NewLoggerHandler("lockit_file_testing"),
		}
		assert.NotNil(testLockitExtractor)

		err = testLockitExtractor.Extract()
		assert.Nil(err)
	})

	t.Run("Initialize lockit extracted file integrity check", func(t *testing.T) {
		testLockitIntegrity := integrity.NewLockitFileExtractorIntegrity(logger.NewLoggerHandler("lockit_file_integrity_testing"))

		err := testLockitIntegrity.VerifyFileIntegrity(destination, fileOptions)
		assert.Nil(err)
	})
}
