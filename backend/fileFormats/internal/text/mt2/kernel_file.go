package mt2

import (
	verify "ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/fileFormats/internal/text/mt2/internal"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type kernelFile struct {
	textVerifyer *verify.DlgKrnlVerify
	dataInfo     interactions.IGameDataInfo

	log zerolog.Logger
}

func NewKernel(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &kernelFile{
		textVerifyer: verify.NewDlgKrnlVerify(),
		dataInfo:     dataInfo,
		log:          logger.Get().With().Str("module", "kernel_file").Logger(),
	}
}

func (k kernelFile) GetFileInfo() interactions.IGameDataInfo {
	return k.dataInfo
}

func (k kernelFile) Extract() {
	if err := internal.KernelFileExtractor(k.GetFileInfo()); err != nil {
		k.log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error unpacking kernel file")
		return
	}

	if err := k.textVerifyer.VerifyExtract(k.dataInfo.GetExtractLocation()); err != nil {
		k.log.Error().Err(err).Interface("DialogFile", util.ErrorObject(k.GetFileInfo())).Msg("Error verifying dialog file")
		return
	}

	k.log.Info().Msgf("Kernel file extracted: %s", k.dataInfo.GetGameData().Name)
}

func (k kernelFile) Compress() {
	if err := internal.KernelFileCompressor(k.GetFileInfo()); err != nil {
		k.log.Error().Err(err).Interface("object", util.ErrorObject(k.GetFileInfo())).Msg("Error compressing kernel file")
		return
	}

	if err := k.textVerifyer.VerifyCompress(k.GetFileInfo(), internal.KernelFileExtractor); err != nil {
		k.log.Error().Err(err).Interface("kialogFile", util.ErrorObject(k.GetFileInfo())).Msg("Error verifying compressed dialog file")
		return
	}

	k.log.Info().Msgf("Kernel file compressed: %s", k.dataInfo.GetGameData().Name)
}
