package dcp

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDcpFile(t *testing.T) {
	assert := assert.New(t)

	testPath := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`
	currentDir, err := filepath.Abs(".")
	assert.Nil(err)

	gameVersionDir := "FFX-2"
	extractTempPath := filepath.Join(common.GetTempDir(), "extract")
	defer func() {
		err = common.RemoveDir(extractTempPath)
		assert.Nil(err)
	}()

	reimportTempPath := filepath.Join(common.GetTempDir(), "reimport")
	translatePath := filepath.Join(currentDir, "/testData/")

	config := &interactions.FFXAppConfig{
		FFXGameVersion:    2,
		GameFilesLocation: translatePath,
		ExtractLocation:   extractTempPath,
		TranslateLocation: translatePath,
		ImportLocation:    reimportTempPath,
	}

	interactionService := interactions.NewInteractionServiceWithConfig(config)
	assert.NotNil(interactionService)

	formatter := &formatters.TxtFormatter{
		TargetExtension: ".txt",
		GameVersionDir:  gameVersionDir,
		GameFilesPath:   translatePath,
	}
	assert.NotNil(formatter)

	interactionService = interactions.NewInteractionWithTextFormatter(formatter)
	assert.NotNil(interactionService)

	source, err := locations.NewSource(testPath)
	assert.Nil(err)

	destination := &locations.Destination{
		ExtractLocation:   locations.NewExtractLocationInfo(locationsBase.WithDirectoryName("extracted"), locationsBase.WithTargetDirectory(extractTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
		TranslateLocation: locations.NewTranslateLocationInfo(locationsBase.WithDirectoryName("translated"), locationsBase.WithTargetDirectory(translatePath), locationsBase.WithGameVersionDir(gameVersionDir)),
		ImportLocation:    locations.NewImportLocationInfo(locationsBase.WithDirectoryName("reimported"), locationsBase.WithTargetDirectory(reimportTempPath), locationsBase.WithGameVersionDir(gameVersionDir)),
	}
	assert.NotNil(destination)

	destination.InitializeLocations(source, formatter)

	dcpFile := &DcpFile{
		FormatsBase: &base.FormatsBase{
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
