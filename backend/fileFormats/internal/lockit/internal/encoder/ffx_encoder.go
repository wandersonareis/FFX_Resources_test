package lockitFileEncoder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"fmt"
)

type LockitEncoder struct{}

func NewEncoder() *LockitEncoder {
	return &LockitEncoder{}
}

func (le *LockitEncoder) LockitEncoderLoc(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	lockitExecutable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := le.encoder(lockitExecutable, sourceFile, outputFile, encodingFile); err != nil {
		return err
	}

	utf8BomExecutable, err := encoding.GetLockitFileHandler().FetchLockitUtf8BomNormalizer()
	if err != nil {
		return err
	}

	return le.normalizeBom(utf8BomExecutable, outputFile)
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

	if err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}

func (ld *LockitEncoder) normalizeBom(executable, target string) error {
	args := []string{"-r", target}

	return components.RunCommand(executable, args)
}
