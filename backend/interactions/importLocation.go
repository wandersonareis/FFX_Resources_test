package interactions

import "os"

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

/* func (i *ImportLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ImportLocation.GetTargetDirectory() != "" {
		return NewInteraction().ImportLocation.GetTargetDirectory(), nil
	}

	return i.LocationBase.ProvideTargetDirectory()
} */

func (i *ImportLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *GameDataInfo) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(fileInfo, i.TargetDirectory)
}

func (i *ImportLocation) Validate() error {
	if i.isTargetFileAvailable() {
		return os.Remove(i.TargetFile)
	}

	return nil
}
