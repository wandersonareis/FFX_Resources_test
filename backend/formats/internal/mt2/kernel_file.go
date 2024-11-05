package mt2

import (
	"context"
	"ffxresources/backend/formats/internal/mt2/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

type kernelFile struct {
	ctx      context.Context
	DataInfo *interactions.GameDataInfo
}

func NewKernel(dataInfo *interactions.GameDataInfo) interactions.IFileProcessor {
	dataInfo.ExtractLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(formatters.NewTxtFormatter(), dataInfo)

	return &kernelFile{
		ctx:      interactions.NewInteraction().Ctx,
		DataInfo: dataInfo,
	}
}

func (k kernelFile) GetFileInfo() *interactions.GameDataInfo {
	return k.DataInfo
}

func (k kernelFile) Extract() {
	if err := mt2_internal.KernelUnpacker(k.GetFileInfo()); err != nil {
		lib.NotifyError(err)
		return
	}
}

func (k kernelFile) Compress() {
	if err := mt2_internal.KernelTextPacker(k.DataInfo); err != nil {
		lib.NotifyError(err)
		return
	}
}
