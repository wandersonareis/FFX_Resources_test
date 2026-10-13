package interactions

import (
	"fmt"
	"path/filepath"
)

type interactionBase struct {
	targetDir string
	config    ISetAppConfig
	configKey string
}

func (e *interactionBase) GetTargetDirectory() string {
	return e.targetDir
}

func (e *interactionBase) SetTargetDirectory(path string) error {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error when obtaining the absolute path: %v", err)
	}

	e.targetDir = fullPath

	if e.config != nil {
		e.config.SetLocation(e.configKey, fullPath)
	}

	return nil
}
