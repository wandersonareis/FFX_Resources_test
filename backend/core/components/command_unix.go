//go:build !windows

package components

import (
	"errors"
	"os/exec"
)

func configureCommand(_ *exec.Cmd) error {
	return errors.New("command configuration is not supported on this platform")
}
