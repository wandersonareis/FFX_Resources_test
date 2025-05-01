package internal

import (
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type IDlgEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination) error
}
type dlgEncoder struct {
	log logger.ILoggerHandler
}

func NewDlgEncoder(logger logger.ILoggerHandler) *dlgEncoder {
	return &dlgEncoder{
		log: logger,
	}
}

func (e *dlgEncoder) Encoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
	defer encoding.Dispose()

	translateLocation := destination.Translate()
	importLocation := destination.Import()

	if err := translateLocation.Validate(); err != nil {
		e.log.LogError(err, fmt.Sprintf("Error validating translate file: %s", translateLocation.GetTargetFile()))

		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		e.log.LogError(err, fmt.Sprintf("Error providing import path: %s", importLocation.GetTargetPath()))

		return err
	}

	sourceFile := source.Get().Path

	encoder := textsEncoding.NewEncoder()

	if err := encoder.DlgEncoder(sourceFile, translateLocation.GetTargetFile(), importLocation.GetTargetFile(), encoding); err != nil {
		e.log.LogError(err, fmt.Sprintf("Error on encoding dialog file: %s", sourceFile))

		return err
	}

	return nil
}
