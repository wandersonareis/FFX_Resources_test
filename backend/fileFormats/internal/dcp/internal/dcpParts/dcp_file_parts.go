package dcpParts

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

func NewDcpFileParts(source interfaces.ISource, destination locations.IDestination, formatter interfaces.ITextFormatter) *DcpFileParts {
	source.Get().RelativePath = filepath.Join("system", source.Get().Name)

	destination.InitializeLocations(source, formatter)

	return &DcpFileParts{
		FormatsBase: base.NewFormatsBase(source, destination),
	}
}

func (d DcpFileParts) Extract() error {
	dlgFile := dlg.NewDialogs(d.GetSource(), d.GetDestination())

	if err := dlgFile.Extract(); err != nil {
		return fmt.Errorf("failed to extract dialog file: %s", d.GetSource().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Compress() error {
	dlgFile := dlg.NewDialogs(d.GetSource(), d.GetDestination())

	if err := dlgFile.Compress(); err != nil {
		return fmt.Errorf("failed to compress dialog file: %s", d.GetSource().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Validate() error {
	if err := d.GetDestination().Translate().Get().Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.GetSource().Get().Name)
	}

	return nil
}
