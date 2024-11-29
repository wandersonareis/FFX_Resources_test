package textsEncoding

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/lib"
	"fmt"
)

type textsDecoder struct{}

func NewDecoder() *textsDecoder {
	return &textsDecoder{}
}

func (d *textsDecoder) DlgDecoder(sourceFile, targetFile string, encoding ffxencoding.IFFXTextDlgEncoding) error {
	encodingFile := encoding.FetchEncoding()

	executable, err := encoding.FetchDlgHandler().FetchDlgTextsHandler()
	if err != nil {
		return err
	}

	if err := d.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *textsDecoder) KnrlDecoder(sourceFile, targetFile string, encoding ffxencoding.IFFXTextKrnlEncoding) error {
	encodingFile := encoding.FetchEncoding()

	executable, err := encoding.FetchKrnlHandler().FetchKrnlTextsHandler()
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

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}