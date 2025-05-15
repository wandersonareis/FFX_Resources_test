package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IDlgEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error
}
type dlgEncoder struct {
	commandRunner command.ICommandRunner
}

func NewDlgEncoder() IDlgEncoder {
	return &dlgEncoder{
		commandRunner: command.NewCommandRunner(),
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

	sourceFile := source.GetPath()

	if err := e.encodeDialog(sourceFile, translatedFile, outputFile, textEncoding); err != nil {
		return err
	}

	return nil
}

func (d *dlgEncoder) encodeDialog(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error {
	encodingFilePath := encodingInfo.GetEncoding()
	if encodingFilePath == "" {
		return fmt.Errorf("dialog encoding file path is empty")
	}

	executablePath, err := encodingInfo.GetDlgHandler().GetDlgHandlerApp()
	if err != nil {
		return fmt.Errorf("failed to get dialog handler executable: %w", err)
	}

	if executablePath == "" {
		return fmt.Errorf("dialog handler executable path is empty")
	}

	if err := d.commandRunner.RunTextEncodingTool(executablePath, sourceFile, targetFile, outputFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to encode dialog file: %w", err)
	}

	// Check if the output file was created successfully
	if err := common.CheckPathExists(outputFile); err != nil {
		return fmt.Errorf("output file was not created: %w", err)
	}

	return nil
}