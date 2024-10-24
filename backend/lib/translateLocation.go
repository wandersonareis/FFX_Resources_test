package lib

type TranslateLocation struct {
	LocationBase
}

var translateLocationInstance *TranslateLocation

func NewTranslateLocation() *TranslateLocation {
	rootDirectoryName := "translated"

	if translateLocationInstance == nil {
		translateLocationInstance = &TranslateLocation{
			LocationBase: NewLocationBase(rootDirectoryName),
		}
	}

	return translateLocationInstance
}

func (t *TranslateLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().TranslateLocation.TargetDirectory != "" {
		return NewInteraction().TranslateLocation.TargetDirectory, nil
	}

	return t.LocationBase.ProvideTargetDirectory()
}
