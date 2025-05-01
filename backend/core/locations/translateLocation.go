package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
)

type (
	ITranslateLocation interface {
		locationsBase.ILocationBase
		interfaces.IValidate
	}

	ITargetTranslateLocation interface {
		GetTranslateLocation() ITranslateLocation
	}

	TranslateLocation struct {
		locationsBase.LocationBase
	}
)

func NewTranslateLocation(translateDirectoryName, translateTargetDirectory, gameVersionDir string) ITranslateLocation {
	targetDirectory := filepath.Join(translateTargetDirectory, gameVersionDir)
	return &TranslateLocation{
		locationsBase.LocationBase{
			TargetDirectoryBase: locationsBase.TargetDirectoryBase{
				TargetDirectory:     targetDirectory,
				TargetDirectoryName: translateDirectoryName,
			},
			TargetFileBase: locationsBase.TargetFileBase{},
		},
	}
}

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.IsTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
