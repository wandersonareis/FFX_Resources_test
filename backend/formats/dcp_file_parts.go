package formats

import "ffxresources/backend/lib"

type DcpFileParts struct {
	FileInfo *lib.FileInfo
}

func NewDcpFileParts(fileInfo *lib.FileInfo) *DcpFileParts {
	fileInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), fileInfo)

	return &DcpFileParts{
		FileInfo: fileInfo,
	}
}

func (d DcpFileParts) GetFileInfo() *lib.FileInfo {
	return d.FileInfo
}

func (d DcpFileParts) Extract() {
	err := dialogsUnpacker(d.FileInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (d DcpFileParts) Compress() {
	err := dialogsTextCompress(d.FileInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
