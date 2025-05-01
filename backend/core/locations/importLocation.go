package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
	"path/filepath"
)

type (
	IImportLocation interface {
		locationsBase.ILocationBase
		interfaces.IValidate
	}

	ITargetImportLocation interface {
		GetImportLocation() IImportLocation
	}

	ImportLocation struct {
		locationsBase.LocationBase
	}
)

func NewImportLocation(importDirectoryName, importTargetDirectory, gameVersionDir string) IImportLocation {
	targetDirectory := filepath.Join(importTargetDirectory, gameVersionDir)
	return &ImportLocation{
		locationsBase.LocationBase{
			TargetDirectoryBase: locationsBase.TargetDirectoryBase{
				TargetDirectory:     targetDirectory,
				TargetDirectoryName: importDirectoryName,
			},
			TargetFileBase: locationsBase.TargetFileBase{},
		},
	}
}

func (i *ImportLocation) GenerateTargetOutput(formatter interfaces.ITextFormatter, source interfaces.ISource) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(source, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.IsTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
