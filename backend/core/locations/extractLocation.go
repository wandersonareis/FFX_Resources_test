package locations

import (
	"ffxresources/backend/bases"
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	ExtractLocation struct {
		LocationBase
	}

	IExtractLocation interface {
		interfaces.ILocationBase
		interfaces.IValidate
	}
)

func NewExtractLocation(options *bases.LocationBaseOptions) *ExtractLocation {
	return &ExtractLocation{
		LocationBase: NewLocationBase(options),
	}
}

func (e *ExtractLocation) Validate() error {
	e.IsExist = e.IsTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist")
	}

	return nil
}
