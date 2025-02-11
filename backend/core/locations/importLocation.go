package locations

import (
	"ffxresources/backend/core/locations/locationsBase"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IImportLocation interface {
	locationsBase.ILocationBase
	interfaces.IValidate
}

type ITargetImportLocation interface {
	GetImportLocation() IImportLocation
}

type ImportLocation struct {
	locationsBase.LocationBase
}

func NewImportLocation(options *locationsBase.LocationBaseOptions) *ImportLocation {
	return &ImportLocation{
		LocationBase: locationsBase.NewLocationBase(options),
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
