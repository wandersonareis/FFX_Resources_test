package lockitParts

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"path/filepath"
)

type (
	LockitFileParts struct {
		baseFormats.IBaseFileFormat
		decoder *lockitFileEncoder.LockitDecoder
		encoder *lockitFileEncoder.LockitEncoder
		logger  logger.ILoggerHandler
	}

	LockitEncodingType int
)

const (
	FFXEncoding LockitEncodingType = iota
	UTF8Encoding
)

func NewLockitFileParts(source interfaces.ISource, destination locations.IDestination) *LockitFileParts {
	source.Get().RelativePath = filepath.Join(util.LOCKIT_TARGET_DIR_NAME, source.Get().NamePrefix)

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &LockitFileParts{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
		decoder:         lockitFileEncoder.NewDecoder(),
		encoder:         lockitFileEncoder.NewEncoder(),
		logger:          logger.NewLoggerHandler("lockit_file_parts"),
	}
}

func (l *LockitFileParts) Extract(dec LockitEncodingType, encoding ffxencoding.IFFXTextLockitEncoding) error {
	errChan := make(chan error, 1)
	defer close(errChan)

	switch dec {
	case FFXEncoding:
		errChan <- l.decoder.LockitDecoderFfx(l.GetSource().Get().Path, l.GetDestination().Extract().GetTargetFile(), encoding)
	case UTF8Encoding:
		errChan <- l.decoder.LockitDecoderLoc(l.GetSource().Get().Path, l.GetDestination().Extract().GetTargetFile(), encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", dec)
	}

	if err := <-errChan; err != nil {
		return fmt.Errorf("error when extracting lockit file parts: %w", err)
	}
	return nil
}

func (l *LockitFileParts) Compress(enc LockitEncodingType, encoding ffxencoding.IFFXTextLockitEncoding) error {
	translatedTextFile := l.GetDestination().Translate().GetTargetFile()
	outputTranslatedBinary := common.RemoveOneFileExtension(translatedTextFile)

	if err := common.CheckPathExists(translatedTextFile); err != nil {
		return err
	}

	switch enc {
	case FFXEncoding:
		if err := l.encoder.LockitEncoderFfx(translatedTextFile, outputTranslatedBinary, encoding); err != nil {
			return err
		}
	case UTF8Encoding:
		if err := l.encoder.LockitEncoderLoc(translatedTextFile, outputTranslatedBinary, encoding); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid encode type: %d", enc)
	}

	return nil
}
