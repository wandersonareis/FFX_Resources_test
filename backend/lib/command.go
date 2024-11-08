package lib

import (
	"bytes"
	"ffxresources/backend/events"
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
		events.LogSeverity(events.SeverityError, fmt.Sprintf("erro ao iniciar comando: %s", err.Error()))
		return fmt.Errorf("erro ao iniciar comando: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		events.LogSeverity(events.SeverityError, fmt.Sprintf("erro cmd.Wait na execução do comando: %s", err.Error()))
		return fmt.Errorf("erro cmd.Wait na execução do comando: %w", err)
	}

	if stderr.Len() > 0 {
		events.LogSeverity(events.SeverityError, fmt.Sprintf("erro stderr na execução do comando: %s", stderr.String()))
		return fmt.Errorf("erro stderr na execução do comando: %s", stderr.String())
	}

	return nil
}