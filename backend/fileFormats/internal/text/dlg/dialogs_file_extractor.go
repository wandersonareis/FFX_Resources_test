package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/dlg/internal"
	"ffxresources/backend/interfaces"
	"fmt"
	"slices"
)

type (
	IDlgExtractor interface {
		interfaces.IExtractor
	}

	DlgExtractor struct {
		source      interfaces.ISource
		destination locations.IDestination
		decoder     internal.IDlgDecoder
	}
)

func NewDlgExtractor(source interfaces.ISource, destination locations.IDestination) *DlgExtractor {
	return &DlgExtractor{
		source:      source,
		destination: destination,
		decoder:     internal.NewDlgDecoder(),
	}
}

func (d *DlgExtractor) Extract() error {
	if slices.Contains(d.source.Get().ClonedItems, d.source.Get().RelativePath) {
		return nil
	}

	if err := d.decoder.Decoder(d.source, d.destination); err != nil {
		return fmt.Errorf("failed to decode dialog file: %s", d.source.Get().Name)
	}

	return nil
}
