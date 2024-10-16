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
		return fmt.Errorf("erro ao iniciar comando: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("erro na execução do comando: %w", err)
	}

	if stderr.Len() > 0 {
		fmt.Println("Stderr:", stderr.String())
		return fmt.Errorf("erro na execução do comando: %s", stderr.String())
	}

	return nil
}
