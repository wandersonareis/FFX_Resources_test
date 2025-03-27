package textsEncoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"fmt"
)

type textsDecoder struct{}

func NewDecoder() *textsDecoder {
	return &textsDecoder{}
}

func (d *textsDecoder) DlgDecoder(sourceFile, targetFile string, encoding ffxencoding.IFFXTextDlgEncoding) error {
	encodingFile := encoding.GetEncoding()

	executable, err := encoding.GetDlgHandler().GetDlgHandlerApp()
	if err != nil {
		return err
	}

	if err := d.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *textsDecoder) KnrlDecoder(sourceFile, targetFile string, encoding ffxencoding.IFFXTextKrnlEncoding) error {
	encodingFile := encoding.GetEncoding()

	executable, err := encoding.GetKrnlHandler().GetKrnlHandlerApp()
	if err != nil {
		return err
	}

	if err := d.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *textsDecoder) decoder(executable, sourceFile, targetFile, encoding string) error {
	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file does not exist")
	}

	if !common.IsFileExists(encoding) {
		return fmt.Errorf("encoding file does not exist")
	}

	args := []string{"-e", "-t", encoding, sourceFile, targetFile}

	if _, err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
