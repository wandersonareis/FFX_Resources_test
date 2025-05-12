package lockitParts

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
	"path/filepath"
)

type (
	LockitFileParts struct {
		baseFormats.IBaseFileFormat

		lockitEncodingService lockitFileEncoder.ILockitEncodingService
		logger                loggingService.ILoggerService
	}

	LockitEncodingType int
)

const (
	FFXEncoding LockitEncodingType = iota
	UTF8Encoding
)

func NewLockitFileParts(source interfaces.ISource, destination locations.IDestination) *LockitFileParts {
	relativePath := filepath.Join(util.LOCKIT_TARGET_DIR_NAME, source.GetNameWithoutExtension())
	source.SetRelativePath(relativePath)
	return &LockitFileParts{
		IBaseFileFormat:       baseFormats.NewFormatsBase(source, destination),
		lockitEncodingService: lockitFileEncoder.NewLockitEncodingService(),
		logger:                loggingService.NewLoggerHandler("lockit_file_parts"),
	}
}

func (l *LockitFileParts) Extract(encoding ffxencoding.IFFXTextLockitEncoding, encodingStrategy lockitFileEncoder.ILockitProcessingStrategy) error {
	sourceFile := l.GetSource().GetPath()
	targetFile := l.GetDestination().Extract().GetTargetFile()

	if err := l.lockitEncodingService.Process(sourceFile, targetFile, encoding, encodingStrategy); err != nil {
		return fmt.Errorf("error when extracting lockit file parts: %w", err)
	}

	return nil
}

func (l *LockitFileParts) Compress(encoding ffxencoding.IFFXTextLockitEncoding, encodingStrategy lockitFileEncoder.ILockitProcessingStrategy) error {
	translatedTextFile := l.GetDestination().Translate().GetTargetFile()
	outputTranslatedBinary := common.RemoveOneFileExtension(translatedTextFile)

	if err := common.CheckPathExists(translatedTextFile); err != nil {
		return err
	}

	if err := l.lockitEncodingService.Process(translatedTextFile, outputTranslatedBinary, encoding, encodingStrategy); err != nil {
		return fmt.Errorf("error when compressing lockit file parts: %w", err)
	}

	return nil
}
