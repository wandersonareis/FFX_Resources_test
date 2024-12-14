package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IKrnlEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination) error
}

type krnlEncoder struct {
	log zerolog.Logger
}

func NewKrnlEncoder() IKrnlEncoder {
	return &krnlEncoder{
		log: logger.Get().With().Str("module", "kernel_file_encoder").Logger(),
	}
}

func (e *krnlEncoder) Encoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	translateLocation := destination.Translate().Get()
	importLocation := destination.Import().Get()

	if err := translateLocation.Validate(); err != nil {
		e.log.Error().
			Err(err).
			Str("path", translateLocation.GetTargetFile()).
			Msg("Error validating translate file")

		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		e.log.Error().
			Err(err).
			Str("path", importLocation.GetTargetPath()).
			Msg("Error providing import path")

		return err
	}

	sourceFile := source.Get().Path

	encoder := textsEncoding.NewEncoder()

	if err := encoder.KnrlEncoder(sourceFile, translateLocation.GetTargetFile(), importLocation.GetTargetFile(), encoding); err != nil {
		e.log.Error().
			Err(err).
			Str("file", sourceFile).
			Msg("Error on encoding kernel file")

		return err
	}

	return nil
}
