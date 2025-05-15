package textverify

import (
	"bytes"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
)

type (
	textContentComparerStrategy struct{}

	TextFileDetails struct {
		Filename              string
		HeaderElements        int
		HeaderElementsNonZero int
	}
)

func NewTextContentComparerStrategy() ITextVerificationStrategy {
	return &textContentComparerStrategy{}
}

func (pc *textContentComparerStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	fromFile := destination.Translate().GetTargetFile()
	toFile := destination.Extract().GetTargetFile()
	
	if err := pc.compareFileData(fromFile, toFile); err != nil {
		return err
	}

	return nil
}

func (pc *textContentComparerStrategy) compareFileData(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		return fmt.Errorf("error: %w when reading file: %s", err, fromFile)
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		return fmt.Errorf("error: %w when reading file: %s", err, toFile)
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		return fmt.Errorf("files are different: %s and %s", fromFile, toFile)
	}

	return nil
}
