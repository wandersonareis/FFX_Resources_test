package textVerifier

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type (
	ITextVerifyInstance interface {
		Verify(source interfaces.ISource, destination locations.IDestination) error
	}

	ITextVerifier interface {
		Verify(source interfaces.ISource, destination locations.IDestination, verifyInstance ITextVerifyInstance) error
	}

	TextVerifier struct {
		log logger.ILoggerHandler
	}
)

func NewTextsVerify(logger logger.ILoggerHandler) *TextVerifier {
	return &TextVerifier{
		log: logger,
	}
}

func (dv *TextVerifier) Verify(source interfaces.ISource, destination locations.IDestination, verifyInstance ITextVerifyInstance) error {
	if err := verifyInstance.Verify(source, destination); err != nil {
		return err
	}

	return nil
}
