package internal

/* import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"time"
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
	targetFile, err := util.GetFromResources(DIALOG_HANDLER_RESOURCES_DIR, DIALOG_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !d.VerifyChecksum(targetFile, DIALOG_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for dialog file handler")
	}

	d.targetFile = targetFile

	return d.targetFile, nil
}

func (d *dialogsHandler) getSpecialHandler() (string, error) {
	targetFile, err := util.GetFromResources(DIALOG_HANDLER_RESOURCES_DIR, DIALOG_SPECIAL_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !d.VerifyChecksum(targetFile, DIALOG_SPECIAL_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for dialog special file handler")
	}

	d.targetFile = targetFile

	return d.targetFile, nil
}

func (d *dialogsHandler) dispose() {
	if d.targetFile != "" {
		time.Sleep(2 * time.Second)
		os.Remove(d.targetFile)
		d.targetFile = ""
	}
}
 */