package locations

import (
	"ffxresources/backend/core/locations/base"
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	ExtractLocation struct {
		internal.LocationBase
	}

	IExtractLocation interface {
		internal.ILocationBase
		interfaces.IValidate
	}
)

func NewExtractLocation(options *internal.LocationBaseOptions) *ExtractLocation {
	return &ExtractLocation{
		LocationBase: internal.NewLocationBase(options),
	}
}

func (e *ExtractLocation) Validate() error {
	e.IsExist = e.IsTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist: %s", e.GetTargetFile())
	}

	return nil
}
