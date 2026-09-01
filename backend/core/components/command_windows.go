//go:build windows

package components

import (
	"os/exec"
	"syscall"
)

func configureCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return nil
}
