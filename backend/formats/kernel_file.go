package formats

import (
	"context"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

type kernelFile struct {
	ctx      context.Context
	DataInfo *interactions.GameDataInfo
}

func NewKernel(dataInfo *interactions.GameDataInfo) interactions.IFileProcessor {
	/* relativePath, err := common.GetRelativePathFromMarker(fileInfo.AbsolutePath)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	fileInfo.RelativePath = relativePath */

	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	return &kernelFile{
		ctx:      interactions.NewInteraction().Ctx,
		DataInfo: dataInfo,
	}
}

func (k kernelFile) GetFileInfo() *interactions.GameDataInfo {
	return k.DataInfo
}

func (k kernelFile) Extract() {
	err := kernelUnpacker(k.GetFileInfo())
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func (k kernelFile) Compress() {
	err := kernelTextPacker(k.DataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}
