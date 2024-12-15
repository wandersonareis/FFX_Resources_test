package interactions

import "ffxresources/backend/interfaces"

type (
	TranslateLocation struct {
		InteractionBase
	}
	ITranslateLocation interface {
		interfaces.IInteractionBase
	}
)

var translateLocationInstance *TranslateLocation

func NewTranslateLocation() *TranslateLocation {
	rootDirectoryName := "translated"

	if translateLocationInstance == nil {
		translateLocationInstance = &TranslateLocation{
			InteractionBase: newInteractionBase(rootDirectoryName),
		}
	}

	return translateLocationInstance
}
