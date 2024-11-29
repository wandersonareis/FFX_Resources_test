package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IKrnlEncoder interface {
	Encoder(fileInfo interactions.IGameDataInfo) error
}

type krnlEncoder struct {
	log zerolog.Logger
}

func NewKrnlEncoder() IKrnlEncoder {
	return &krnlEncoder{
		log: logger.Get().With().Str("module", "kernel_file_encoder").Logger(),
	}
}

func (e *krnlEncoder) Encoder(fileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	translateLocation := fileInfo.GetTranslateLocation()
	importLocation := fileInfo.GetImportLocation()

	if err := translateLocation.Validate(); err != nil {
		e.log.Error().Err(err).Msgf("Error validating translate file: %s", translateLocation.TargetFile)
		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		e.log.Error().Err(err).Msgf("Error providing import path: %s", importLocation.TargetPath)
		return err
	}

	sourceFile := fileInfo.GetGameData().FullFilePath

	encoder := textsEncoding.NewEncoder()

	if err := encoder.KnrlEncoder(sourceFile, translateLocation.TargetFile, importLocation.TargetFile, encoding); err != nil {
		e.log.Error().Err(err).Msgf("Error on encoding kernel file: %s", sourceFile)
		return err
	}

	return nil
}
