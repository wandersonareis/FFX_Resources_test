package interactions

import (
	"ffxresources/backend/common"
	"fmt"
	"path/filepath"
)

type ITextFormatter interface {
	ReadFile(fileInfo *GameDataInfo, targetDirectory string) (string, string)
	WriteFile(fileInfo *GameDataInfo, targetDirectory string) (string, string)
}

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
		return lb.TargetDirectory, providerTargetDirectory(lb.TargetDirectory)
	}

	path := filepath.Join(common.GetExecDir(), lb.TargetDirectoryName)

	err := providerTargetDirectory(path)
	if err != nil {
		return "", err
	}

	lb.TargetDirectory = path

	return path, nil
}

func (lb *LocationBase) ProvideTargetPath() error {
	if lb.TargetPath != "" {
		return providerTargetDirectory(lb.TargetPath)
	}

	return fmt.Errorf("target path is empty")
}

func (t *LocationBase) GenerateTargetOutput(formatter ITextFormatter, fileInfo *GameDataInfo) {
	t.TargetFile, t.TargetPath = formatter.ReadFile(fileInfo, t.TargetDirectory)

	t.TargetFileName = filepath.Base(t.TargetFile)
}

func (t *LocationBase) targetFileExists() bool {
	t.IsExist = common.IsFileExists(t.TargetFile)
	return t.IsExist
}

func providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
