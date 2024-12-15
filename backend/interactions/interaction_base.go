package interactions

import (
	"ffxresources/backend/common"
	"path/filepath"
)

type InteractionBase struct {
	TargetDirectory     string
	TargetDirectoryName string
}

func newInteractionBase(targetDirectoryName string) InteractionBase {
	targetDirectory := filepath.Join(common.GetExecDir(), targetDirectoryName)

	return InteractionBase{
		TargetDirectoryName: targetDirectoryName,
		TargetDirectory:     targetDirectory,
	}
}

func (lb *InteractionBase) SetTargetDirectory(path string) {
	if path == "" {
		return
	}

	lb.TargetDirectory = path
}

func (lb *InteractionBase) GetTargetDirectory() string {
	return lb.TargetDirectory
}

func (lb *InteractionBase) ProvideTargetDirectory() error {
	path := filepath.Join(common.GetExecDir(), lb.TargetDirectoryName)

	err := lb.providerTargetDirectory(path)
	if err != nil {
		return err
	}

	lb.TargetDirectory = path

	return nil
}

func (t *InteractionBase) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory != "" && common.IsPathExists(targetDirectory) {
		return nil
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
