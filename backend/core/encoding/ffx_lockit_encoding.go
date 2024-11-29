package ffxencoding

import (
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"os"
)

type IFFXTextLockitEncoding interface {
	GetFFXTextLockitEncoding() string
	GetFFXTextLockitLocalizationEncoding() string
	GetLockitFileHandler() encodingHandler.ILockitEncodingHandler
	Dispose()
}

type FFXTextLockitEncoding struct {
	charsEncoding string
	locEncoding   string
	lockitHandler encodingHandler.ILockitEncodingHandler
}

func newFFXTextLockitEncoding(locEncoding, charsEncoding string) *FFXTextLockitEncoding {
	return &FFXTextLockitEncoding{
		lockitHandler: encodingHandler.NewLockitHandler(),
		charsEncoding: charsEncoding,
		locEncoding:   locEncoding,
	}
}

func (e *FFXTextLockitEncoding) GetFFXTextLockitEncoding() string {
	return e.charsEncoding
}

func (e *FFXTextLockitEncoding) GetFFXTextLockitLocalizationEncoding() string {
	return e.locEncoding
}

func (e *FFXTextLockitEncoding) GetLockitFileHandler() encodingHandler.ILockitEncodingHandler {
	return e.lockitHandler
}

func (e *FFXTextLockitEncoding) Dispose() {
	os.Remove(e.charsEncoding)
	os.Remove(e.locEncoding)

	e.charsEncoding = ""
	e.locEncoding = ""

	e.lockitHandler.Dispose()
}
