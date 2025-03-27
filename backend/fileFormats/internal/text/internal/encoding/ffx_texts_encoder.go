package textsEncoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"fmt"
)

type textsEncoder struct{}

func NewEncoder() *textsEncoder {
	return &textsEncoder{}
}

func (d *textsEncoder) DlgEncoder(sourceFile, targetFile, outputFile string, encoding ffxencoding.IFFXTextDlgEncoding) error {
	encodingFile := encoding.GetEncoding()

	executable, err := encoding.GetDlgHandler().GetDlgHandlerApp()
	if err != nil {
		return err
	}

	if err := d.encoder(executable, sourceFile, targetFile, outputFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *textsEncoder) KnrlEncoder(sourceFile, targetFile, outputFile string, encoding ffxencoding.IFFXTextKrnlEncoding) error {
	encodingFile := encoding.GetEncoding()

	executable, err := encoding.GetKrnlHandler().GetKrnlHandlerApp()
	if err != nil {
		return err
	}

	if err := d.encoder(executable, sourceFile, targetFile, outputFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *textsEncoder) encoder(executable, sourceFile, targetFile, outputFile, encoding string) error {
	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file does not exist")
	}

	if !common.IsFileExists(targetFile) {
		return fmt.Errorf("target file does not exist")
	}

	if !common.IsFileExists(encoding) {
		return fmt.Errorf("encoding file does not exist")
	}

	args := []string{"-i", "-t", encoding, sourceFile, targetFile, outputFile}

	if _, err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
