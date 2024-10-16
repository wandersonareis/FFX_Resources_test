package lib

type ImportLocation struct {
	LocationBase
}

var importLocationInstance *ImportLocation

func NewImportLocation() *ImportLocation {
	rootDirectoryName = "reimported"

	targetDirectory := PathJoin(GetExecDir(), rootDirectoryName)

	if importLocationInstance == nil {
		importLocationInstance = &ImportLocation{
			LocationBase: LocationBase{
				TargetDirectory:     targetDirectory,
				TargetDirectoryName: rootDirectoryName,
			},
		}
	}

	return importLocationInstance
}

func (i *ImportLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ImportLocation.TargetDirectory != "" {
		return NewInteraction().ImportLocation.TargetDirectory, nil
	}

	path := PathJoin(rootDirectory, rootDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (i *ImportLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *FileInfo) {
	i.TargetFile, i.TargetPath = formatter.WriteFile(fileInfo, i.TargetDirectory)
}
