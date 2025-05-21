package encodingHandler

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/models"
	"fmt"
)

type (
	IKernelTextHandler interface {
		GetKernelTextHandler(gameVersion models.GameVersion) (string, error)
		Dispose()
	}

	kernelTextHandler struct {
		util.Checksum

		kernelTextAppFile string
	}
)

func NewKrnlTextsHandler() IKernelTextHandler {
	return &kernelTextHandler{
		Checksum: util.Checksum{},
	}
}

func (kth *kernelTextHandler) GetKernelTextHandler(gameVersion models.GameVersion) (string, error) {
	switch gameVersion {
	case models.FFX:
		return kth.ffxKernelTextHandler()
	case models.FFX2:
		return kth.ffx2KernelTexthandler()
	default:
		return "", fmt.Errorf("game version not supported for kernel text handler: %s", gameVersion)
	}
}

func (kth *kernelTextHandler) ffxKernelTextHandler() (string, error) {
	targetFile, err := util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

	if err != nil {
		return "", err
	}

	if !kth.VerifyChecksum(targetFile, FFX_KERNEL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts file handler")
	}

	kth.kernelTextAppFile = targetFile

	return kth.kernelTextAppFile, nil
}

func (kth *kernelTextHandler) ffx2KernelTexthandler() (string, error) {
	targetFile, err := util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX2_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

	if err != nil {
		return "", err
	}

	if !kth.VerifyChecksum(targetFile, FFX2_KERNEL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts file handler")
	}

	kth.kernelTextAppFile = targetFile

	return kth.kernelTextAppFile, nil
}

func (kth *kernelTextHandler) Dispose() {
	if kth.kernelTextAppFile != "" {
		if err := common.RemoveFileWithRetries(kth.kernelTextAppFile, 5); err != nil {
			fmt.Println("error when removing texts file handler: %w", err)
		}
	}
}
