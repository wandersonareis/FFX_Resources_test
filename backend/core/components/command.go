package components

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func RunCommand(tool string, args []string) (string, error) {
	cmd := exec.Command(tool, args...)
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

func splitResult(stdout string) []string {
	return strings.Split(stdout, "\n")
}

func parseCountFromOutput(lines []string, linePrefix, formatStr string) (int, error) {
	for _, line := range lines {
		if strings.HasPrefix(line, linePrefix) {
			var count int
			_, err := fmt.Sscanf(line, formatStr, &count)
			if err != nil {
				return 0, fmt.Errorf("parse error: %w", err)
			}
			return count, nil
		}
	}
	return 0, fmt.Errorf("count not found for prefix: %s", linePrefix)
}

func GetDialogSegmentsCount(tool string, args []string) (int, error) {
	stdout, err := RunCommand(tool, args)
	if err != nil {
		return 0, err
	}

	lines := splitResult(stdout)
	linePrefix := "header elements (nonzero)"
	lineFormat := "header elements (nonzero): %d"

	return parseCountFromOutput(lines, linePrefix, lineFormat)
}

func GetKernelSegmentsCount(tool string, args []string) (int, error) {
	stdout, err := RunCommand(tool, args)
	if err != nil {
		return 0, err
	}

	lines := splitResult(stdout)
	linePrefix := "texts count:"
	lineFormat := "texts count: %d"

	return parseCountFromOutput(lines, linePrefix, lineFormat)
}
