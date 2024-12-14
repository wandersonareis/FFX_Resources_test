package interactions

import (
	"ffxresources/backend/interfaces"
	"fmt"
)

type ITranslateLocation interface {
	interfaces.ILocationBase
	interfaces.IValidate
}

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

func (t *TranslateLocation) Validate() error {
	t.IsExist = t.isTargetFileAvailable()

	if !t.IsExist {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
