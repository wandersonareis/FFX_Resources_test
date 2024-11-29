package ffxencoding

import (
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"ffxresources/backend/models"
	"os"
)

type IFFXTextDlgEncoding interface {
	FetchEncoding() string
	FetchDlgHandler() encodingHandler.IDlgEncodingHandler
	Dispose()
}

type ffxTextDlgEncoding struct {
	encoding     string
	textsHandler encodingHandler.IDlgEncodingHandler
}

func newFFXTextDlgEncoding(encoding string, textsType models.NodeType) *ffxTextDlgEncoding {
	return &ffxTextDlgEncoding{
		textsHandler: encodingHandler.NewDlgTextsHandler(textsType),
		encoding:     encoding,
	}
}

func (e *ffxTextDlgEncoding) FetchEncoding() string {
	return e.encoding
}

func (e *ffxTextDlgEncoding) FetchDlgHandler() encodingHandler.IDlgEncodingHandler {
	return e.textsHandler
}

func (e *ffxTextDlgEncoding) Dispose() {
	os.Remove(e.encoding)

	e.encoding = ""

	e.textsHandler.Dispose()
}
