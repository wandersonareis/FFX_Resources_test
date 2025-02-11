package locationsBase

import (
	"ffxresources/backend/common"
	"fmt"
	"path/filepath"
)

type (
	TargetFileBase struct {
		IsExist         bool
		TargetFile      string
		TargetPath      string
		TargetFileName  string
	}

	ITargetFileBase interface {
		GetTargetFile() string
		SetTargetFile(targetFile string)
		GetTargetPath() string
		SetTargetPath(targetPath string)
		GetTargetExtension() string
		ProvideTargetPath() error
	}
)

func (lb *TargetFileBase) GetTargetFile() string {
	return lb.TargetFile
}

func (lb *TargetFileBase) SetTargetFile(targetFile string) {
	lb.TargetFile = targetFile
	lb.IsExist = lb.IsTargetFileAvailable()
}

func (lb *TargetFileBase) GetTargetPath() string {
	return lb.TargetPath
}

func (lb *TargetFileBase) SetTargetPath(targetPath string) {
	lb.TargetPath = targetPath
}

func (lb *TargetFileBase) GetTargetExtension() string {
	return filepath.Ext(lb.TargetFile)
}

func (lb *TargetFileBase) ProvideTargetPath() error {
	if lb.TargetPath != "" {
		return lb.providerTargetDirectory(lb.TargetPath)
	}

	return fmt.Errorf("target path is empty")
}

func (t *TargetFileBase) IsTargetFileAvailable() bool {
	t.IsExist = common.IsFileExists(t.TargetFile)
	return t.IsExist
}


func (t *TargetFileBase) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
