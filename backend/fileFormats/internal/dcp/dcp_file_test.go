package dcp

/* import (
	"ffxresources/backend/common"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDcpFile(t *testing.T) {
	assert := assert.New(t)

	currentDir, err := filepath.Abs(".")
	assert.Nil(err)

	extractTempPath := filepath.Join(common.GetTempDir(), "extract")
	reimportTempPath := filepath.Join(common.GetTempDir(), "reimport")
	translatePath := filepath.Join(currentDir, "/testData/")

	interactions.NewInteraction().FFXGameVersion.SetGameVersionNumber(2)
	interactions.NewInteraction().ExtractLocation.SetTargetDirectory(extractTempPath)
	interactions.NewInteraction().ImportLocation.SetTargetDirectory(reimportTempPath)
	interactions.NewInteraction().TranslateLocation.SetTargetDirectory(translatePath)

	path := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`

	fileInfo := interactions.NewGameDataInfo(path)
	fileInfo.InitializeLocations(formatters.NewTxtFormatter())

	dcpFile := NewDcpFile(fileInfo)
	assert.NotNil(dcpFile)

	dcpFile.GetFileInfo().GetExtractLocation().ProvideTargetPath()
	dcpFile.GetFileInfo().GetImportLocation().ProvideTargetPath()

	err = dcpFile.Extract()
	assert.Nil(err)
} */
