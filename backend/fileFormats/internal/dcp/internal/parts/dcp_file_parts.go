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

func (d DcpFileParts) Extract() error {
	dlg := dlg.NewDialogs(d.GetFileInfo())

	if err := dlg.Extract(); err != nil {
		return fmt.Errorf("failed to extract dialog file: %s", d.GetFileInfo().GetGameData().Name)
	}

	return nil
}

func (d DcpFileParts) Compress() error {
	dlg := dlg.NewDialogs(d.GetFileInfo())

	if err := dlg.Compress(); err != nil {
		return fmt.Errorf("failed to compress dialog file: %s", d.GetFileInfo().GetGameData().Name)
	}

	return nil
}

func (d DcpFileParts) Validate() error {
	if err := d.GetTranslateLocation().Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.GetGameData().Name)
	}

	return nil
}
