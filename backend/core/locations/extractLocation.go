package locations

import (
	"ffxresources/backend/interfaces"
	"fmt"
)

type (
	ExtractLocation struct {
		LocationBase
	}

	IExtractLocation interface {
		interfaces.ILocationBase
		interfaces.IValidate
	}
)

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
	e.IsExist = e.isTargetFileAvailable()

	if !e.IsExist {
		return fmt.Errorf("extracted file does not exist")
	}

	return nil
}
