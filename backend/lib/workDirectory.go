package lib

import "path/filepath"

type WorkDirectory struct {
	WorkDirectoryName   string
	WorkDirectory       string
	ExtractedDirectory  string
	TranslatedDirectory string
}

const (
	workDirectoryName       = "bin"
	extractedDirectoryName  = "extracted"
	translatedDirectoryName = "translated"
)

var workDirectory *WorkDirectory

func NewWorkDirectory() *WorkDirectory {
	rootDirectory = getTempDir()

	if workDirectory == nil {
		workDirectory = &WorkDirectory{
			WorkDirectoryName: workDirectoryName,
			WorkDirectory:     filepath.Join(rootDirectory, workDirectoryName),
		}
	}

	return workDirectory
}

func (w WorkDirectory) ProvideExtractedDirectory() (string, error) {
	if NewInteraction().WorkingLocation.ExtractedDirectory != "" {
		return NewInteraction().WorkingLocation.ExtractedDirectory, nil
	}

	path := filepath.Join(rootDirectory, extractedDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (w WorkDirectory) ProvideTranslatedDirectory() (string, error) {
	if NewInteraction().WorkingLocation.TranslatedDirectory != "" {
		return NewInteraction().WorkingLocation.TranslatedDirectory, nil
	}

	path := filepath.Join(rootDirectory, translatedDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func GetExtractDirectory() string {
	return filepath.Join(rootDirectory, extractedDirectoryName)
}

func GetTranslateDirectory() string {
	return filepath.Join(rootDirectory, translatedDirectoryName)
}
