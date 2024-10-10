package lib

import "path/filepath"

type ExtractLocation struct {
	rootDirectory       string
	rootDirectoryName   string
	targetFileExtension string
	TargetDirectoryName string
	TargetDirectory     string
	TargetFile          string
	TargetPath          string
	IsExist             bool
}

var extractLocationInstance *ExtractLocation

func NewExtractLocation() *ExtractLocation {
	const (
		rootDirectoryName   = "extracted"
		targetFileExtension = ".txt"
	)

	targetDirectory := filepath.Join(GetExecDir(), rootDirectoryName)

	if extractLocationInstance == nil {
		extractLocationInstance = &ExtractLocation{
			rootDirectory:       targetDirectory,
			targetFileExtension: targetFileExtension,
			TargetDirectory:     targetDirectory,
		}
	}

	return extractLocationInstance
}

func (e ExtractLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ExtractLocation.TargetDirectory != "" {
		return NewInteraction().ExtractLocation.TargetDirectory, nil
	}

	path := filepath.Join(e.rootDirectory, e.rootDirectoryName)
	err := EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (e *ExtractLocation) GenerateTargetOutput(formatter ITextFormatter, fileInfo FileInfo) {
	extractedFile, extractedPath := formatter.Write(fileInfo, e.TargetDirectory)

	e.TargetFile = extractedFile
	e.TargetPath = extractedPath
}

func (e ExtractLocation) TargetFileExists() bool {
	e.IsExist = FileExists(e.TargetFile)
	return e.IsExist
}