package textVerifier

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"fmt"
	"time"
)

type textSegmentsVerificationStrategy struct {
	FileSegmentCounter ISegmentCounter
}

func NewTextSegmentsVerificationStrategy() ITextVerificationStrategy {
	return &textSegmentsVerificationStrategy{
		FileSegmentCounter: newSegmentCounter(),
	}
}

func (sv *textSegmentsVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	sourceFile := source.Get().Path
	sourceFileType := source.Get().Type
	targetFile := destination.Translate().GetTargetFile()

	if err := common.CheckPathExists(sourceFile); err != nil {
		return fmt.Errorf("failed to check source file path: %s", err)
	}

	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("failed to check target file path: %s", err)
	}

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	if err := sv.FileSegmentCounter.CompareTextSegmentsCount(sourceFile, targetFile, sourceFileType, gameVersion); err != nil {
		if err := common.RemoveFileWithRetries(targetFile, 5, time.Second); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", targetFile)
		}
		return err
	}

	return nil
}
