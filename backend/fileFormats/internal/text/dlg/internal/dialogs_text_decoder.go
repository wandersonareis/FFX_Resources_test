package internal

import (
	"ffxresources/backend/core/encoding"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IDlgDecoder interface {
	Decoder(dialogsFileInfo interactions.IGameDataInfo) error
}

type dlgDecoder struct {
	log zerolog.Logger
}

func NewDlgDecoder() IDlgDecoder {
	return &dlgDecoder{
		log: logger.Get().With().Str("module", "dialogs_file_decoder").Logger(),
	}
}

func (d *dlgDecoder) Decoder(dialogsFileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(dialogsFileInfo.GetGameData().Type)
	defer encoding.Dispose()

	sourceFile := dialogsFileInfo.GetGameData().FullFilePath

	extractLocation := dialogsFileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		d.log.Error().
			Err(err).
			Str("path", extractLocation.TargetPath).
			Msg("Error providing extract path")

		return err
	}

	decoder := textsEncoding.NewDecoder()

	if err := decoder.DlgDecoder(sourceFile, extractLocation.TargetFile, encoding); err != nil {
		d.log.Error().
			Err(err).
			Str("file", sourceFile).
			Msg("Error on decoding dialog file")

		return err
	}

	return nil
}
