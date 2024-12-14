package locations

import (
	"ffxresources/backend/interfaces"
	"fmt"
)

/* type IImportLocation interface {
	ILocationBase
	IValidate
} */

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

var importLocationInstance *ImportLocation

func NewImportLocation() *ImportLocation {
	rootDirectoryName := "reimported"

	if importLocationInstance == nil {
		importLocationInstance = &ImportLocation{
			LocationBase: NewLocationBase(rootDirectoryName),
		}
	}

	return importLocationInstance
}

/* func (i *ImportLocation) Get() *ImportLocation {
	return i
} */

func (i *ImportLocation) GenerateTargetOutput(formatter interfaces.ITextFormatterDev, source interfaces.ISource) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(source, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.isTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
