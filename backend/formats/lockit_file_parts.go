package formats

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
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
	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

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
		lib.NotifyError(err)
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
		lib.NotifyError(err)
		return
	}
}
