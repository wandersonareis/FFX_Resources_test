package textVerifier

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
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

	extractedFile := extractLocation.GetTargetFile()

	sourceFileType := source.Get().Type
	sourceFile := source.Get().Path

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	if err := ev.FileSegmentCounter.CompareTextSegmentsCount(sourceFile, extractedFile, sourceFileType, gameVersion); err != nil {
		if err := common.RemoveFileWithRetries(extractedFile, 5, 5); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", extractedFile)
		}
		return err
	}

	return nil
}
