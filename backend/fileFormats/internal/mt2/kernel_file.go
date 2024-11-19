package mt2

import (
	"ffxresources/backend/fileFormats/internal/mt2/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
)

type kernelFile struct {
	*util.DlgKrnlVerify
}

func NewKernel(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &kernelFile{
		DlgKrnlVerify: util.NewDlgKrnlVerify(dataInfo),
	}
}

func (k kernelFile) Extract() {
	if err := internal.KernelFileExtractor(k.GetFileInfo()); err != nil {
		k.Log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error unpacking kernel file")
		return
	}

	if err := k.VerifyExtract(k.GetExtractLocation()); err != nil {
		k.Log.Error().Err(err).Interface("DialogFile", util.ErrorObject(k.GetFileInfo())).Msg("Error verifying dialog file")
		return
	}

	k.Log.Info().Msgf("Kernel file extracted: %s", k.GetGameData().Name)
}

func (k kernelFile) Compress() {
	if err := internal.KernelFileCompressor(k.GetFileInfo()); err != nil {
		k.Log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error compressing kernel file")
		return
	}

	if err := k.VerifyCompress(k.GetFileInfo(), internal.KernelFileExtractor); err != nil {
		k.Log.Error().Err(err).Interface("kialogFile", util.ErrorObject(k.GetFileInfo())).Msg("Error verifying compressed dialog file")
		return
	}

	k.Log.Info().Msgf("Kernel file compressed: %s", k.GetGameData().Name)
}
