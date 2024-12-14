package interactions

import (
	"ffxresources/backend/interfaces"
	"fmt"
)

type IImportLocation interface {
	interfaces.ILocationBase
	interfaces.IValidate
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

func (i *ImportLocation) GenerateTargetOutput(formatter interfaces.ITextFormatterDev, fileInfo interfaces.ISource) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(fileInfo, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.isTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
