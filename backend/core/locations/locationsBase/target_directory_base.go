package locationsBase

import "ffxresources/backend/common"

type (
	TargetDirectoryBase struct {
		TargetDirectory     string
		TargetDirectoryName string
	}

	ITargetDirectoryBase interface {
		GetTargetDirectory() string
		SetTargetDirectory(path string)
		GetTargetDirectoryName() string
		SetTargetDirectoryName(name string)
		ProvideTargetDirectory() error
	}
)

func (lb *TargetDirectoryBase) SetTargetDirectory(path string) {
	if path == "" {
		return
	}

	lb.TargetDirectory = path
}

func (lb *TargetDirectoryBase) GetTargetDirectory() string {
	return lb.TargetDirectory
}

func (lb *TargetDirectoryBase) SetTargetDirectoryName(name string) {
	if name == "" {
		return
	}

	lb.TargetDirectoryName = name
}

func (lb *TargetDirectoryBase) GetTargetDirectoryName() string {
	return lb.TargetDirectoryName
}

func (lb *TargetDirectoryBase) ProvideTargetDirectory() error {
	return lb.providerTargetDirectory(lb.GetTargetDirectory())
}

func (t *TargetDirectoryBase) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
