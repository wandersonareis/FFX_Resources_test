package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IDlgDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination) error
}

type dlgDecoder struct {
	log zerolog.Logger
}

func NewDlgDecoder() IDlgDecoder {
	return &dlgDecoder{
		log: logger.Get().With().Str("module", "dialogs_file_decoder").Logger(),
	}
}

func (d *dlgDecoder) Decoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.Get().Type)
	defer encoding.Dispose()

	sourceFile := source.Get().Path

	extractLocation := destination.Extract().Get()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		d.log.Error().
			Err(err).
			Str("path", extractLocation.GetTargetPath()).
			Msg("Error providing extract path")

		return err
	}

	decoder := textsEncoding.NewDecoder()

	if err := decoder.DlgDecoder(sourceFile, extractLocation.GetTargetFile(), encoding); err != nil {
		d.log.Error().
			Err(err).
			Str("file", sourceFile).
			Msg("Error on decoding dialog file")

		return err
	}

	return nil
}
