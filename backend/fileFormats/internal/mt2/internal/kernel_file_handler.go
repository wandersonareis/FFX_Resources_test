package internal

import (
	"ffxresources/backend/core"
	"ffxresources/backend/fileFormats/util"
	"fmt"
	"os"
)

type kernelHandler struct {
	gamePart core.GamePart
	handler  string
}

func newKernelHandler(gamePart core.GamePart) *kernelHandler {
	return &kernelHandler{
		gamePart: gamePart,
	}
}

func (kh *kernelHandler) getKernelFileHandler() (string, error) {
	var (
		err        error
		targetFile string
	)

	switch kh.gamePart {
	case core.FFX:
		targetFile, err = util.GetFromResources(util.KERNEL_HANDLER_RESOURCES_DIR, util.FFX_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	case core.FFX2:
		targetFile, err = util.GetFromResources(util.KERNEL_HANDLER_RESOURCES_DIR, util.FFX2_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	default:
		return "", fmt.Errorf("unknown game part: %v", kh.gamePart)
	}

	if err != nil {
		return "", err
	}

	kh.handler = targetFile

	return kh.handler, nil
}

func (kh *kernelHandler) Dispose() {
	if kh.handler != "" {
		os.Remove(kh.handler)
	}
}
