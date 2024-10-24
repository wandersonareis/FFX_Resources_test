package lib

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

func (i *ImportLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ImportLocation.TargetDirectory != "" {
		return NewInteraction().ImportLocation.TargetDirectory, nil
	}

	return i.LocationBase.ProvideTargetDirectory()
}

func (i *ImportLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *FileInfo) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(fileInfo, i.TargetDirectory)
}
