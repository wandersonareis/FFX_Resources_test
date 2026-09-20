package interactions

import "ffxresources/backend/interfaces"

type (
	ImportLocation struct {
		*interactionBase
	}
	IImportLocation interface {
		interfaces.IInteractionBase

		WithTargetDirectory(path string) IImportLocation
	}
)

func newImportLocation(path string, config ISetAppConfig) IImportLocation {
	importLocation := &ImportLocation{
		interactionBase: &interactionBase{
			targetDir: path,
			config:    config,
			configKey: "ImportLocation",
		},
	}
	importLocation.SetTargetDirectory(path)
	return importLocation
}

func (i *ImportLocation) WithTargetDirectory(path string) IImportLocation {
	_ = i.SetTargetDirectory(path)
	return i
}
