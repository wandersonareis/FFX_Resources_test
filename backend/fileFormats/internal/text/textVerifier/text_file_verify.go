package textVerifier

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type (
	ITextVerificationStrategy interface {
		Verify(source interfaces.ISource, destination locations.IDestination) error
	}

	ITextVerificationService interface {
		Verify(source interfaces.ISource, destination locations.IDestination, strategy ITextVerificationStrategy) error
	}

	TextVerificationService struct {
		log logger.ILoggerHandler
	}
)

func NewTextVerificationService(logger logger.ILoggerHandler) *TextVerificationService {
	return &TextVerificationService{
		log: logger,
	}
}

func (svc *TextVerificationService) Verify(source interfaces.ISource, destination locations.IDestination, strategy ITextVerificationStrategy) error {
	if err := strategy.Verify(source, destination); err != nil {
		return err
	}

	return nil
}
