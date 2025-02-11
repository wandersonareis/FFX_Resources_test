package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
)

type ITranslateLocation interface {
	locationsBase.ILocationBase
	interfaces.IValidate
}

type ITargetTranslateLocation interface {
	GetTranslateLocation() ITranslateLocation
}

type TranslateLocation struct {
	locationsBase.LocationBase
}

func NewTranslateLocation(options *locationsBase.LocationBaseOptions) *TranslateLocation {
	return &TranslateLocation{
		LocationBase: locationsBase.NewLocationBase(options),
	}
}

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.IsTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
