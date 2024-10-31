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

	fmt.Println("Args: ", cmd.Args)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar comando: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		fmt.Println("Args: ", args)
		return fmt.Errorf("erro cmd.Wait na execução do comando: %w", err)
	}

	if stderr.Len() > 0 {
		fmt.Println("Stderr:", stderr.String())
		return fmt.Errorf("erro stderr na execução do comando: %s", stderr.String())
	}

	return nil
}
