package lockitencoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"fmt"
)

type LockitDecoder struct {}

func NewDecoder() *LockitDecoder {
	return &LockitDecoder{}
}

func (ld *LockitDecoder) LockitDecoderLoc(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	executable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (ld *LockitDecoder) LockitDecoderFfx(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitEncoding()

	executable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *LockitDecoder) decoder(executable, sourceFile, targetFile string, encoding string) error {
	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file does not exist")
	}	

	if !common.IsFileExists(encoding) {
		return fmt.Errorf("encoding file does not exist")
	}

	args := []string{"-t", encoding, sourceFile, targetFile}

	if err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
