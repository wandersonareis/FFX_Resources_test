package baseFormats

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

type (
	IBaseFileFormat interface {
		GetSource() interfaces.ISource
		GetDestination() locations.IDestination
	}

	BaseFileFormat struct {
		Source      interfaces.ISource
		Destination locations.IDestination
	}
)

func NewFormatsBase(source interfaces.ISource, destination locations.IDestination) IBaseFileFormat {
	return &BaseFileFormat{
		Source:      source,
		Destination: destination,
	}
}

func (f *BaseFileFormat) GetSource() interfaces.ISource {
	return f.Source
}

func (f *BaseFileFormat) GetDestination() locations.IDestination {
	return f.Destination
}
