package interactions

import "ffxresources/backend/interfaces"

type (
	ExtractLocation struct {
		*interactionBase
	}
	IExtractLocation interface {
		interfaces.IInteractionBase

		WithTargetDirectory(path string) IExtractLocation
	}
)

func newExtractLocation(path string, config ISetAppConfig) IExtractLocation {
	extractLocation := &ExtractLocation{
		interactionBase: &interactionBase{
			targetDir: path,
			config:    config,
			configKey: "ExtractLocation",
		},
	}
	extractLocation.SetTargetDirectory(path)
	return extractLocation
}

func (e *ExtractLocation) WithTargetDirectory(path string) IExtractLocation {
	_ = e.SetTargetDirectory(path)
	return e
}
