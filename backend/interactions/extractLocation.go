package interactions

type (
	ExtractLocation struct {
		InteractionBase
	}
	IExtractLocation interface {
		GetTargetDirectory() string
		SetTargetDirectory(path string)
	}
)

var extractLocationInstance *ExtractLocation

func NewExtractLocation() *ExtractLocation {
	rootDirectoryName := "extracted"

	if extractLocationInstance == nil {
		extractLocationInstance = &ExtractLocation{
			InteractionBase: newInteractionBase(rootDirectoryName),
		}
	}

	return extractLocationInstance
}
