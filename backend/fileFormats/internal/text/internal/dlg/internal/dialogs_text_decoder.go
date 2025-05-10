package internal

import (
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
)

type IDlgDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error
}

type dlgDecoder struct {
	TextDecoder textsEncoding.ITextDecoder
}

func NewDlgDecoder() IDlgDecoder {
	return &dlgDecoder{
		TextDecoder: textsEncoding.NewTextDecoder(command.NewCommandRunner()),
	}
}

func (d *dlgDecoder) Decoder(
	source interfaces.ISource,
	destination locations.IDestination,
	textEncoding ffxencoding.IFFXTextDlgEncoding) error {
	sourceFile := source.GetPath()
	extractFile := destination.Extract().GetTargetFile()

	if err := d.TextDecoder.DecodeDialog(sourceFile, extractFile, textEncoding); err != nil {
		return err
	}

	return nil
}
