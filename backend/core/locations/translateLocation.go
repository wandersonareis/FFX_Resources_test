package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	ITranslateLocation interface {
		locationsBase.ILocationBase
		interfaces.IValidate
		Copy() TranslateLocation
	}

	ITargetTranslateLocation interface {
		GetTranslateLocation() ITranslateLocation
	}

	TranslateLocation struct {
		locationsBase.LocationBase
	}
)

func NewTranslateLocation(translateDirectoryName, translateTargetDirectoryPath, gameVersionDir string) ITranslateLocation {
	return &TranslateLocation{
		LocationBase: locationsBase.NewLocationBase(translateDirectoryName, translateTargetDirectoryPath, gameVersionDir),
	}
}

func (t *TranslateLocation) Copy() TranslateLocation {
	return *t
}

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.IsTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
