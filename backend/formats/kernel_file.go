package formats

import (
	"context"
	"ffxresources/backend/lib"
)

type kernelFile struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewKernel(fileInfo lib.FileInfo) lib.IFileProcessor {
	relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	fileInfo.RelativePath = relativePath

	fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.TranslateLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)
	fileInfo.ImportLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)

	return &kernelFile{
		ctx:      lib.NewInteraction().Ctx,
		FileInfo: fileInfo,
	}
}

func (k kernelFile) GetFileInfo() lib.FileInfo {
	return k.FileInfo
}

func (k kernelFile) Extract() {
	err := kernelUnpacker(k.GetFileInfo())
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (k kernelFile) Compress() {
	err := kernelTextPacker(k.FileInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
