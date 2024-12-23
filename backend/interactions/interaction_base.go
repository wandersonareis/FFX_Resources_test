package interactions

import (
	"ffxresources/backend/common"
	"path/filepath"
)

type interactionBase struct {
	defaultDirName string
}

func (e *interactionBase) GetTargetDirectoryBase(field ConfigField) (interface{}, error) {
	return NewInteraction().FFXAppConfig().GetField(field)
}

func (e *interactionBase) SetTargetDirectoryBase(field ConfigField, path string) {
	NewInteraction().FFXAppConfig().UpdateField(field, path)
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
