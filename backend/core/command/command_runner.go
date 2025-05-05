package command

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

type (
	// ICommandRunner defines the interface for running external commands.
	ICommandRunner interface {
		// Run executes a command with the given arguments.
		// It returns the combined standard output and standard error, and an error if the command fails.
		Run(executablePath string, args []string) (string, error)
	}
	commandRunner struct{}
)

func NewCommandRunner() ICommandRunner {
	return &commandRunner{}
}

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
