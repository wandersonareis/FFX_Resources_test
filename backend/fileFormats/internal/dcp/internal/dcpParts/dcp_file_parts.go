package dcpParts

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
)

type DcpFileParts struct {
	baseFormats.IBaseFileFormat
}

func NewDcpFileParts(source interfaces.ISource, destination locations.IDestination, formatter interfaces.ITextFormatter) *DcpFileParts {
	source.Get().RelativePath = filepath.Join("system", source.Get().Name)

	destination.InitializeLocations(source, formatter)

	return &DcpFileParts{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
	}
}

func (d DcpFileParts) Extract() error {
	dlgFile := text.NewDialogs(d.GetSource(), d.GetDestination())

	if err := dlgFile.Extract(); err != nil {
		return fmt.Errorf("failed to extract dialog file: %s", d.GetSource().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Compress() error {
	dlgFile := text.NewDialogs(d.GetSource(), d.GetDestination())

	if err := dlgFile.Compress(); err != nil {
		return fmt.Errorf("failed to compress dialog file: %s", d.GetSource().Get().Name)
	}

	return nil
}

func (d DcpFileParts) Validate() error {
	if err := d.GetDestination().Translate().Validate(); err != nil {
		return fmt.Errorf("translated dcp parts file not found: %s", d.GetSource().Get().Name)
	}

	return nil
}
