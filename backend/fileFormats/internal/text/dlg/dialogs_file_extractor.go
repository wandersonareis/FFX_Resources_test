package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	"ffxresources/backend/fileFormats/internal/text/lib/dlg_krnl_verify"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"slices"
)

type (
	DlgExtractor struct {
		source       interfaces.ISource
		destination  locations.IDestination
		decoder      internal.IDlgDecoder
		textVerifyer *verify.TextsVerify
		logger       logger.LogHandler
	}
)

func NewDlgExtractor(source interfaces.ISource, destination locations.IDestination) *DlgExtractor {
	return &DlgExtractor{
		source:       source,
		destination:  destination,
		decoder:      internal.NewDlgDecoder(),
		textVerifyer: verify.NewTextsVerify(),

		logger: logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d DlgExtractor) Extract() error {
	if slices.Contains(d.source.Get().ClonedItems, d.source.Get().RelativePath) {
		return nil
	}

	if err := d.decoder.Decoder(d.source, d.destination); err != nil {
		/* 		d.log.Error().
		Err(err).
		Str("file", d.source.Get().Path).
		Msg("Error decoding dialog file") */
		d.logger.LogError(err, "Error decoding dialog file: %s", d.source.Get().Name)

		return fmt.Errorf("failed to decode dialog file: %s", d.source.Get().Name)
	}

	if err := d.textVerifyer.VerifyExtract(d.destination.Extract().Get()); err != nil {
		/* d.log.Error().
		Err(err).
		Str("file", d.destination.Extract().Get().GetTargetFile()).
		Msg("Error verifying text file") */

		d.logger.LogError(err, "Error verifying text file: %s", d.destination.Extract().Get().GetTargetFile())

		return fmt.Errorf("failed to verify text file: %s", d.destination.Extract().Get().GetTargetFile())
	}

	return nil
}
