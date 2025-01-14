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

func (e *interactionBase) SetTargetDirectoryBase(field ConfigField, path string) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Errorf("erro ao obter o caminho absoluto: %v", err))
	}

	if err := NewInteractionService().FFXAppConfig().UpdateField(field, fullPath); err != nil {
		return
	}
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
