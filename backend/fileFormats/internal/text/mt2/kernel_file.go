package mt2

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/fileFormats/internal/text/mt2/internal"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"

	"github.com/rs/zerolog"
)

type kernelFile struct {
	textVerifyer verify.ITextsVerify
	decoder      internal.IKrnlDecoder
	encoder      internal.IKrnlEncoder
	source       interfaces.ISource
	destination  locations.IDestination

	log zerolog.Logger
}

func NewKernel(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	return &kernelFile{
		textVerifyer: verify.NewTextsVerify(),
		decoder:      internal.NewKrnlDecoder(),
		encoder:      internal.NewKrnlEncoder(),
		source:       source,
		destination:  destination,
		log:          logger.Get().With().Str("module", "kernel_file").Logger(),
	}
}

func (k kernelFile) Source() interfaces.ISource {
	return k.source
}

func (k kernelFile) Extract() error {
	if err := k.decoder.Decoder(k.source, k.destination); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.source.Get().Path).
			Msg("Error on decoding kernel file")

		return fmt.Errorf("failed to decode kernel file: %s", k.source.Get().Name)
	}

	if err := k.textVerifyer.VerifyExtract(k.destination.Extract().Get()); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.destination.Extract().Get().GetTargetFile()).
			Msg("Error verifying kernel file")

		return fmt.Errorf("failed to verify kernel file: %s", k.source.Get().Name)
	}

	k.log.Info().Msgf("Kernel file decoded: %s", k.source.Get().Name)

	return nil
}

func (k kernelFile) Compress() error {
	if err := k.encoder.Encoder(k.source, k.destination); err != nil {
		k.log.Error().
			Err(err).
			Str("file", k.destination.Translate().Get().GetTargetFile()).
			Msg("Error compressing kernel file")

		return fmt.Errorf("failed to compress kernel file: %s", k.source.Get().Name)
	}

	importTargetFile := k.destination.Import().Get().GetTargetFile()
	if err := k.textVerifyer.VerifyCompress(k.source, k.destination, k.decoder.Decoder); err != nil {
		k.log.Error().
			Err(err).
			Str("file", importTargetFile).
			Msg("Error verifying compressed dialog file")

		return fmt.Errorf("failed to verify compressed kernel file: %s", importTargetFile)
	}

	k.log.Info().
		Str("file", importTargetFile).
		Msg("Kernel file compressed")

	return nil
}
