package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
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

func NewExtractLocation(extractDirectoryName, extractTargetDirectory, gameVersionDir string) IExtractLocation {
	targetDirectory := filepath.Join(extractTargetDirectory, gameVersionDir)
	return &ExtractLocation{
		locationsBase.LocationBase{
			TargetDirectoryBase: locationsBase.TargetDirectoryBase{
				TargetDirectory:     targetDirectory,
				TargetDirectoryName: extractDirectoryName,
			},
			TargetFileBase: locationsBase.TargetFileBase{},
		},
	}
}

func (e *ExtractLocation) Validate() error {
	e.IsExist = e.IsTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist: %s", e.GetTargetFile())
	}

	return nil
}
