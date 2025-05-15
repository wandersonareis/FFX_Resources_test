package command

import (
	"bytes"
	"ffxresources/backend/common"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type (
	// ICommandRunner defines the interface for running external commands.
	ICommandRunner interface {
		// Run executes a command with the given arguments.
		// It returns the combined standard output and standard error, and an error if the command fails.
		Run(executablePath string, args []string) (string, error)

		// RunTextDecodingTool executes an external text decoding tool with the specified parameters.
		// It checks the existence of the executable, source file, and encoding file before execution.
		// The function ensures the target directory exists, then runs the external tool with the following arguments:
		//   - "-e" to indicate extraction/decoding mode
		//   - "-t" followed by the encoding file path
		//   - the source file path
		//   - the target file path
		//
		// If the external tool execution fails, an error is returned containing the tool's output.
		//
		// This function is compatible with processing dialog files such as .bin, .msg, and .00[1-6] from .dcp archives,
		// as well as kernel files with the .bin extension.
		//
		// Parameters:
		//   - executablePath: Path to the external decoding tool executable.
		//   - sourceFile: Path to the source file to be decoded.
		//   - targetFile: Path where the decoded output should be written.
		//   - encodingFilePath: Path to the encoding definition file.
		//
		// Returns:
		//   - error: An error if any step fails, or nil on success.
		RunTextDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath string) error

		// RunTextEncodingTool executes an external text encoding tool with the specified parameters.
		// It checks the existence of the executable, source, target, and encoding files before execution.
		// The function ensures the output directory exists, then runs the external tool with the following arguments:
		//   - "-i" to indicate compression/encoding mode
		//   - "-t" followed by the encoding file path
		//   - the source file path
		//   - the target file path
		//   - the output file path
		//
		// If the external tool execution fails, an error is returned containing the tool's output.
		//
		// This function is compatible with processing dialog files such as .bin, .msg, and .00[1-6] from .dcp archives,
		// as well as kernel files with the .bin extension.
		//
		// Parameters:
		//   - executablePath: Path to the external encoding tool executable.
		//   - sourceFile: Path to the source file to be encoded.
		//   - targetFile: Path to the target file used in the encoding process.
		//   - outputFile: Path where the output file will be written.
		//   - encodingFile: Path to the encoding configuration file.
		//
		// Returns:
		//   - error: An error if any file does not exist, the output directory cannot be created,
		//     or the external tool execution fails; otherwise, nil.
		RunTextEncodingTool(executablePath, sourceFile, targetFile, outputFile, encodingFile string) error
	}

	commandRunner struct{}
)

func NewCommandRunner() ICommandRunner {
	return &commandRunner{}
}

// Run executes an external command specified by the executablePath with the provided arguments.
// It captures both stdout and stderr output, and hides the command window on Windows systems.
// Returns the standard output as a string if the command executes successfully.
// If the command fails to start or returns an error, it includes any stderr output in the error message.
// This function is compatible with running external tools for processing dialog and kernel files,
// such as those with extensions .bin, .msb, and .00[1-6] (for dialogs in DCP archives) and .bin (for kernel).
func (c *commandRunner) Run(executablePath string, args []string) (string, error) {
	cmd := exec.Command(executablePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("error stderr:\n %s", stderr.String())
		}
		return "", fmt.Errorf("error cmd.Wait:\n %w", err)
	}

	return stdout.String(), nil
}

func (c *commandRunner) RunTextDecodingTool(executablePath, sourceFile, targetFile, encodingFilePath string) error {
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

	if output, err := c.Run(executablePath, args); err != nil {
		return fmt.Errorf("external decoder execution failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}

func (e *commandRunner) RunTextEncodingTool(
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

	if output, err := e.Run(executablePath, args); err != nil {
		return fmt.Errorf("external encoder execution failed: %w\noutput:\n%s", err, output)
	}

	return nil
}
