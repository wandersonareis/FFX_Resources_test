package parts

import (
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
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

func (l *LockitFileParts) Extract(dec LockitPartEncodeType) {
	errChan := make(chan error, 1)

	switch dec {
	case FfxEnc:
		errChan <- lib.LockitDecoderFfx(l.GetFileInfo())
	case LocEnc:
		errChan <- lib.LockitDecoderLoc(l.GetFileInfo())
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", dec)
	}

	err := <-errChan
	if err != nil {
		l.Log.Error().Err(err).Msgf("error when decoding lockit file: %s", l.GetExtractLocation().TargetFile)
		return
	}
}

func (l *LockitFileParts) Compress(enc LockitPartEncodeType) {
	errChan := make(chan error, 1)

	switch enc {
	case FfxEnc:
		errChan <- lib.LockitEncoderFfx(l.GetFileInfo())
	case LocEnc:
		errChan <- lib.LockitEncoderLoc(l.GetFileInfo())
	default:
		errChan <- fmt.Errorf("invalid encode type: %d", enc)
	}

	err := <-errChan
	if err != nil {
		l.Log.Error().Err(err).Msgf("error when encoding lockit file: %s", l.GetTranslateLocation().TargetFile)
		return
	}
}
