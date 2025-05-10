package internal

import (
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
)

type IKrnlDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination) error
}

type krnlDecoder struct {
	TextDecoder textsEncoding.ITextDecoder
}

func NewKrnlDecoder() IKrnlDecoder {
	return &krnlDecoder{
		TextDecoder: textsEncoding.NewTextDecoder(command.NewCommandRunner()),
	}
}

func (d *krnlDecoder) Decoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	extractLocation := destination.Extract()

	sourceFile := source.GetPath()

	if err := d.TextDecoder.DecodeKernel(sourceFile, extractLocation.GetTargetFile(), encoding); err != nil {
		return err
	}

	return nil
}
