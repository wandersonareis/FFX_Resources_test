package ffxencoding

import (
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"os"
)

type (
	IFFXTextKrnlEncoding interface {
		GetEncodingFile() string
		GetKrnlHandler() encodingHandler.IKernelTextHandler
		Dispose()
	}

	ffxTextKrnlEncoding struct {
		fileEncoding string
		textsHandler encodingHandler.IKernelTextHandler
	}
)

func newFFXTextKrnlEncoding(encoding string) *ffxTextKrnlEncoding {
	return &ffxTextKrnlEncoding{
		textsHandler: encodingHandler.NewKrnlTextsHandler(),
		fileEncoding: encoding,
	}
}

func (e *ffxTextKrnlEncoding) GetEncodingFile() string {
	return e.fileEncoding
}

func (e *ffxTextKrnlEncoding) GetKrnlHandler() encodingHandler.IKernelTextHandler {
	return e.textsHandler
}

func (e *ffxTextKrnlEncoding) Dispose() {
	_ = os.Remove(e.fileEncoding)

	e.fileEncoding = ""

	e.textsHandler.Dispose()
}
