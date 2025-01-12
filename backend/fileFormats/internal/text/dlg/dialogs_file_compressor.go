package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	IDlgCompressor interface {
		interfaces.ICompressor
	}

	DlgCompressor struct {
		source        interfaces.ISource
		destination   locations.IDestination
		dialogsClones internal.IDlgClones
		encoder       internal.IDlgEncoder

		log logger.LogHandler
	}
)

func NewDlgCompressor(source interfaces.ISource, destination locations.IDestination) *DlgCompressor {
	return &DlgCompressor{
		source:      source,
		destination: destination,

		dialogsClones: internal.NewDlgClones(source, destination),
		encoder:       internal.NewDlgEncoder(),

		log: logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d *DlgCompressor) Compress() error {
	translateLocation := d.destination.Translate().Get()

	if err := d.encoder.Encoder(d.source, d.destination); err != nil {
		d.log.LogError(err, "Error compressing dialog file: %s", translateLocation.GetTargetFile())

		return fmt.Errorf("failed to compress dialog file: %s", translateLocation.GetTargetFile())
	}

	d.dialogsClones.Clone()

	return nil
}
