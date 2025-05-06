package ffxencoding

import (
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"ffxresources/backend/interactions"
	"os"
)

type IFFXTextKrnlEncoding interface {
	GetEncodingFile() string
	GetKrnlHandler() encodingHandler.IKrnlEncodingHandler
	Dispose()
}

type ffxTextKrnlEncoding struct {
	fileEncoding string
	textsHandler encodingHandler.IKrnlEncodingHandler
}

func newFFXTextKrnlEncoding(encoding string) *ffxTextKrnlEncoding {
	gamePart := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	return &ffxTextKrnlEncoding{
		textsHandler: encodingHandler.NewKrnlTextsHandler(gamePart),
		fileEncoding: encoding,
	}
}

func (e *ffxTextKrnlEncoding) GetEncodingFile() string {
	return e.fileEncoding
}

func (e *ffxTextKrnlEncoding) GetKrnlHandler() encodingHandler.IKrnlEncodingHandler {
	return e.textsHandler
}

func (e *ffxTextKrnlEncoding) Dispose() {
	_ = os.Remove(e.fileEncoding)

	e.fileEncoding = ""

	e.textsHandler.Dispose()
}
