package interactions

import "ffxresources/backend/interfaces"

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
