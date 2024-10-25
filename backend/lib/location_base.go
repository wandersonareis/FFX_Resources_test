package lib

import (
	"ffxresources/backend/common"
	"path/filepath"
)

type LocationBase struct {
	IsExist             bool
	TargetFile          string
	TargetPath          string
	TargetFileName      string
	TargetDirectory     string
	TargetDirectoryName string
}

func NewLocationBase(targetDirectoryName string) LocationBase {
	targetDirectory := filepath.Join(common.GetExecDir(), targetDirectoryName)

	return LocationBase{
		TargetDirectoryName: targetDirectoryName,
		TargetDirectory:     targetDirectory,
	}
}

func (lb *LocationBase) SetPath(path string) {
	if path == "" {
		return
	}

	lb.TargetDirectory = path
}

func (lb *LocationBase) GetPath() string {
	return lb.TargetDirectory
}

func (lb *LocationBase) ProvideTargetDirectory() (string, error) {
	if lb.TargetDirectory != "" {
		return lb.TargetDirectory, nil
	}

	path := filepath.Join(common.GetExecDir(), lb.TargetDirectoryName)
	err := common.EnsurePathExists(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (t *LocationBase) GenerateTargetOutput(formatter ITextFormatter, fileInfo *FileInfo) {
	t.TargetFile, t.TargetPath = formatter.ReadFile(fileInfo, t.TargetDirectory)

	t.TargetFileName = filepath.Base(t.TargetFile)
}

func (t *LocationBase) TargetFileExists() bool {
	t.IsExist = common.FileExists(t.TargetFile)
	return t.IsExist
}
