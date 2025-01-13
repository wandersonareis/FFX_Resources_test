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
		Extract(source interfaces.ISource, destination locations.IDestination) error
	}

	DlgExtractor struct {
		decoder     internal.IDlgDecoder
	}
)

func NewDlgExtractor() *DlgExtractor {
	return &DlgExtractor{
		decoder:     internal.NewDlgDecoder(),
	}
}

func (d *DlgExtractor) Extract(source interfaces.ISource, destination locations.IDestination) error {
	if slices.Contains(source.Get().ClonedItems, source.Get().RelativePath) {
		return nil
	}

	if err := d.decoder.Decoder(source, destination); err != nil {
		return fmt.Errorf("failed to decode dialog file: %s", source.Get().Name)
	}

	return nil
}
