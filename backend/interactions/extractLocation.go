package interactions

type ExtractLocation struct {
	LocationBase
}

var extractLocationInstance *ExtractLocation

func NewExtractLocation() *ExtractLocation {
	rootDirectoryName := "extracted"

	if extractLocationInstance == nil {
		extractLocationInstance = &ExtractLocation{
			LocationBase: NewLocationBase(rootDirectoryName),
		}
	}

	return extractLocationInstance
}

func (e ExtractLocation) ProvideTargetDirectory() (string, error) {
	if NewInteraction().ExtractLocation.TargetDirectory != "" {
		return NewInteraction().ExtractLocation.TargetDirectory, nil
	}

	return e.LocationBase.ProvideTargetDirectory()
}
