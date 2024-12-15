package interactions

import "ffxresources/backend/interfaces"

type (
	ImportLocation struct {
		InteractionBase
	}
	IImportLocation interface {
		interfaces.IInteractionBase
	}
)

var importLocationInstance *ImportLocation

func NewImportLocation() *ImportLocation {
	rootDirectoryName := "reimported"

	if importLocationInstance == nil {
		importLocationInstance = &ImportLocation{
			InteractionBase: newInteractionBase(rootDirectoryName),
		}
	}

	return importLocationInstance
}
