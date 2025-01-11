package locations

import (
	internal "ffxresources/backend/core/locations/base"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IImportLocation interface {
	internal.ILocationBase
	interfaces.IValidate
}

type ITargetImportLocation interface {
	GetImportLocation() IImportLocation
}

type ImportLocation struct {
	internal.LocationBase
}

func NewImportLocation(options *internal.LocationBaseOptions) *ImportLocation {
	return &ImportLocation{
		LocationBase: internal.NewLocationBase(options),
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
