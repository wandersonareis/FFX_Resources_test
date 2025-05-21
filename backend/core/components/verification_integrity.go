package components

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

type (
	IVerificationStrategy interface {
		Verify(source interfaces.ISource, destination locations.IDestination) error
	}

	IVerificationService interface {
		Verify(source interfaces.ISource, destination locations.IDestination, strategy IVerificationStrategy) error
	}

	VerificationService struct{}
)

func NewVerificationService() *VerificationService {
	return &VerificationService{}
}

func (svc *VerificationService) Verify(source interfaces.ISource, destination locations.IDestination, strategy IVerificationStrategy) error {
	if err := strategy.Verify(source, destination); err != nil {
		return err
	}

	return nil
}
