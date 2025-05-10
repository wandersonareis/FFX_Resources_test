package internal

import (
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IKrnlEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextKrnlEncoding) error
}

type krnlEncoder struct {
	TextEncoder textsEncoding.ITextEncoder
}

func NewKrnlEncoder() IKrnlEncoder {
	return &krnlEncoder{
		TextEncoder: textsEncoding.NewTextEncoder(command.NewCommandRunner()),
	}
}

func (e *krnlEncoder) Encoder(
	source interfaces.ISource,
	destination locations.IDestination,
	textEncoding ffxencoding.IFFXTextKrnlEncoding) error {
	translatedFile := destination.Translate().GetTargetFile()
	outputFile := destination.Import().GetTargetFile()

	if err := destination.Translate().Validate(); err != nil {
		return fmt.Errorf("error validating translate file: %s | error: %w", translatedFile, err)
	}

	sourceFile := source.GetPath()

	if err := e.TextEncoder.EncodeKernel(sourceFile, translatedFile, outputFile, textEncoding); err != nil {
		return err
	}

	return nil
}
