package ffxencoding

import (
	"ffxresources/backend/common"
	encodingHandler "ffxresources/backend/core/encoding/handlers"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
)

type IFFXTextDlgEncoding interface {
	GetEncoding() string
	GetDlgHandler() encodingHandler.IDlgEncodingHandler
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

func (e *ffxTextDlgEncoding) GetEncoding() string {
	return e.encoding
}

func (e *ffxTextDlgEncoding) GetDlgHandler() encodingHandler.IDlgEncodingHandler {
	return e.textsHandler
}

func (e *ffxTextDlgEncoding) Dispose() {
	if err := common.RemoveFileWithRetries(e.encoding, 3); err != nil {
		l := loggingService.Get().With().Str("module", "ffx_text_dlg_encoding").Logger()
		l.Error().Err(err).Str("file", e.encoding).Msg("Error on removing encoding file")
		return
	}

	e.encoding = ""

	e.textsHandler.Dispose()
}
