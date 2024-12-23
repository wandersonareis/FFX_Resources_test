package interactions

import "ffxresources/backend/interfaces"

type (
	ExtractLocation struct {
		*interactionBase
	}
	IExtractLocation interface {
		interfaces.IInteractionBase
	}
)

func newExtractLocation() IExtractLocation {
	rootDirectoryName := "extracted"

	return &ExtractLocation{
		interactionBase: &interactionBase{
			defaultDirName: rootDirectoryName,
		},
	}
}

func (e *ExtractLocation) GetTargetDirectory() string {
	path, _ := e.interactionBase.GetTargetDirectoryBase(ConfigExtractLocation)
	return path.(string)
}

func (e *ExtractLocation) SetTargetDirectory(path string) {
	e.interactionBase.SetTargetDirectoryBase(ConfigExtractLocation, path)
}

func (e *ExtractLocation) ProvideTargetDirectory() error {
	path := e.GetTargetDirectory()

	err := e.interactionBase.ProviderTargetDirectoryBase(ConfigExtractLocation, path)
	if err != nil {
		return err
	}

	return nil
}
