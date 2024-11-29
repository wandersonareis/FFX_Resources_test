package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"

	"github.com/rs/zerolog"
)

type IKrnlDecoder interface {
	Decoder(fileInfo interactions.IGameDataInfo) error
}

type krnlDecoder struct {
	log zerolog.Logger
}

func NewKrnlDecoder() IKrnlDecoder {
	return &krnlDecoder{
		log: logger.Get().With().Str("module", "kernel_file_decoder").Logger(),
	}
}

func (d *krnlDecoder) Decoder(fileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	extractLocation := fileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		d.log.Error().Err(err).Msgf("Error providing extract path: %s", extractLocation.TargetPath)
		return err
	}

	decoder := textsEncoding.NewDecoder()

	sourceFile := fileInfo.GetGameData().FullFilePath

	if err := decoder.KnrlDecoder(sourceFile, extractLocation.TargetFile, encoding); err != nil {
		d.log.Error().Err(err).Msgf("Error on decoding kernel file: %s", sourceFile)
		return err
	}

	return nil
}
