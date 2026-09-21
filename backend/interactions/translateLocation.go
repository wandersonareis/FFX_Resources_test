package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
)

type (
	TranslateLocation struct {
		*interactionBase
	}
	ITranslateLocation interface {
		interfaces.IInteractionBase

		WithTargetDirectory(path string) ITranslateLocation
	}
)

func newTranslateLocation(path string, config ISetAppConfig) ITranslateLocation {
	translateLocation := &TranslateLocation{
		interactionBase: &interactionBase{
			targetDir: path,
			config:    config,
			configKey: "TranslateLocation",
		},
	}
	translateLocation.SetTargetDirectory(path)
	return translateLocation
}

func (t *TranslateLocation) WithTargetDirectory(path string) ITranslateLocation {
	_ = t.SetTargetDirectory(path)
	return t
}

// SetTargetDirectory persiste imediatamente no config.json (via base)
// e garante a existência do diretório traduzido (<game>/mods/translated
// por padrão). É a fonte da tradução usada no reimport (Compress).
func (t *TranslateLocation) SetTargetDirectory(path string) error {
	if err := t.interactionBase.SetTargetDirectory(path); err != nil {
		return err
	}
	if target := t.interactionBase.GetTargetDirectory(); target != "" {
		_ = common.EnsurePathExists(target)
	}
	return nil
}
