package lockitFileEncoder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"fmt"
)

type LockitEncoderUTF8Strategy struct{}

func NewLockitEncoderUTF8Strategy() *LockitEncoderUTF8Strategy {
	return &LockitEncoderUTF8Strategy{}
}

func (le *LockitEncoderUTF8Strategy) Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	lockitExecutable, err := encoding.GetLockitFileHandler().GetLockitFileHandler()
	if err != nil {
		return err
	}

	if err := encoder(lockitExecutable, sourceFile, outputFile, encodingFile); err != nil {
		return err
	}

	utf8BomExecutable, err := encoding.GetLockitFileHandler().GetLockitFileUtf8BomNormalizer()
	if err != nil {
		return err
	}

	if _, err := normalizeBom(utf8BomExecutable, outputFile); err != nil {
		return err
	}

	return nil
}

type LockitEncoderFFXStrategy struct{}

func NewLockitEncoderFFXStrategy() *LockitEncoderFFXStrategy {
	return &LockitEncoderFFXStrategy{}
}

func (le *LockitEncoderFFXStrategy) Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	return fmt.Errorf("v1 encoder not implemented")
}

type LockitEncoderFFX2Strategy struct{}

func NewLockitEncoderFFX2Strategy() *LockitEncoderFFX2Strategy {
	return &LockitEncoderFFX2Strategy{}
}

func (le *LockitEncoderFFX2Strategy) Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitEncoding()

	executable, err := encoding.GetLockitFileHandler().GetLockitFileHandler()
	if err != nil {
		return err
	}

	if err := encoder(executable, sourceFile, outputFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func encoder(executable, sourceFile, outputFile string, encoding string) error {
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

	if _, err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}

func normalizeBom(executable, target string) (string, error) {
	args := []string{"-r", target}

	return components.RunCommand(executable, args)
}
