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
	IDlgExtractor interface {
		Extract(source interfaces.ISource, destination locations.IDestination) error
	}

	dlgExtractor struct {
		decoder internal.IDlgDecoder
		logger logger.ILoggerHandler
	}
)

func NewDlgExtractor(logger logger.ILoggerHandler) IDlgExtractor {
	return &dlgExtractor{
		decoder: internal.NewDlgDecoder(logger),
		logger: logger,
	}
}

func (d *dlgExtractor) Extract(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Extract().ProvideTargetDirectory(); err != nil {
		return fmt.Errorf("failed to provide target directory: %s", err)
	}
	
	textEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
	defer textEncoding.Dispose()

	if err := d.decoder.Decoder(source, destination, textEncoding); err != nil {
		return fmt.Errorf("failed to decode dialog file: %s", source.Get().Name)
	}

	return nil
}
