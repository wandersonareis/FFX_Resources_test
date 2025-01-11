package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	verify "ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	DlgCompressor struct {
		source        interfaces.ISource
		destination   locations.IDestination
		dialogsClones internal.IDlgClones
		encoder       internal.IDlgEncoder
		textVerifyer  verify.ITextsVerify
		log           logger.LogHandler
	}
)

func NewDlgCompressor(source interfaces.ISource, destination locations.IDestination) *DlgCompressor {
	return &DlgCompressor{
		source:      source,
		destination: destination,

		dialogsClones: internal.NewDlgClones(source, destination),
		encoder:       internal.NewDlgEncoder(),
		textVerifyer:  verify.NewTextsVerify(),

		log: logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d *DlgCompressor) Compress() error {
	if err := d.encoder.Encoder(d.source, d.destination); err != nil {
		/* d.log.Error().
		Err(err).
		Str("file", d.destination.Translate().Get().GetTargetFile()).
		Msg("Error compressing dialog file") */
		d.log.LogError(err, "Error compressing dialog file: %s", d.destination.Translate().Get().GetTargetFile())

		return fmt.Errorf("failed to compress dialog file: %s", d.destination.Translate().Get().GetTargetFile())
	}

	decoder := internal.NewDlgDecoder()

	if err := d.textVerifyer.VerifyCompress(d.source, d.destination, decoder.Decoder); err != nil {
		/* d.log.Error().
		Err(err).
		Str("file", d.destination.Import().Get().GetTargetFile()).
		Msg("Error verifying text file") */

		d.log.LogError(err, "Error verifying text file: %s", d.destination.Import().Get().GetTargetFile())

		return fmt.Errorf("failed to verify text file: %s", d.destination.Import().Get().GetTargetFile())
	}

	d.dialogsClones.Clone()

	return nil
}
