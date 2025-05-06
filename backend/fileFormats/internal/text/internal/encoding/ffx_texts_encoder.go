package textsEncoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/command"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"path/filepath"
)

type (
	ITextEncoder interface {
		// EncodeDialog encodes a dialog-type text file using the specified encoding information.
		// Target file types are: .bin, .msb and .00[0-6] from DCP file container.
		EncodeDialog(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error
		// EncodeKernel encodes a kernel-type text file using the specified encoding information.
		// Target file types are: .bin
		EncodeKernel(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding) error
	}

	textsEncoder struct {
		commandRunner command.ICommandRunner
	}
)

func NewTextEncoder(runner command.ICommandRunner) ITextEncoder {
	if runner == nil {
		runner = command.NewCommandRunner()
	}
	return &textsEncoder{
		commandRunner: runner,
	}
}

func (d *textsEncoder) EncodeDialog(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error {
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

	if err := d.runEncodingTool(executablePath, sourceFile, targetFile, outputFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to encode dialog file: %w", err)
	}

	// Check if the output file was created successfully
	if err := common.CheckPathExists(outputFile); err != nil {
		return fmt.Errorf("output file was not created: %w", err)
	}

	return nil
}

func (d *textsEncoder) EncodeKernel(sourceFile, targetFile, outputFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding) error {
	encodingFilePath := encodingInfo.GetEncodingFile()
	if encodingFilePath == "" {
		return fmt.Errorf("kernel encoding file path is empty")
	}

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	executablePath, err := encodingInfo.GetKrnlHandler().GetKernelTextHandler(gameVersion)
	if err != nil {
		return fmt.Errorf("failed to get kernel handler executable: %w", err)
	}

	if executablePath == "" {
		return fmt.Errorf("kernel handler executable path is empty")
	}

	if err := d.runEncodingTool(executablePath, sourceFile, targetFile, outputFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to encode kernel file: %w", err)
	}

	// Check if the output file was created successfully
	if err := common.CheckPathExists(outputFile); err != nil {
		return fmt.Errorf("output file was not created: %w", err)
	}

	return nil
}

func (e *textsEncoder) runEncodingTool(
	executablePath, sourceFile, targetFile, outputFile, encodingFile string,
) error {
	if err := common.CheckPathExists(executablePath); err != nil {
		return fmt.Errorf("executable path does not exist: %w", err)
	}

	if err := common.CheckPathExists(sourceFile); err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("target file does not exist: %w", err)
	}

	if err := common.CheckPathExists(encodingFile); err != nil {
		return fmt.Errorf("encoding file does not exist: %w", err)
	}

	targetDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Prepare arguments for the external tool
	// Assuming the tool expects: -i (compress/encode flag), -t <encoding file>, <source file>, <target file>
	// and <output file>
	args := []string{"-i", "-t", encodingFile, sourceFile, targetFile, outputFile}

	if output, err := e.commandRunner.Run(executablePath, args); err != nil {
		return fmt.Errorf("external encoder execution failed: %w\noutput:\n%s", err, output)
	}

	return nil
}
