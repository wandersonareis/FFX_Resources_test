package internal

import (
	"ffxresources/backend/formats/internal/dlg"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
)

type DcpFileParts struct {
	gameDataInfo *interactions.GameDataInfo
}

func NewDcpFileParts(dataInfo *interactions.GameDataInfo) *DcpFileParts {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DcpFileParts{
		gameDataInfo: dataInfo,
	}
}

func (d DcpFileParts) GetFileInfo() *interactions.GameDataInfo {
	return d.gameDataInfo
}

func (d DcpFileParts) Extract() {
	dlg := dlg.NewDialogs(d.gameDataInfo)
	dlg.Extract()
}

func (d DcpFileParts) Compress() {
	dlg := dlg.NewDialogs(d.gameDataInfo)
	dlg.Compress()
}

func (d DcpFileParts) Validate() error {
	if err := d.gameDataInfo.TranslateLocation.Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.gameDataInfo.GameData.Name)
	}

	return nil
}