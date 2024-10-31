package formats

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

type DcpFileParts struct {
	DataInfo *interactions.GameDataInfo
}

func NewDcpFileParts(dataInfo *interactions.GameDataInfo) *DcpFileParts {
	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	return &DcpFileParts{
		DataInfo: dataInfo,
	}
}

func (d DcpFileParts) GetFileInfo() *interactions.GameDataInfo {
	return d.DataInfo
}

func (d DcpFileParts) Extract() {
	err := dialogsUnpacker(d.DataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (d DcpFileParts) Compress() {
	err := dialogsTextCompress(d.DataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
