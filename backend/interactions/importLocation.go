package interactions

import "ffxresources/backend/interfaces"

type (
	ImportLocation struct {
		*interactionBase
	}
	IImportLocation interface {
		interfaces.IInteractionBase
	}
)

func newImportLocation() IImportLocation {
	rootDirectoryName := "reimported"

	return &ImportLocation{
		interactionBase: &interactionBase{
			defaultDirName: rootDirectoryName,
		},
	}
}

func (i *ImportLocation) GetTargetDirectory() string {
	path, _ := i.interactionBase.GetTargetDirectoryBase(ConfigImportLocation)
	return path.(string)
}

func (i *ImportLocation) SetTargetDirectory(path string) {
	i.interactionBase.SetTargetDirectoryBase(ConfigImportLocation, path)
}

func (i *ImportLocation) ProvideTargetDirectory() error {
	path := i.GetTargetDirectory()

	err := i.interactionBase.ProviderTargetDirectoryBase(ConfigImportLocation, path)
	if err != nil {
		return err
	}

	return nil
}
