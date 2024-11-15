package interactions

import "fmt"

type IExtractLocation interface {
	ILocationBase
	IValidate
}

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

func (e *ExtractLocation) Validate() error {
	if !e.isTargetFileAvailable() {
		return fmt.Errorf("extracted file does not exist")
	}

	return nil
}
