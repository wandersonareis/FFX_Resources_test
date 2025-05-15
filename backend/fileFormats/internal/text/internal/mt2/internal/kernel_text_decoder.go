package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IKrnlDecoder interface {
	Decoder(source interfaces.ISource, destination locations.IDestination) error
}

type krnlDecoder struct {
	CommandRunner command.ICommandRunner
}

func NewKrnlDecoder() IKrnlDecoder {
	return &krnlDecoder{
		CommandRunner: command.NewCommandRunner(),
	}
}

func (d *krnlDecoder) Decoder(source interfaces.ISource, destination locations.IDestination) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	if err := d.decodeKernel(source, destination.Extract().GetTargetFile(), encoding); err != nil {
		return err
	}

	return nil
}

func (d *krnlDecoder) decodeKernel(source interfaces.ISource, targetFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding) error {
	encodingFilePath := encodingInfo.GetEncodingFile()
	if encodingFilePath == "" {
		return fmt.Errorf("kernel encoding file path is empty")
	}

	sourceFile := source.GetPath()
	sourceFileVersion := source.GetVersion()

	executablePath, err := encodingInfo.GetKrnlHandler().GetKernelTextHandler(sourceFileVersion)
	if err != nil {
		return fmt.Errorf("failed to get kernel handler executable: %w", err)
	}

	if executablePath == "" {
		return fmt.Errorf("kernel handler executable path is empty")
	}

	if err := d.CommandRunner.RunTextDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to decode kernel file: %w", err)
	}

	// Check if the target file was created successfully
	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("target file was not created: %w", err)
	}

	return nil
}
