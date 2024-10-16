package formats

import (
	"context"
	"ffxresources/backend/lib"
)

type DialogsFile struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewDialogs(fileInfo lib.FileInfo) *DialogsFile {
	//if lib.IsEmptyOrWhitespace(&fileInfo.ExtractLocation.TargetFile) || lib.IsEmptyOrWhitespace(&fileInfo.TranslatedFile) {

	relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	fileInfo.RelativePath = relativePath

	fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	//}

	return &DialogsFile{
		ctx:      lib.NewInteraction().Ctx,
		FileInfo: fileInfo,
	}
}

func (d DialogsFile) GetFileInfo() lib.FileInfo {
	return d.FileInfo
}

func (d DialogsFile) Extract() {
	err := dialogsUnpacker(d.FileInfo)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}
}

func (d DialogsFile) Compress() {
	err := dialogsTextPacker(d.FileInfo)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}
}
