package lockitFileEncoder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"fmt"
)

type LockitDecoderUTF8Strategy struct{}

func NewLockitDecoderUTF8Strategy() *LockitDecoderUTF8Strategy {
	return &LockitDecoderUTF8Strategy{}
}

func (ld *LockitDecoderUTF8Strategy) Process(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	executable, err := encoding.GetLockitFileHandler().GetLockitFileHandler()
	if err != nil {
		return err
	}

	if err := decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

type LockitDecoderFFXStrategy struct{}

func NewLockitDecoderFFXStrategy() *LockitDecoderFFXStrategy {
	return &LockitDecoderFFXStrategy{}
}

func (ld *LockitDecoderFFXStrategy) Process(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	return fmt.Errorf("v1 decoder not implemented")
}

type LockitDecoderFFX2Strategy struct{}

func NewLockitDecoderFFX2Strategy() *LockitDecoderFFX2Strategy {
	return &LockitDecoderFFX2Strategy{}
}

func (ld *LockitDecoderFFX2Strategy) Process(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitEncoding()

	executable, err := encoding.GetLockitFileHandler().GetLockitFileHandler()
	if err != nil {
		return err
	}

	if err := decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func decoder(executable, sourceFile, targetFile string, encodingFile string) error {
	if err := common.CheckPathExists(executable); err != nil {
		return fmt.Errorf("executable file not found: %s", executable)
	}

	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file not found: %s", sourceFile)
	}

	if err := common.CheckPathExists(encodingFile); err != nil {
		return fmt.Errorf("encoding file not found: %s", encodingFile)
	}

	args := []string{"-t", encodingFile, sourceFile, targetFile}

	if _, err := components.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
