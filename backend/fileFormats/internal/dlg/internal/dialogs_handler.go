package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/models"
	"fmt"
	"os"
)

type dialogsHandler struct {
	util.Checksum
	targetFile string
	fileType   models.NodeType
}

func newDialogsHandler(fileType models.NodeType) *dialogsHandler {
	return &dialogsHandler{
		Checksum: util.Checksum{},
		fileType: fileType,
	}
}

func (d *dialogsHandler) getDialogsHandler() (string, error) {
	if d.fileType == models.DialogsSpecial {
		return d.getSpecialHandler()
	}

	return d.getHandler()
}
func (d *dialogsHandler) getHandler() (string, error) {
	targetFile, err := util.GetFromResources(util.DIALOG_HANDLER_RESOURCES_DIR, util.DIALOG_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	d.SetChecksumString(util.DIALOG_HANDLER_SHA256)
	
	if !d.IsValid(targetFile) {
		return "", fmt.Errorf("invalid checksum for dialog file handler")
	}

	d.targetFile = targetFile

	return d.targetFile, nil
}

func (d *dialogsHandler) getSpecialHandler() (string, error) {
	targetFile, err := util.GetFromResources(util.DIALOG_HANDLER_RESOURCES_DIR, util.DIALOG_SPECIAL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	d.SetChecksumString(util.DIALOG_SPECIAL_HANDLER_SHA256)

	if !d.IsValid(targetFile) {
		return "", fmt.Errorf("invalid checksum for dialog special file handler")
	}

	d.targetFile = targetFile

	return d.targetFile, nil
}

func (d *dialogsHandler) dispose() {
	if d.targetFile != "" {
		os.Remove(d.targetFile)
		d.targetFile = ""
	}
}
