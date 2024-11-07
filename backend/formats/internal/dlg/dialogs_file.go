package dlg

import (
	"ffxresources/backend/formats/internal/dlg/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

type DialogsFile struct {
	dataInfo *interactions.GameDataInfo
}

func NewDialogs(dataInfo *interactions.GameDataInfo) *DialogsFile {
	dataInfo.ExtractLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)

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
		lib.NotifyError(err)
		return
	}
}

func (d DialogsFile) Compress() {
	err := internal.DialogsTextCompress(d.dataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
