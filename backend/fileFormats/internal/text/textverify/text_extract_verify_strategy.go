package textverify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type textExtractionVerificationStrategy struct {
	FileSegmentCounter ISegmentCounter
}

func NewTextExtractionVerificationStrategy() ITextVerificationStrategy {
	return &textExtractionVerificationStrategy{
		FileSegmentCounter: newSegmentCounter(),
	}
}

func (ev *textExtractionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	extractLocation := destination.Extract()
	if err := extractLocation.Validate(); err != nil {
		return err
	}

	
	sourceFile := source.GetPath()
	sourceFileType := source.GetType()
	sourceFileVersion := source.GetVersion()
	extractedFile := extractLocation.GetTargetFile()

	if err := ev.FileSegmentCounter.CompareTextSegmentsCount(sourceFile, extractedFile, sourceFileType, sourceFileVersion); err != nil {
		if err := common.RemoveFileWithRetries(extractedFile, 5, 5); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", extractedFile)
		}
		return err
	}

	return nil
}
