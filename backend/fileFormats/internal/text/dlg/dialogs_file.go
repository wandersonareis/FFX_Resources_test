package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	"ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"slices"

	"github.com/rs/zerolog"
)

type DialogsFile struct {
	dialogsClones internal.IDlgClones
	decoder       internal.IDlgDecoder
	encoder       internal.IDlgEncoder
	textVerifyer  verify.ITextsVerify
	source        interfaces.ISource
	destination   locations.IDestination
	log           zerolog.Logger
}

func NewDialogs(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	return &DialogsFile{
		source:      source,
		destination: destination,

		dialogsClones: internal.NewDlgClones(source, destination),
		decoder:       internal.NewDlgDecoder(),
		encoder:       internal.NewDlgEncoder(),
		textVerifyer:  verify.NewTextsVerify(),

		log: logger.Get().With().Str("module", "dialogs_file").Logger(),
	}
}

func (d DialogsFile) Source() interfaces.ISource {
	return d.source
}

func (d DialogsFile) Extract() error {
	if slices.Contains(d.source.Get().ClonedItems, d.source.Get().RelativePath) {
		return nil
	}

	if err := d.decoder.Decoder(d.source, d.destination); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.source.Get().Path).
			Msg("Error decoding dialog file")

		return fmt.Errorf("failed to decode dialog file: %s", d.source.Get().Name)
	}

	if err := d.textVerifyer.VerifyExtract(d.destination.Extract().Get()); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.destination.Extract().Get().GetTargetFile()).
			Msg("Error verifying text file")

		return fmt.Errorf("failed to verify text file: %s", d.destination.Extract().Get().GetTargetFile())
	}

	d.log.Info().
		Str("file", d.destination.Extract().Get().GetTargetFile()).
		Msg("Dialog file extracted successfully")

	return nil
}

func (d DialogsFile) Compress() error {
	if err := d.encoder.Encoder(d.source, d.destination); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.destination.Translate().Get().GetTargetFile()).
			Msg("Error compressing dialog file")

		return fmt.Errorf("failed to compress dialog file: %s", d.destination.Translate().Get().GetTargetFile())
	}

	if err := d.textVerifyer.VerifyCompress(d.source, d.destination, d.decoder.Decoder); err != nil {
		d.log.Error().
			Err(err).
			Str("file", d.destination.Import().Get().GetTargetFile()).
			Msg("Error verifying text file")

		return fmt.Errorf("failed to verify text file: %s", d.destination.Import().Get().GetTargetFile())
	}

	d.dialogsClones.Clone()

	d.log.Info().
		Str("file", d.destination.Import().Get().GetTargetFile()).
		Msg("Dialog file compressed")

	return nil
}
