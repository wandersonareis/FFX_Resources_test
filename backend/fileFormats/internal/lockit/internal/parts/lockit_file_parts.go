package parts

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/encoding"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

type LockitFileParts struct {
	*base.FormatsBase
}

type LockitPartEncodeType int

const (
	FfxEnc LockitPartEncodeType = iota
	LocEnc
)

func NewLockitFileParts(dataInfo interactions.IGameDataInfo) *LockitFileParts {
	gData := dataInfo.GetGameData()
	gData.RelativeGameDataPath = filepath.Join(util.LOCKIT_TARGET_DIR_NAME, dataInfo.GetGameData().NamePrefix)
	dataInfo.SetGameData(gData)

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &LockitFileParts{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (l *LockitFileParts) Extract(dec LockitPartEncodeType, encoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, 1)

	decoder := lockitencoding.NewDecoder()

	switch dec {
	case FfxEnc:
		errChan <- decoder.LockitDecoderFfx(l.GetGameData().FullFilePath, l.GetExtractLocation().TargetFile, encoding)
	case LocEnc:
		errChan <- decoder.LockitDecoderLoc(l.GetGameData().FullFilePath, l.GetExtractLocation().TargetFile, encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", dec)
	}

	err := <-errChan
	if err != nil {
		l.Log.Error().Err(err).Msgf("error when decoding lockit file: %s", l.GetExtractLocation().TargetFile)
		return
	}
}

func (l *LockitFileParts) Compress(enc LockitPartEncodeType, encoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, 1)

	encoder := lockitencoding.NewEncoder()

	l.GetImportLocation().ProvideTargetPath()

	switch enc {
	case FfxEnc:
		errChan <- encoder.LockitEncoderFfx(l.GetTranslateLocation().TargetFile, l.GetImportLocation().TargetFile, encoding)
	case LocEnc:
		errChan <- encoder.LockitEncoderLoc(l.GetTranslateLocation().TargetFile, l.GetImportLocation().TargetFile, encoding)
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", enc)
	}

	err := <-errChan
	if err != nil {
		l.Log.Error().Err(err).Msgf("error when encoding lockit file: %s", l.GetTranslateLocation().TargetFile)
		return
	}
}
