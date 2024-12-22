package interactions

type (
	ExtractLocation struct {
		*interactionBase
	}
	IExtractLocation interface {
		GetTargetDirectory() string
		SetTargetDirectory(path string)
		ProvideTargetDirectory() error
	}
)

func newExtractLocation(ffxAppConfig IFFXAppConfig) IExtractLocation {
	rootDirectoryName := "extracted"

	return &ExtractLocation{
		interactionBase: &interactionBase{
			ffxAppConfig:   ffxAppConfig,
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

/* func (e *ExtractLocation) providerTargetDirectory(targetDirectory string) error {
	if targetDirectory == "" {
		targetDirectory = filepath.Join(common.GetExecDir(), e.defaultDirName)

		e.SetTargetDirectory(targetDirectory)
	}

	err := common.EnsurePathExists(targetDirectory)
	if err != nil {
		return err
	}
	return nil
} */
