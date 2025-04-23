package interactions

import (
	"ffxresources/backend/common"
	"fmt"
	"path/filepath"
)

type interactionBase struct {
	defaultDirName string
}

func (e *interactionBase) GetTargetDirectoryBase(field ConfigField) (interface{}, error) {
	return NewInteractionService().FFXAppConfig().GetField(field)
}

func (e *interactionBase) SetTargetDirectoryBase(field ConfigField, path string) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error when obtaining the absolute path: %v", err)
	}

	if err := NewInteractionService().FFXAppConfig().UpdateField(field, fullPath); err != nil {
		return err
	}

	return nil
}

func (e *interactionBase) ProviderTargetDirectoryBase(field ConfigField, targetDirectory string) error {
	if targetDirectory == "" {
		targetDirectory = filepath.Join(common.GetExecDir(), e.defaultDirName)

		if err := e.SetTargetDirectoryBase(field, targetDirectory); err != nil {
			return err
		}
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
}
