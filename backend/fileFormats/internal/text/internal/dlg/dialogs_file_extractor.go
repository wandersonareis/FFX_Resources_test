package dlg

import (
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/dlg/internal"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type (
	IDlgExtractor interface {
		Extract(source interfaces.ISource, destination locations.IDestination) error
	}

	DialogExtractor struct {
		DialogDecoder internal.IDlgDecoder
		Logger        loggingService.ILoggerService
	}
)

func NewDlgExtractor(log loggingService.ILoggerService) IDlgExtractor {
	return &DialogExtractor{
		DialogDecoder: internal.NewDlgDecoder(),
		Logger:        log,
	}
}

func (d *DialogExtractor) Extract(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Extract().ProvideTargetDirectory(); err != nil {
		d.Logger.LogError(err, "Error providing extract path")
		return fmt.Errorf("error providing extract directory: %s", err)
	}

	textEncoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.GetType())
	defer textEncoding.Dispose()

	if err := d.DialogDecoder.Decoder(source, destination, textEncoding); err != nil {
		d.Logger.LogError(err, "Error on decoding dialog file")
		return fmt.Errorf("error on decoding dialog file: %s", err)
	}

	return nil
}
