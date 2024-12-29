package locations

import (
	"ffxresources/backend/core/locations/base"
	"ffxresources/backend/interfaces"
	"fmt"
)

type ITranslateLocation interface {
	internal.ILocationBase
	interfaces.IValidate
}

type ITargetTranslateLocation interface {
	GetTranslateLocation() ITranslateLocation
}

type TranslateLocation struct {
	internal.LocationBase
}

func NewTranslateLocation(options *internal.LocationBaseOptions) *TranslateLocation {
	return &TranslateLocation{
		LocationBase: internal.NewLocationBase(options),
	}
}

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.IsTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
