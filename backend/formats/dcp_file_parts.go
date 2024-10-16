package formats

import "ffxresources/backend/lib"

type DcpFileParts struct {
	FileInfo *lib.FileInfo
}

func NewDcpFileParts(fileInfo *lib.FileInfo) *DcpFileParts {
	fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)

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
	err := dialogsTextPacker(d.FileInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
