package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/logger"
	"fmt"
	"os"
	"path/filepath"
)

type IValidate interface {
	Validate() error
}

type ILocationBase interface {
	SetTargetDirectory(path string)
	GetTargetDirectory() string
	SetTargetFile(targetFile string)
	SetTargetPath(targetPath string)
	ProvideTargetDirectory() error
	ProvideTargetPath() error
	BuildTargetOutput(formatter ITextFormatter, fileInfo *GameDataInfo)
}

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

func (lb *LocationBase) SetTargetDirectory(path string) {
	if path == "" {
		return
	}

	lb.TargetDirectory = path
}

func (lb *LocationBase) GetTargetDirectory() string {
	return lb.TargetDirectory
}

func (lb *LocationBase) SetTargetFile(targetFile string) {
	lb.TargetFile = targetFile
	lb.IsExist = lb.isTargetFileAvailable()
}

func (lb *LocationBase) SetTargetPath(targetPath string) {
	lb.TargetPath = targetPath
}

func (lb *LocationBase) ProvideTargetDirectory() error {
	path := filepath.Join(common.GetExecDir(), lb.TargetDirectoryName)

	err := lb.providerTargetDirectory(path)
	if err != nil {
		return err
	}

	lb.TargetDirectory = path

	return nil
}

func (lb *LocationBase) ProvideTargetPath() error {
	if lb.TargetPath != "" {
		return lb.providerTargetDirectory(lb.TargetPath)
	}

	return fmt.Errorf("target path is empty")
}

func (lb *LocationBase) BuildTargetOutput(formatter ITextFormatter, fileInfo *GameDataInfo) {
	lb.TargetFile, lb.TargetPath = formatter.ReadFile(fileInfo, lb.TargetDirectory)

	lb.TargetFileName = filepath.Base(lb.TargetFile)
}

func (lb *LocationBase) DisposeTargetFile() {
	if common.IsFileExists(lb.TargetFile) {
		err := os.Remove(lb.TargetFile)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing file: %s", err)
		}
	}
}

func (lb *LocationBase) DisposeTargetPath() {
	if common.IsPathExists(lb.TargetPath) {
		err := os.RemoveAll(lb.TargetPath)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing path: %s", err)
			return
		}
	}
}

func (t *LocationBase) isTargetFileAvailable() bool {
	t.IsExist = common.IsFileExists(t.TargetFile)
	return t.IsExist
}

func (t *LocationBase) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
