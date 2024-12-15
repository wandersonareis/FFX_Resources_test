package locations

import (
	"ffxresources/backend/bases"
	"ffxresources/backend/interfaces"
	"fmt"
)

type ITranslateLocation interface {
	interfaces.ILocationBase
	interfaces.IValidate
}

type ITargetTranslateLocation interface {
	GetTranslateLocation() ITranslateLocation
}

type TranslateLocation struct {
	LocationBase
}

func NewTranslateLocation(options *bases.LocationBaseOptions) *TranslateLocation {
	return &TranslateLocation{
		LocationBase: NewLocationBase(options),
	}
}

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.IsTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
