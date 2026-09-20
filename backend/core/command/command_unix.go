//go:build !windows

package command

import "os/exec"
import "errors"

func configureCommand(_ *exec.Cmd) error {
	return errors.New("RunCommand: HideWindow is only supported on Windows")
}
