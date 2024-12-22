package interactions

import (
	"ffxresources/backend/common"
	"path/filepath"
)

type interactionBase struct {
	ffxAppConfig   IFFXAppConfig
	defaultDirName string
}

func (e *interactionBase) GetTargetDirectoryBase(field ConfigField) (interface{}, error) {
	//v := NewInteraction().FFXGameVersion().GetGameVersionNumber()
	return NewInteraction().FFXAppConfig().GetField(field)
}

func (e *interactionBase) SetTargetDirectoryBase(field ConfigField, path string) {
	e.ffxAppConfig.UpdateField(field, path)
}

func (e *interactionBase) ProviderTargetDirectoryBase(field ConfigField, targetDirectory string) error {
	if targetDirectory == "" {
		targetDirectory = filepath.Join(common.GetExecDir(), e.defaultDirName)

		e.SetTargetDirectoryBase(field, targetDirectory)
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}

/* import (
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
*/
