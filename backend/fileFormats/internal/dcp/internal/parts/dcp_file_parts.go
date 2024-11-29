package parts

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/text/dlg"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

type DcpFileParts struct {
	*base.FormatsBase
}

func NewDcpFileParts(dataInfo interactions.IGameDataInfo) *DcpFileParts {
	dataInfo.GetGameData().RelativeGameDataPath = filepath.Join("system", dataInfo.GetGameData().Name)
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DcpFileParts{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (d DcpFileParts) Extract() {
	dlg := dlg.NewDialogs(d.GetFileInfo())
	dlg.Extract()
}

func (d DcpFileParts) Compress() {
	dlg := dlg.NewDialogs(d.GetFileInfo())
	dlg.Compress()
}

func (d DcpFileParts) Validate() error {
	if err := d.GetTranslateLocation().Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.GetGameData().Name)
	}

	return nil
}