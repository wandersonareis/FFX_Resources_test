package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IDlgEncoder interface {
	Encoder(fileInfo interactions.IGameDataInfo) error
}
type dlgEncoder struct {
	log zerolog.Logger
}

func NewDlgEncoder() *dlgEncoder {
	return &dlgEncoder{
		log: logger.Get().With().Str("module", "dialogs_file_encoder").Logger(),
	}
}

func (e *dlgEncoder) Encoder(fileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(fileInfo.GetGameData().Type)
	defer encoding.Dispose()

	translateLocation := fileInfo.GetTranslateLocation()
	importLocation := fileInfo.GetImportLocation()

	if err := translateLocation.Validate(); err != nil {
		e.log.Error().
			Err(err).
			Str("file", translateLocation.TargetFile).
			Msg("Error validating translate file")

		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		e.log.Error().
			Err(err).
			Str("path", importLocation.TargetPath).
			Msg("Error providing import path")

		return err
	}

	sourceFile := fileInfo.GetGameData().FullFilePath

	encoder := textsEncoding.NewEncoder()

	if err := encoder.DlgEncoder(sourceFile, translateLocation.TargetFile, importLocation.TargetFile, encoding); err != nil {
		e.log.Error().
			Err(err).
			Str("file", sourceFile).
			Msg("Error on encoding dialog file")
			
		return err
	}

	return nil
}
