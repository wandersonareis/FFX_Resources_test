package ffxencoding

import encodingHandler "ffxresources/backend/core/encoding/handlers"

type FFXTextEncoding struct {
	LockitEncodingHandler ILockitEncodingHandler

	Encoding string
}

func NewFFXTextEncoding(encoding string) *FFXTextEncoding {
	return &FFXTextEncoding{
		LockitEncodingHandler: encodingHandler.NewLockitHandler(),
		Encoding: encoding,
	}
}

func (e *FFXTextEncoding) GetEncoding() string {
	return e.Encoding
}