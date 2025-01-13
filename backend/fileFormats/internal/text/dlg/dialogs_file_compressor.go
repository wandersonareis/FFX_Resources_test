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
		Compress(source interfaces.ISource, destination locations.IDestination) error
	}

	DlgCompressor struct {
		source        interfaces.ISource
		destination   locations.IDestination
		dialogsClones internal.IDlgClones
		encoder       internal.IDlgEncoder

		log logger.LogHandler
	}
)

func newDlgCompressor() *DlgCompressor {
	return &DlgCompressor{
		dialogsClones: internal.NewDlgClones(),
		encoder:       internal.NewDlgEncoder(),

		log: logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d *DlgCompressor) Compress(source interfaces.ISource, destination locations.IDestination) error {
	translateLocation := d.destination.Translate().Get()

	if err := d.encoder.Encoder(d.source, d.destination); err != nil {
		d.log.LogError(err, "Error compressing dialog file: %s", translateLocation.GetTargetFile())

		return fmt.Errorf("failed to compress dialog file: %s", translateLocation.GetTargetFile())
	}

	d.dialogsClones.Clone(source, destination)

	return nil
}
