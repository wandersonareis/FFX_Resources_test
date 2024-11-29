package internal

import (
	"ffxresources/backend/core"
	"ffxresources/backend/fileFormats/util"
	"fmt"
	"os"
)

type kernelHandler struct {
	util.Checksum
	gamePart core.GamePart
	handler  string
}

func newKernelHandler(gamePart core.GamePart) *kernelHandler {
	return &kernelHandler{
		Checksum: util.Checksum{},
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
		targetFile, err = util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

		if !kh.VerifyChecksum(targetFile, FFX_KERNEL_HANDLER_SHA256) {
			return "", fmt.Errorf("invalid checksum for kernel file handler")
		}
	case core.FFX2:
		targetFile, err = util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX2_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

		if !kh.VerifyChecksum(targetFile, FFX2_KERNEL_HANDLER_SHA256) {
			return "", fmt.Errorf("invalid checksum for kernel file handler")
		}
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
