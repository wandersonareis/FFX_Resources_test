package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDcpFile(t *testing.T) {
	assert := assert.New(t)

	var interactionService *interactions.InteractionService

	t.Run("Set APP_BASE_PATH", func(t *testing.T) {
		err := os.Setenv("APP_BASE_PATH", `F:\ffxWails\FFX_Resources\build\bin`)
		assert.Nil(err)
	})

	testPath := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`
	currentDir, err := filepath.Abs(".")
	assert.Nil(err)

	temp := common.NewTempProvider("", "")

	extractTempPath := filepath.Join(temp.TempFilePath, "extract")
	defer func() {
		err = common.RemoveDir(extractTempPath)
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
		expected := filepath.Join(extractTempPath, gameVersionDir, lib.DCP_PARTS_TARGET_DIR_NAME)
		expected = filepath.ToSlash(expected)

		actual := destination.Extract().Get().GetTargetPath()
		actual = filepath.ToSlash(actual)

		assert.Equal(expected, actual)
	})


	dcpFile := &DcpFile{
		IBaseFileFormat: &baseFormats.BaseFileFormat{
			Source:      source,
			Destination: destination,
		},
		formatter:   formatter,
		fileOptions: core.NewDcpFileOptions(2),
		log:         logger.NewLoggerHandler("dcp_file_testing"),
	}
	assert.NotNil(dcpFile)

	err = dcpFile.Extract()
	assert.Nil(err)
}
