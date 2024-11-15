package lib

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

func RunCommand(tool string, args []string) error {
	cmd := exec.Command(tool, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error cmd.Wait:\n %w", err)
	}

	if stderr.Len() > 0 {
		return fmt.Errorf("error stderr:\n %s", stderr.String())
	}

	return nil
}