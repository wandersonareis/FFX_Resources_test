package mt2

import (
	verify "ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/fileFormats/internal/text/mt2/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type kernelFile struct {
	textVerifyer verify.ITextsVerify
	decoder      internal.IKrnlDecoder
	encoder      internal.IKrnlEncoder
	dataInfo     interactions.IGameDataInfo

	log zerolog.Logger
}

func NewKernel(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &kernelFile{
		textVerifyer: verify.NewTextsVerify(),
		decoder:      internal.NewKrnlDecoder(),
		encoder:      internal.NewKrnlEncoder(),
		dataInfo:     dataInfo,
		log:          logger.Get().With().Str("module", "kernel_file").Logger(),
	}
}

func (k kernelFile) GetFileInfo() interactions.IGameDataInfo {
	return k.dataInfo
}

func (k kernelFile) Extract() error {
	if err := k.decoder.Decoder(k.GetFileInfo()); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.GetFileInfo().GetGameData().FullFilePath).
			Msg("Error on decoding kernel file")

		return fmt.Errorf("failed to decode kernel file: %s", k.GetFileInfo().GetGameData().Name)
	}

	if err := k.textVerifyer.VerifyExtract(k.dataInfo.GetExtractLocation()); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.GetFileInfo().GetExtractLocation().TargetFile).
			Msg("Error verifying kernel file")

		return fmt.Errorf("failed to verify kernel file: %s", k.GetFileInfo().GetGameData().Name)
	}

	k.log.Info().Msgf("Kernel file decoded: %s", k.dataInfo.GetGameData().Name)

	return nil
}

func (k kernelFile) Compress() error {
	if err := k.encoder.Encoder(k.GetFileInfo()); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.GetFileInfo().GetTranslateLocation().TargetFile).
			Msg("Error compressing kernel file")

		return fmt.Errorf("failed to compress kernel file: %s", k.GetFileInfo().GetGameData().Name)
	}

	if err := k.textVerifyer.VerifyCompress(k.GetFileInfo(), k.decoder.Decoder); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.GetFileInfo().GetImportLocation().TargetFile).
			Msg("Error verifying compressed dialog file")

		return fmt.Errorf("failed to verify compressed kernel file: %s", k.GetFileInfo().GetImportLocation().TargetFile)
	}

	k.log.Info().
		Str("file", k.GetFileInfo().GetImportLocation().TargetFile).
		Msg("Kernel file compressed")

	return nil
}
