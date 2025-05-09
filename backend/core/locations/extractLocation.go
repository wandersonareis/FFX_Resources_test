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
		Copy() ExtractLocation
	}
)

func NewExtractLocation(extractDirectoryName, extractTargetDirectory, gameVersionDir string) IExtractLocation {
	return &ExtractLocation{
		LocationBase: locationsBase.NewLocationBase(extractDirectoryName, extractTargetDirectory, gameVersionDir),
	}
}

func (e *ExtractLocation) Copy() ExtractLocation {
	return *e
}

func (e *ExtractLocation) Validate() error {
	e.IsExist = e.IsTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist: %s", e.GetTargetFile())
	}

	return nil
}
