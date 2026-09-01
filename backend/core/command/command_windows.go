//go:build windows

package command

import (
	"os/exec"
	"syscall"
)

func configureCommand(cmd *exec.Cmd) error  {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return nil
}
