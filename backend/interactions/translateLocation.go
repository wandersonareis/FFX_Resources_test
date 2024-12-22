package interactions

import "ffxresources/backend/interfaces"

type (
	TranslateLocation struct {
		*interactionBase
	}
	ITranslateLocation interface {
		interfaces.IInteractionBase
	}
)

func newTranslateLocation(ffxAppConfig IFFXAppConfig) ITranslateLocation {
	rootDirectoryName := "translated"

	return &TranslateLocation{
		interactionBase: &interactionBase{
			ffxAppConfig:      ffxAppConfig,
			defaultDirName: rootDirectoryName,
		},
	}
}

func (t *TranslateLocation) GetTargetDirectory() string {
	path, _ := t.interactionBase.GetTargetDirectoryBase(ConfigTranslateLocation)
	return path.(string)
}

func (t *TranslateLocation) SetTargetDirectory(path string) {
	t.interactionBase.SetTargetDirectoryBase(ConfigTranslateLocation, path)
}

func (e *TranslateLocation) ProvideTargetDirectory() error {
	path := e.GetTargetDirectory()

	err := e.interactionBase.ProviderTargetDirectoryBase(ConfigTranslateLocation, path)
	if err != nil {
		return err
	}

	return nil
}
