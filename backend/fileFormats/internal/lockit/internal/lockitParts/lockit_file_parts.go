package lockitParts

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
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
		*base.FormatsBase
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
		FormatsBase: base.NewFormatsBase(source, destination),
		decoder:     lockitFileEncoder.NewDecoder(),
		encoder:     lockitFileEncoder.NewEncoder(),
		logger:      logger.NewLoggerHandler("lockit_file_parts"),
	}
}

func (l *LockitFileParts) Extract(dec LockitEncodingType, encoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, 1)
	defer close(errChan)

	switch dec {
	case FFXEncoding:
		errChan <- l.decoder.LockitDecoderFfx(l.Source().Get().Path, l.Destination().Extract().Get().GetTargetFile(), encoding)
	case UTF8Encoding:
		errChan <- l.decoder.LockitDecoderLoc(l.Source().Get().Path, l.Destination().Extract().Get().GetTargetFile(), encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", dec)
	}

	if err := <-errChan; err != nil {
		l.logger.LogError(err, "error when extracting lockit file parts")
	}

}

func (l *LockitFileParts) Compress(enc LockitEncodingType, encoding ffxencoding.IFFXTextLockitEncoding, errChan chan error) {
	translatedTextFile := l.Destination().Translate().Get().GetTargetFile()
	outputTranslatedBinary := common.RemoveOneFileExtension(translatedTextFile)

	switch enc {
	case FFXEncoding:
		errChan <- l.encoder.LockitEncoderFfx(translatedTextFile, outputTranslatedBinary, encoding)
	case UTF8Encoding:
		errChan <- l.encoder.LockitEncoderLoc(translatedTextFile, outputTranslatedBinary, encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", enc)
	}
}
