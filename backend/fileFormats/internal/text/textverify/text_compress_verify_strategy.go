package textverify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
	"time"
)

type textCompressionVerificationStrategy struct {
	FileSegmentCounter  ISegmentCounter
	FileContentComparer IComparer
}

func NewTextCompressionVerificationStrategy() ITextVerificationStrategy {
	return &textCompressionVerificationStrategy{
		FileSegmentCounter:  newSegmentCounter(),
		FileContentComparer: newPartComparer(),
	}
}

func (cv *textCompressionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Translate().Validate(); err != nil {
		return err
	}

	if err := destination.Extract().Validate(); err != nil {
		return err
	}

	if err := cv.FileContentComparer.CompareTextPartsContents(destination.Translate().GetTargetFile(), destination.Extract().GetTargetFile()); err != nil {
		if err := common.RemoveFileWithRetries(destination.Import().GetTargetFile(), 5, time.Second); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", destination.Import().GetTargetFile())
		}

		return err
	}

	return nil
}
