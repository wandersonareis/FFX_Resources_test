package lockitencoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/lib"
	"fmt"
)

type LockitEncoder struct {}

func NewEncoder() *LockitEncoder {
	return &LockitEncoder{}
}

func (ld *LockitEncoder) LockitEncoderLoc(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	lockitExecutable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.encoder(lockitExecutable, sourceFile, outputFile, encodingFile); err != nil {
		return err
	}

	utf8BomExecutable, err := encoding.GetLockitFileHandler().FetchLockitUtf8BomNormalizer()
	if err != nil {
		return err
	}

	return ensureUtf8Bom(utf8BomExecutable, outputFile)
}

func (ld *LockitEncoder) LockitEncoderFfx(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitEncoding()

	executable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.encoder(executable, sourceFile, outputFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *LockitEncoder) encoder(executable, sourceFile, outputFile string, encoding string) error {
	if !common.IsFileExists(executable) {
		return fmt.Errorf("executable does not exist")
	}

	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file does not exist")
	}

	if !common.IsFileExists(encoding) {
		return fmt.Errorf("encoding file does not exist")
	}

	args := []string{"-tr", encoding, sourceFile, outputFile}

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}

func ensureUtf8Bom(executable, target string) error {
	args := []string{"-r", target}

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}