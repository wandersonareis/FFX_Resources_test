package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type IDlgDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error
}

type dlgDecoder struct {
	log logger.ILoggerHandler
}

func NewDlgDecoder(logger logger.ILoggerHandler) IDlgDecoder {
	return &dlgDecoder{
		log: logger,
	}
}

func (d *dlgDecoder) Decoder(
	source interfaces.ISource,
	destination locations.IDestination,
	textEncoding ffxencoding.IFFXTextDlgEncoding) error {
	extractLocation := destination.Extract().Get()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		d.log.LogError(err, "Error providing extract path")
		return err
	}

	decoder := textsEncoding.NewDecoder()

	sourceFile := source.Get().Path
	extractFile := extractLocation.GetTargetFile()

	if err := decoder.DlgDecoder(sourceFile, extractFile, textEncoding); err != nil {
		d.log.LogError(err, "Error on decoding dialog file")
		return err
	}

	return nil
}
