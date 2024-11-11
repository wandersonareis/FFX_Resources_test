package internal

import (
	"ffxresources/backend/events"
	"ffxresources/backend/formats/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

type LockitFileParts struct {
	dataInfo *interactions.GameDataInfo
}

type LockitPartEncodeType int

const (
	FfxEnc LockitPartEncodeType = iota
	LocEnc
)

func NewLockitFileParts(dataInfo *interactions.GameDataInfo) *LockitFileParts {
	dataInfo.GameData.RelativePath = filepath.Join(lib.LOCKIT_TARGET_DIR_NAME, dataInfo.GameData.NamePrefix)

	dataInfo.InitializeLocations(formatters.NewTxtFormatter())

	return &LockitFileParts{
		dataInfo: dataInfo,
	}
}

func (l LockitFileParts) GetFileInfo() *interactions.GameDataInfo {
	return l.dataInfo
}

func (l *LockitFileParts) Extract(enc LockitPartEncodeType) {
	var err error

	switch enc {
	case FfxEnc:
		err = lockitDecoderFfx(l.dataInfo)
	case LocEnc:
		err = lockitDecoderLoc(l.dataInfo)
	default:
		err = fmt.Errorf("invalid encode type: %d", enc)
	}

	if err != nil {
		events.NotifyError(err)
		return
	}
}

func (l *LockitFileParts) Compress(enc LockitPartEncodeType) {
	var err error

	switch enc {
	case FfxEnc:
		err = lockitEncoderFfx(l.dataInfo)
	case LocEnc:
		err = lockitEncoderLoc(l.dataInfo)
	default:
		err = fmt.Errorf("invalid encode type: %d", enc)
	}

	if err != nil {
		events.NotifyError(err)
		return
	}
}
