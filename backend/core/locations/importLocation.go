package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	IImportLocation interface {
		locationsBase.ILocationBase
		interfaces.IValidate
		Copy() ImportLocation
	}

	ITargetImportLocation interface {
		GetImportLocation() IImportLocation
	}

	ImportLocation struct {
		locationsBase.LocationBase
	}
)

func NewImportLocation(importDirectoryName, importTargetDirectoryPath, gameVersionDir string) IImportLocation {
	return &ImportLocation{
		LocationBase: locationsBase.NewLocationBase(importDirectoryName, importTargetDirectoryPath, gameVersionDir),
	}
}

func (i *ImportLocation) Copy() ImportLocation {
	return *i
}

func (i *ImportLocation) GenerateTargetOutput(formatter interfaces.ITextFormatter, source interfaces.ISource) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(source, i.GetTargetDirectory())
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.IsTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
