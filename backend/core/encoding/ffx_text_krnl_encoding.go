package ffxencoding

import (
	"ffxresources/backend/core/encoding/handlers"
	"ffxresources/backend/interactions"
	"os"
)

type IFFXTextKrnlEncoding interface {
	FetchEncoding() string
	FetchKrnlHandler() encodingHandler.IKrnlEncodingHandler
	Dispose()
}

type ffxTextKrnlEncoding struct {
	encoding     string
	textsHandler encodingHandler.IKrnlEncodingHandler
}

func newFFXTextKrnlEncoding(encoding string) *ffxTextKrnlEncoding {
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()
	return &ffxTextKrnlEncoding{
		textsHandler: encodingHandler.NewKrnlTextsHandler(gamePart),
		encoding:     encoding,
	}
}

func (e *ffxTextKrnlEncoding) FetchEncoding() string {
	return e.encoding
}

func (e *ffxTextKrnlEncoding) FetchKrnlHandler() encodingHandler.IKrnlEncodingHandler {
	return e.textsHandler
}

func (e *ffxTextKrnlEncoding) Dispose() {
	os.Remove(e.encoding)

	e.encoding = ""

	e.textsHandler.Dispose()
}
