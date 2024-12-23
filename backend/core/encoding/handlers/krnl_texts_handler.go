package encodingHandler

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/models"
	"fmt"
	"os"
)

type IKrnlEncodingHandler interface {
	FetchKrnlTextsHandler() (string, error)
	Dispose()
}

type krnlEncodingHandler struct {
	util.Checksum
	handlerFile string
	gamePart    models.GameVersion
}

func NewKrnlTextsHandler(gamePart models.GameVersion) IKrnlEncodingHandler {
	return &krnlEncodingHandler{
		Checksum: util.Checksum{},
		gamePart: gamePart,
	}
}

func (th *krnlEncodingHandler) FetchKrnlTextsHandler() (string, error) {
	switch th.gamePart {
	case models.FFX:
		return th.fetchFFXKrnlTextsHandler()
	case models.FFX2:
		return th.fetchFFX2KrnlTextsHandler()
	default:
		return "", fmt.Errorf("invalid game part")
	}
}

func (th *krnlEncodingHandler) fetchFFXKrnlTextsHandler() (string, error) {
	targetFile, err := util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

	if err != nil {
		return "", err
	}

	if !th.VerifyChecksum(targetFile, FFX_KERNEL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts file handler")
	}

	th.handlerFile = targetFile

	return th.handlerFile, nil
}

func (th *krnlEncodingHandler) fetchFFX2KrnlTextsHandler() (string, error) {
	targetFile, err := util.GetFromResources(KERNEL_HANDLER_RESOURCES_DIR, FFX2_KERNEL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)

	if err != nil {
		return "", err
	}

	if !th.VerifyChecksum(targetFile, FFX2_KERNEL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts file handler")
	}

	th.handlerFile = targetFile

	return th.handlerFile, nil
}

func (th *krnlEncodingHandler) Dispose() {
	if th.handlerFile != "" {
		if err := os.Remove(th.handlerFile); err != nil {
			fmt.Println("error when removing texts file handler")
		}
	}
}
