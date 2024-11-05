package interactions

import "fmt"

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
	if !e.targetFileExists() {
		return fmt.Errorf("extracted file does not exist")
	}

	return nil
}
