package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	IDlgCompressor interface {
		Compress(source interfaces.ISource, destination locations.IDestination) error
	}

	dialogCompressor struct {
		dialogsClones internal.IDlgClones
		encoder       internal.IDlgEncoder
		logger        logger.ILoggerHandler
	}
)

func NewDlgCompressor(logger logger.ILoggerHandler) IDlgCompressor {
	return &dialogCompressor{
		dialogsClones: internal.NewDlgClones(logger),
		encoder:       internal.NewDlgEncoder(logger),
		logger:        logger,
	}
}

func (d *dialogCompressor) Compress(source interfaces.ISource, destination locations.IDestination) error {
	if err := d.encoder.Encoder(source, destination); err != nil {
		return fmt.Errorf("failed to compress dialog file: %s", destination.Translate().Get().GetTargetFile())
	}

	d.dialogsClones.Clone(source, destination)

	return nil
}
