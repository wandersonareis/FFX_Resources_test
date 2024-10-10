package fileFormat

import (
	"context"
	"ffxresources/backend/lib"
)

type DialogsFile struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewDialogs(ctx context.Context, fileInfo lib.FileInfo) *DialogsFile {
	if lib.IsEmptyOrWhitespace(&fileInfo.ExtractLocation.TargetFile) || lib.IsEmptyOrWhitespace(&fileInfo.TranslatedFile) {		
		translatedDirectory, err := lib.GetWorkdirectory().ProvideTranslatedDirectory()
		if err != nil {
			lib.EmitError(ctx, err)
			return nil
		}

		relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
		if err != nil {
			lib.EmitError(ctx, err)
			return nil
		}

		fileInfo.RelativePath = relativePath

		translatedFile, translatedPath := lib.GeneratedTranslatedOutput(fileInfo, translatedDirectory)
		fileInfo.TranslatedFile = translatedFile
		fileInfo.TranslatedPath = translatedPath

		fileInfo.ExtractLocation = *lib.NewExtractLocation()

		fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	}

	return &DialogsFile{
		ctx:      ctx,
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
