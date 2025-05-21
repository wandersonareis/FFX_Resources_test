package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IDlgDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error
}

type dlgDecoder struct {
	CommandRunner command.ICommandRunner
}

func NewDlgDecoder() IDlgDecoder {
	return &dlgDecoder{
		CommandRunner: command.NewCommandRunner(),
	}
}

func (d *dlgDecoder) Decoder(
	source interfaces.ISource,
	destination locations.IDestination,
	textEncoding ffxencoding.IFFXTextDlgEncoding) error {
	sourceFile := source.GetPath()
	extractFile := destination.Extract().GetTargetFile()

	if extractFile == "" {
		return fmt.Errorf("extract target file path is empty")
	}

	if err := d.decodeDialog(sourceFile, extractFile, textEncoding); err != nil {
		return err
	}

	return nil
}

func (d *dlgDecoder) decodeDialog(sourceFile, targetFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error {
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

	if err := d.CommandRunner.RunTextDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to decode dialog file: %w", err)
	}

	// Check if the target file was created successfully
	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("target file was not created: %w", err)
	}

	return nil
}