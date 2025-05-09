package locationsBase

import "ffxresources/backend/common"

type (
	TargetDirectoryBase struct {
		targetDirectory     string
		targetDirectoryName string
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

	lb.targetDirectory = path
}

func (lb *TargetDirectoryBase) GetTargetDirectory() string {
	return lb.targetDirectory
}

func (lb *TargetDirectoryBase) SetTargetDirectoryName(name string) {
	if name == "" {
		return
	}

	lb.targetDirectoryName = name
}

func (lb *TargetDirectoryBase) GetTargetDirectoryName() string {
	return lb.targetDirectoryName
}

func (lb *TargetDirectoryBase) ProvideTargetDirectory() error {
	return lb.providerTargetDirectory(lb.GetTargetDirectory())
}

func (t *TargetDirectoryBase) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	if err := common.EnsurePathExists(targetDirectory); err != nil {
		return err
	}
	return nil
}
