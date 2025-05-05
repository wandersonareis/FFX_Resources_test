package dlg

import (
	ffxencoding "ffxresources/backend/core/encoding"
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
		DialogEncoder internal.IDlgEncoder
		Logger        logger.ILoggerHandler
	}
)

func NewDlgCompressor(logger logger.ILoggerHandler) IDlgCompressor {
	return &dialogCompressor{
		dialogsClones: internal.NewDlgClones(logger),
		DialogEncoder: internal.NewDlgEncoder(),
		Logger:        logger,
	}
}

func (d *dialogCompressor) Compress(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Import().ProvideTargetPath(); err != nil {
		outputPath := destination.Import().GetTargetPath()

		return fmt.Errorf("error providing import path: %s | error: %w", outputPath, err)
	}

	textEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
	defer textEncoding.Dispose()

	if err := d.DialogEncoder.Encoder(source, destination, textEncoding); err != nil {
		d.Logger.LogError(err, "Error on compressing dialog file")
		return fmt.Errorf("error on compressing dialog file: %s", err)
	}

	d.dialogsClones.Clone(source, destination)

	return nil
}
