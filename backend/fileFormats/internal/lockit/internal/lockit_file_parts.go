package internal

import (
	"ffxresources/backend/fileFormats/internal/base"
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
	var err error

	switch dec {
	case FfxEnc:
		err = lockitDecoderFfx(l.GetFileInfo())
	case LocEnc:
		err = lockitDecoderLoc(l.GetFileInfo())
	default:
		err = fmt.Errorf("invalid encode type: %d", dec)
	}

	if err != nil {
		l.Log.Error().Err(err).Msg("error when decoding lockit file")
		return
	}
}

func (l *LockitFileParts) Compress(enc LockitPartEncodeType) {
	var err error

	switch enc {
	case FfxEnc:
		err = lockitEncoderFfx(l.GetFileInfo())
	case LocEnc:
		err = lockitEncoderLoc(l.GetFileInfo())
	default:
		err = fmt.Errorf("invalid encode type: %d", enc)
	}

	if err != nil {
		l.Log.Error().Err(err).Msg("error when encoding lockit file")
		return
	}
}
