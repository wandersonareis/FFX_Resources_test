package formats

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

type DialogsFile struct {
	dataInfo *interactions.GameDataInfo
}

func NewDialogs(dataInfo *interactions.GameDataInfo) *DialogsFile {
	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	return &DialogsFile{
		dataInfo: dataInfo,
	}
}

func (d DialogsFile) GetFileInfo() *interactions.GameDataInfo {
	return d.dataInfo
}

func (d DialogsFile) Extract() {
	err := dialogsUnpacker(d.dataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (d DialogsFile) Compress() {
	err := dialogsTextCompress(d.dataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
