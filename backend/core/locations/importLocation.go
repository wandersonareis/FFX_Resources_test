package locations

import (
	"ffxresources/backend/bases"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IImportLocation interface {
	interfaces.ILocationBase
	interfaces.IValidate
}

type ITargetImportLocation interface {
	GetImportLocation() IImportLocation
}

type ImportLocation struct {
	LocationBase
}

func NewImportLocation(options *bases.LocationBaseOptions) *ImportLocation {
	return &ImportLocation{
		LocationBase: NewLocationBase(options),
	}
}

func (i *ImportLocation) GenerateTargetOutput(formatter interfaces.ITextFormatterDev, source interfaces.ISource) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(source, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.IsTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
