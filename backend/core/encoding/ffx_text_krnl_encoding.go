package ffxencoding

import (
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"ffxresources/backend/interactions"
	"os"
)

type IFFXTextKrnlEncoding interface {
	GetEncoding() string
	GetKrnlHandler() encodingHandler.IKrnlEncodingHandler
	Dispose()
}

type ffxTextKrnlEncoding struct {
	encoding     string
	textsHandler encodingHandler.IKrnlEncodingHandler
}

func newFFXTextKrnlEncoding(encoding string) *ffxTextKrnlEncoding {
	gamePart := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	return &ffxTextKrnlEncoding{
		textsHandler: encodingHandler.NewKrnlTextsHandler(gamePart),
		encoding:     encoding,
	}
}

func (e *ffxTextKrnlEncoding) GetEncoding() string {
	return e.encoding
}

func (e *ffxTextKrnlEncoding) GetKrnlHandler() encodingHandler.IKrnlEncodingHandler {
	return e.textsHandler
}

func (e *ffxTextKrnlEncoding) Dispose() {
	_ = os.Remove(e.encoding)

	e.encoding = ""

	e.textsHandler.Dispose()
}
