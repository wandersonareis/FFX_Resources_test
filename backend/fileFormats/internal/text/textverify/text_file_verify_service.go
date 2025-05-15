package textverify

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

type (
	ITextVerificationStrategy interface {
		Verify(source interfaces.ISource, destination locations.IDestination) error
	}

	ITextVerificationService interface {
		Verify(source interfaces.ISource, destination locations.IDestination, strategy ITextVerificationStrategy) error
	}

	TextVerificationService struct{}
)

func NewTextVerificationService() *TextVerificationService {
	return &TextVerificationService{}
}

func (svc *TextVerificationService) Verify(source interfaces.ISource, destination locations.IDestination, strategy ITextVerificationStrategy) error {
	if err := strategy.Verify(source, destination); err != nil {
		return err
	}

	return nil
}
