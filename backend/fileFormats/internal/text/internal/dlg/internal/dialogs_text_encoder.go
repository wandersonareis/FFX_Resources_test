package internal

import (
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/internal/encoding"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IDlgEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error
}
type dlgEncoder struct {
	TextEncoder textsEncoding.ITextEncoder
}

func NewDlgEncoder() IDlgEncoder {
	return &dlgEncoder{
		TextEncoder: textsEncoding.NewTextEncoder(command.NewCommandRunner()),
	}
}

func (e *dlgEncoder) Encoder(
	source interfaces.ISource,
	destination locations.IDestination,
	textEncoding ffxencoding.IFFXTextDlgEncoding) error {
	translatedFile := destination.Translate().GetTargetFile()
	outputFile := destination.Import().GetTargetFile()

	if err := destination.Translate().Validate(); err != nil {
		return fmt.Errorf("error validating translate file: %s | error: %w", translatedFile, err)
	}

	sourceFile := source.Get().Path

	if err := e.TextEncoder.EncodeDialog(sourceFile, translatedFile, outputFile, textEncoding); err != nil {
		return err
	}

	return nil
}
