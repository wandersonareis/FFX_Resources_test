package mt2

import (
	"context"
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/mt2/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
)

type kernelFile struct {
	ctx      context.Context
	DataInfo *interactions.GameDataInfo
}

func NewKernel(dataInfo *interactions.GameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &kernelFile{
		ctx:      interactions.NewInteraction().Ctx,
		DataInfo: dataInfo,
	}
}

func (k kernelFile) GetFileInfo() *interactions.GameDataInfo {
	return k.DataInfo
}

func (k kernelFile) Extract() {
	if err := internal.KernelUnpacker(k.GetFileInfo()); err != nil {
		events.NotifyError(err)
		return
	}
}

func (k kernelFile) Compress() {
	if err := internal.KernelTextPacker(k.DataInfo); err != nil {
		events.NotifyError(err)
		return
	}
}
