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
	ITextDecoder interface {
		// DecodeDialog decodes a dialog-type text file using the specified encoding information.
		// Source file types are: .bin, .msb and .00[0-6] from DCP file container.
		DecodeDialog(sourceFile, targetFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error
		// DecodeKernel decodes a kernel-type text file using the specified encoding information.
		// Source file types are: .bin
		DecodeKernel(sourceFile, targetFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding) error
	}

	// TextDecoder handles the decoding of FFX text files by invoking external tools.
	textDecoder struct {
		commandRunner command.ICommandRunner
	}
)

func NewTextDecoder(runner command.ICommandRunner) ITextDecoder {
	if runner == nil {
		runner = command.NewCommandRunner()
	}
	return &textDecoder{
		commandRunner: runner,
	}
}

func (d *textDecoder) DecodeDialog(sourceFile, targetFile string, encodingInfo ffxencoding.IFFXTextDlgEncoding) error {
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

	if err := d.runDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to decode dialog file: %w", err)
	}

	// Check if the target file was created successfully
	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("target file was not created: %w", err)
	}

	return nil
}

func (d *textDecoder) DecodeKernel(sourceFile, targetFile string, encodingInfo ffxencoding.IFFXTextKrnlEncoding) error {
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

	if err := d.runDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath); err != nil {
		return fmt.Errorf("failed to decode kernel file: %w", err)
	}

	// Check if the target file was created successfully
	if err := common.CheckPathExists(targetFile); err != nil {
		return fmt.Errorf("target file was not created: %w", err)
	}

	return nil
}

func (d *textDecoder) runDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath string) error {
	if err := common.CheckPathExists(executablePath); err != nil {
		return err
	}

	if err := common.CheckPathExists(sourceFile); err != nil {
		return err
	}

	if err := common.CheckPathExists(encodingFilePath); err != nil {
		return err
	}

	targetDir := filepath.Dir(targetFile)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Prepare arguments for the external tool
	// Assuming the tool expects: -e (extract/decode flag), -t <encoding_file>, <source_file>, <target_file>
	args := []string{"-e", "-t", encodingFilePath, sourceFile, targetFile}

	if output, err := d.commandRunner.Run(executablePath, args); err != nil {
		return fmt.Errorf("external decoder execution failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}
