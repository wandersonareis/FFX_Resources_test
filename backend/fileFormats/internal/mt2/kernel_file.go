package mt2

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/mt2/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
)

type kernelFile struct {
	*base.FormatsBase
}

func NewKernel(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &kernelFile{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (k kernelFile) Extract() {
	if err := internal.KernelUnpacker(k.GetFileInfo()); err != nil {
		k.Log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error unpacking kernel file")
		return
	}
}

func (k kernelFile) Compress() {
	if err := internal.KernelTextPacker(k.GetFileInfo()); err != nil {
		k.Log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error compressing kernel file")
		return
	}
}
