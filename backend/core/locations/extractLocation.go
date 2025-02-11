package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	ExtractLocation struct {
		locationsBase.LocationBase
	}

	IExtractLocation interface {
		locationsBase.ILocationBase
		interfaces.IValidate
	}
)

func NewExtractLocation(options *locationsBase.LocationBaseOptions) *ExtractLocation {
	return &ExtractLocation{
		LocationBase: locationsBase.NewLocationBase(options),
	}
}

func (e *ExtractLocation) Validate() error {
	e.IsExist = e.IsTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist: %s", e.GetTargetFile())
	}

	return nil
}
