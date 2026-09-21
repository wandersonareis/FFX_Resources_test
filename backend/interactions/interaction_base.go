package interactions

import (
	"ffxresources/backend/common"
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

func (e *interactionBase) ProvideTargetDirectory() error {
	if e.targetDir == "" {
		e.targetDir = defaultDirForKey(e.configKey)
	}

	err := common.EnsurePathExists(e.targetDir)
	if err != nil {
		return err
	}
	return nil
}

// defaultDirForKey devolve o default por chave de config.
// GameFiles = <exec>/data; Translate deriva dele (<game>/mods/translated);
// Extract/Import são internos ocultos do frontend.
func defaultDirForKey(configKey string) string {
	execDir := common.GetExecDir()
	switch configKey {
	case "GameFilesLocation":
		return filepath.Join(execDir, common.DirData)
	case "TranslateLocation":
		return common.DefaultTranslatedDir(filepath.Join(execDir, common.DirData))
	case "ExtractLocation":
		return filepath.Join(execDir, common.DirExtracted)
	case "ImportLocation":
		return filepath.Join(execDir, common.DirReimported)
	default:
		return filepath.Join(common.ResourcesRoot, common.DirData)
	}
}
