package interactions

import "fmt"

type IImportLocation interface {
	ILocationBase
	IValidate
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

func (i *ImportLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *GameDataInfo) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(fileInfo, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	i.IsExist = i.isTargetFileAvailable()

	if !i.IsExist {
		return fmt.Errorf("reimport file not exists: %s", i.TargetFile)
	}

	return nil
}
