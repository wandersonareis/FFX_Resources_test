package lockitFileParts

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/encoding"
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
		logger.ILoggerHandler
	}
	LockitPartEncodeType int
)

const (
	FfxEnc LockitPartEncodeType = iota
	LocEnc
)

func NewLockitFileParts(source interfaces.ISource, destination locations.IDestination) *LockitFileParts {
	gData := source.Get()
	gData.RelativePath = filepath.Join(util.LOCKIT_TARGET_DIR_NAME, gData.NamePrefix)
	source.Set(gData)

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &LockitFileParts{
		FormatsBase: base.NewFormatsBase(source, destination),
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "lockit_file_parts").Logger(),
		},
	}
}

func (l *LockitFileParts) Extract(dec LockitPartEncodeType, encoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, 1)
	defer close(errChan)

	decoder := lockitencoding.NewDecoder()

	switch dec {
	case FfxEnc:
		errChan <- decoder.LockitDecoderFfx(l.Source().Get().Path, l.Destination().Extract().Get().GetTargetFile(), encoding)
	case LocEnc:
		errChan <- decoder.LockitDecoderLoc(l.Source().Get().Path, l.Destination().Extract().Get().GetTargetFile(), encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", dec)
	}

	select {
	case err := <-errChan:
		l.LogError(err, "error when extracting lockit file parts")
	}

}

func (l *LockitFileParts) Compress(enc LockitPartEncodeType, encoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, 1)
	defer close(errChan)

	encoder := lockitencoding.NewEncoder()

	l.Destination().Import().Get().ProvideTargetPath()

	switch enc {
	case FfxEnc:
		errChan <- encoder.LockitEncoderFfx(l.Destination().Translate().Get().GetTargetFile(), l.Destination().Import().Get().GetTargetFile(), encoding)
	case LocEnc:
		errChan <- encoder.LockitEncoderLoc(l.Destination().Translate().Get().GetTargetFile(), l.Destination().Import().Get().GetTargetFile(), encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", enc)
	}

	select {
	case err := <-errChan:
		l.LogError(err, "error when compressing lockit file parts")
	}
}
