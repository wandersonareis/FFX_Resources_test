package fileFormat

import (
	"context"
	"ffxresources/backend/lib"
)

type kernelFile struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewKernel(ctx context.Context, fileInfo lib.FileInfo) lib.IFileProcessor {	
	translatedDirectory, err := lib.NewInteraction().WorkingLocation.ProvideTranslatedDirectory()
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

	return &kernelFile{
		ctx:      ctx,
		FileInfo: fileInfo,
	}
}

func (k kernelFile) GetFileInfo() lib.FileInfo {
	return k.FileInfo
}

func (k kernelFile) Extract() {
	err := kernelUnpacker(k.GetFileInfo())
	if err != nil {
		lib.EmitError(k.ctx, err)
		return
	}
}

func (k kernelFile) Compress() {
	err := kernelTextPacker(k.FileInfo)
	if err != nil {
		lib.EmitError(k.ctx, err)
		return
	}
}
