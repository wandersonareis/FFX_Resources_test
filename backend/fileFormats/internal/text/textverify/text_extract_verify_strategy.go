package textverify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type textExtractionVerificationStrategy struct {
	verifyService ITextVerificationService
}

func NewTextExtractionVerificationStrategy() ITextVerificationStrategy {
	return &textExtractionVerificationStrategy{
		verifyService: NewTextVerificationService(),
	}
}

func (ev *textExtractionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	extractLocation := destination.Extract()
	if err := extractLocation.Validate(); err != nil {
		return err
	}
	
	if err := ev.compareTextSegmentsCount(source, destination); err != nil {
		extractedFile := extractLocation.GetTargetFile()
		if err := common.RemoveFileWithRetries(extractedFile, 5, 5); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", extractedFile)
		}
		return err
	}

	return nil
}

func (ev *textExtractionVerificationStrategy) compareTextSegmentsCount(source interfaces.ISource, destination locations.IDestination) error {
	destinationFile := destination.Extract().GetTargetFile()
	if err := ev.verifyService.Verify(source, destination, NewTextSegmentsVerificationStrategy(destinationFile)); err != nil {
		return fmt.Errorf("an error occurred while verifying segments count of the extracted file: %v", err)
	}

	return nil
}