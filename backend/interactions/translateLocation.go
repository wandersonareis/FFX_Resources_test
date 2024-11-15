package interactions

import "fmt"

type ITranslateLocation interface {
	ILocationBase
	IValidate
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
	if !t.isTargetFileAvailable() {
		return fmt.Errorf("translated file does not exist")
	}

	return nil
}
