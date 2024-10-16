package lib

type ExtractLocation struct {
	LocationBase
}

var extractLocationInstance *ExtractLocation

func NewExtractLocation() *ExtractLocation {
	rootDirectoryName = "extracted"

	targetDirectory := PathJoin(GetExecDir(), rootDirectoryName)

	if extractLocationInstance == nil {
		extractLocationInstance = &ExtractLocation{
			LocationBase: LocationBase{
				TargetDirectory:     targetDirectory,
				TargetDirectoryName: rootDirectoryName,
			},
		}
	}

	return extractLocationInstance
}

func (e *ExtractLocation) SetTargetDirectory(path string) {
	if path == "" {
		return
	}

	e.TargetDirectory = path
}

func (e ExtractLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ExtractLocation.TargetDirectory != "" {
		return NewInteraction().ExtractLocation.TargetDirectory, nil
	}

	path := PathJoin(rootDirectory, rootDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (e *ExtractLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *FileInfo) {
	e.TargetFile, e.TargetPath = formatter.ReadFile(fileInfo, e.TargetDirectory)
}

func (e ExtractLocation) TargetFileExists() bool {
	e.IsExist = FileExists(e.TargetFile)
	return e.IsExist
}
