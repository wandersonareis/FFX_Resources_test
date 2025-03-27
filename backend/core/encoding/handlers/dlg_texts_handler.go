package encodingHandler

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/models"
	"fmt"
	"os"
)

type IDlgEncodingHandler interface {
	GetDlgHandlerApp() (string, error)
	Dispose()
}

type dlgEncodingHandler struct {
	util.Checksum
	dlgFileType models.NodeType
	handlerFile string
}

func NewDlgTextsHandler(dlgFileType models.NodeType) IDlgEncodingHandler {
	return &dlgEncodingHandler{
		Checksum:    util.Checksum{},
		dlgFileType: dlgFileType,
	}
}

// TODO: Separate the special and standard dialog file handling
func (th *dlgEncodingHandler) GetDlgHandlerApp() (string, error) {
	if th.dlgFileType == models.DialogsSpecial {
		return th.getSpecialHandler()
	}

	targetFile, err := util.GetFromResources(DIALOG_HANDLER_RESOURCES_DIR, DIALOG_V2_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !th.VerifyChecksum(targetFile, DIALOG_V2_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts file handler")
	}

	th.handlerFile = targetFile

	return th.handlerFile, nil
}

func (th *dlgEncodingHandler) getSpecialHandler() (string, error) {
	targetFile, err := util.GetFromResources(DIALOG_HANDLER_RESOURCES_DIR, DIALOG_SPECIAL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !th.VerifyChecksum(targetFile, DIALOG_SPECIAL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for texts special file handler")
	}

	th.handlerFile = targetFile

	return th.handlerFile, nil
}

func (th *dlgEncodingHandler) Dispose() {
	if th.handlerFile != "" {
		if err := os.Remove(th.handlerFile); err != nil {
			fmt.Println("error when removing texts file handler")
		}
	}
}
