package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IKrnlDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination) error
}

type krnlDecoder struct {
	log zerolog.Logger
}

func NewKrnlDecoder() IKrnlDecoder {
	return &krnlDecoder{
		log: logger.Get().With().Str("module", "kernel_file_decoder").Logger(),
	}
}

func (d *krnlDecoder) Decoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	extractLocation := destination.Extract().Get()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		d.log.Error().
			Err(err).
			Str("path", extractLocation.GetTargetPath()).
			Msg("Error providing extract path")

		return err
	}

	decoder := textsEncoding.NewDecoder()

	sourceFile := source.Get().Path

	if err := decoder.KnrlDecoder(sourceFile, extractLocation.GetTargetFile(), encoding); err != nil {
		d.log.Error().
			Err(err).
			Str("file", sourceFile).
			Msg("Error on decoding kernel file")

		return err
	}

	return nil
}
