package lib

import "path/filepath"

type TranslateLocation struct {
	LocationBase
}

var translateLocationInstance *TranslateLocation

func NewTranslateLocation() *TranslateLocation {
	rootDirectoryName = "translated"

	targetDirectory := filepath.Join(GetExecDir(), rootDirectoryName)

	if translateLocationInstance == nil {
		translateLocationInstance = &TranslateLocation{
			LocationBase: LocationBase{
				TargetDirectoryName: rootDirectoryName,
				TargetDirectory:     targetDirectory,
			},
		}
	}

	return translateLocationInstance
}

/* func (t *TranslateLocation) SetPath(path string) {
	if path == "" {
		return
	}

	t.TargetDirectory = path
} */

func (t *TranslateLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().TranslateLocation.TargetDirectory != "" {
		return NewInteraction().TranslateLocation.TargetDirectory, nil
	}

	path := filepath.Join(rootDirectory, rootDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (t *TranslateLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo *FileInfo) {
	t.TargetFile, t.TargetPath = formatter.ReadFile(fileInfo, t.TargetDirectory)

	t.TargetFileName = filepath.Base(t.TargetFile)
}

func (t *TranslateLocation) TargetFileExists() bool {
	t.IsExist = FileExists(t.TargetFile)
	return t.IsExist
}
