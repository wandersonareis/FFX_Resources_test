package dlg

import (
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/dlg/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
)

type DialogsFile struct {
	dataInfo *interactions.GameDataInfo
}

func NewDialogs(dataInfo *interactions.GameDataInfo) *DialogsFile {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &DialogsFile{
		dataInfo: dataInfo,
	}
}

func (d DialogsFile) GetFileInfo() *interactions.GameDataInfo {
	return d.dataInfo
}

func (d DialogsFile) Extract() {
	err := internal.DialogsUnpacker(d.dataInfo)
	if err != nil {
		events.NotifyError(err)
		return
	}
}

func (d DialogsFile) Compress() {
	err := internal.DialogsTextCompress(d.dataInfo)
	if err != nil {
		events.NotifyError(err)
		return
	}
}
