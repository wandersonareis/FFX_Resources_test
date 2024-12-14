package parts

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/text/dlg"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
)

type DcpFileParts struct {
	*base.FormatsBase
}

func NewDcpFileParts(source interfaces.ISource, destination locations.IDestination) *DcpFileParts {
	source.Get().RelativePath = filepath.Join("system", source.Get().Name)
	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	return &DcpFileParts{
		FormatsBase: base.NewFormatsBaseDev(source, destination),
	}
}

func (d DcpFileParts) Extract() error {
	dlg := dlg.NewDialogs(d.Source(), d.Destination())

	if err := dlg.Extract(); err != nil {
		return fmt.Errorf("failed to extract dialog file: %s", d.Source().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Compress() error {
	dlg := dlg.NewDialogs(d.Source(), d.Destination())

	if err := dlg.Compress(); err != nil {
		return fmt.Errorf("failed to compress dialog file: %s", d.Source().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Validate() error {
	if err := d.Destination().Translate().Get().Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.Source().Get().Name)
	}

	return nil
}
