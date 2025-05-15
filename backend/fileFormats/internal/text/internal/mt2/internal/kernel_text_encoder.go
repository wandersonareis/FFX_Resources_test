package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"fmt"
)

type IKrnlEncoder interface {
	Encoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextKrnlEncoding) error
}

type krnlEncoder struct {
	commandRunner command.ICommandRunner
}

func NewKrnlEncoder() IKrnlEncoder {
	return &krnlEncoder{
		commandRunner: command.NewCommandRunner(),
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
	sourceFileVersion := source.GetVersion()

	if err := e.encodeKernel(sourceFile, translatedFile, outputFile, textEncoding, sourceFileVersion); err != nil {
		return err
	}

	return nil
}

func (d *krnlEncoder) encodeKernel(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding, gameVersion models.GameVersion) error {
	encodingFilePath := encodingInfo.GetEncodingFile()
	if encodingFilePath == "" {
		return fmt.Errorf("kernel encoding file path is empty")
	}

	executablePath, err := encodingInfo.GetKrnlHandler().GetKernelTextHandler(gameVersion)
	if err != nil {
		return fmt.Errorf("failed to get kernel handler executable: %w", err)
	}

	if executablePath == "" {
		return fmt.Errorf("kernel handler executable path is empty")
	}

	if err := d.commandRunner.RunTextEncodingTool(executablePath, sourceFile, targetFile, outputFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to encode kernel file: %w", err)
	}

	// Check if the output file was created successfully
	if err := common.CheckPathExists(outputFile); err != nil {
		return fmt.Errorf("output file was not created: %w", err)
	}

	return nil
}