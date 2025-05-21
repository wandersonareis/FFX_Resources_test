package textverify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type textCompressionVerificationStrategy struct {
	verifyService       ITextVerificationService
}

func NewTextCompressionVerificationStrategy() ITextVerificationStrategy {
	return &textCompressionVerificationStrategy{
		verifyService:       NewTextVerificationService(),
	}
}

func (cv *textCompressionVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	if err := destination.Translate().Validate(); err != nil {
		return err
	}

	if err := destination.Extract().Validate(); err != nil {
		return err
	}

	if err := cv.compareTextSegmentsCount(source, destination); err != nil {
		return err
	}

	if err := cv.compareFileData(source, destination); err != nil {
		if err := common.RemoveFileWithRetries(destination.Import().GetTargetFile(), 5); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", destination.Import().GetTargetFile())
		}

		return err
	}

	return nil
}

func (cv *textCompressionVerificationStrategy) compareTextSegmentsCount(source interfaces.ISource, destination locations.IDestination) error {
	destinationFile := destination.Translate().GetTargetFile()
	if err := cv.verifyService.Verify(source, destination, NewTextSegmentsVerificationStrategy(destinationFile)); err != nil {
		return fmt.Errorf("an error occurred while verifying segments count of the compressed file '%s': %v", destinationFile, err)
	}

	return nil
}

func (cv *textCompressionVerificationStrategy) compareFileData(source interfaces.ISource, destination locations.IDestination) error {
	if err := cv.verifyService.Verify(source, destination, NewTextContentComparerStrategy()); err != nil {
		destinationFile := destination.Translate().GetTargetFile()
		return fmt.Errorf("an error occurred while verifying the content of the compressed file '%s': %v", destinationFile, err)
	}

	return nil
}
