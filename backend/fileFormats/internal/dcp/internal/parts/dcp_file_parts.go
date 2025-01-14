package parts

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/text/dlg"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
)

type DcpFileParts struct {
	*base.FormatsBase
}

func NewDcpFileParts(source interfaces.ISource, destination locations.IDestination) *DcpFileParts {
	source.Get().RelativePath = filepath.Join("system", source.Get().Name)

	return &DcpFileParts{
		FormatsBase: base.NewFormatsBase(source, destination),
	}
}

func (d DcpFileParts) Extract() error {
	dlgFile := dlg.NewDialogs(d.Source(), d.Destination())

	if err := dlgFile.Extract(); err != nil {
		return fmt.Errorf("failed to extract dialog file: %s", d.Source().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Compress() error {
	dlgFile := dlg.NewDialogs(d.Source(), d.Destination())

	if err := dlgFile.Compress(); err != nil {
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
