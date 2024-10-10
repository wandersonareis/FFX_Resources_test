package fileFormat

import (
	"ffxresources/backend/lib"
	"os/exec"
)

/* func getDcpFileXpliter() (string, error) {
	exeName := "bin/SHSplit.exe"
	return lib.GetToolExcutable(exeName)
} */

func getDcpFileXpliterDev() (string, error) {
	handlerPath, err := lib.GetDcpXplitHandler()
	if err != nil {
		return "", err
	}
	return exec.LookPath(handlerPath)
}
