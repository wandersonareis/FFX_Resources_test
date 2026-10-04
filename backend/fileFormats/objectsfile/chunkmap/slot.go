package chunkmap

import (
	"bytes"
	"encoding/binary"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/models"
)

/*
Slot de texto do fluxo novo: TextContent + TextSlot substituem, dentro do
chunkmap, o par KeyedString/LocalizedKeyedStringObject do fluxo legado.

Os dois tipos implementam os contratos do datastore (IGlobalKeyedString e
IGlobalLocalizedKeyedStringObject) para que as funções compartilhadas da
camada Wails (builders/objectsfile.ExportFieldTexts, builders.ApplyObjectsEntry,
formatters JSON/Strings) continuem operando sem mudança — os contratos são a
camada orquestradora, não o fluxo legado. Nenhuma função do fluxo legado é
chamada: a resolução e a codificação de texto passam direto pelo converter,
que é o primitivo compartilhado de charset.
*/

// TextContent é o conteúdo de texto de um idioma num slot: o models.Segment
// (offset na string table + key) e os bytes codificados no charset do idioma.
type TextContent struct {
	Charset string
	Version common.GameVersion
	Segment models.Segment
	Bytes   []byte
	// Status é a classificação do Segment.Offset contra a tabela lida — a
	// mesma guarda do fluxo legado (objectsfile.KeyedString). Refs quebrados
	// ficam SEM texto e mantêm o segmento original no save.
	Status objectsfile.PointerStatus
	// edited marca texto trocado depois da leitura (SetString).
	edited bool
}

// PointerStatus devolve a classificação do ref lido (contrato tableRef do
// objectsfile — é o que permite o rebuild compartilhar a mesma guarda).
func (c *TextContent) PointerStatus() objectsfile.PointerStatus { return c.Status }

// StoredBytes devolve os bytes ARMAZENADOS do ref, sem terminador.
func (c *TextContent) StoredBytes() []byte { return c.Bytes }

// Edited reporta se o texto do ref foi trocado depois da leitura.
func (c *TextContent) Edited() bool { return c.edited }

func (c *TextContent) GetOffset() models.Offset { return c.Segment.Offset }
func (c *TextContent) SetOffset(offset models.Offset) {
	if offset != c.Segment.Offset {
		c.Segment.Offset = offset
	}
}
func (c *TextContent) GetKey() models.Key { return c.Segment.Key }
func (c *TextContent) SetKey(key models.Key) {
	if key != c.Segment.Key {
		c.Segment.Key = key
	}
}

func (c *TextContent) GetCharset() string {
	if c.Charset == "" {
		return common.DefaultLocalization
	}
	return c.Charset
}

func (c *TextContent) SetCharset(charset string) {
	if charset != "" && charset != c.Charset {
		c.Charset = charset
	}
}

// GetString decodifica os bytes no charset do idioma.
func (c *TextContent) GetString() string {
	return converter.BytesToString(c.Bytes, c.GetCharset(), c.Version)
}

// SetString codifica str no charset e substitui os bytes.
func (c *TextContent) SetString(str, newCharset string) {
	if newCharset != "" && newCharset != c.Charset {
		c.Charset = newCharset
	}
	encoded, err := converter.StringToBytes(str, c.Charset, c.Version)
	if err != nil {
		common.LogError("TextContent.SetString: %v", err)
		return
	}
	c.Bytes = encoded
	c.edited = true
}

func (c *TextContent) IsEmpty() bool { return c.GetString() == "" }
func (c *TextContent) String() string {
	return c.GetString()
}

// GetHeaderBytes escreve o segmento (offset+key, 4 bytes LE) no buffer.
func (c *TextContent) GetHeaderBytes(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, uint16(c.Segment.Offset))
	_ = binary.Write(buf, binary.LittleEndian, uint16(c.Segment.Key))
}

// SetHeaderBytes grava o segmento no buffer — segmento zerado não escreve.
func (c *TextContent) SetHeaderBytes(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if c.Segment.Offset == 0 && c.Segment.Key == 0 {
		return
	}
	c.GetHeaderBytes(buf)
}

// readTextContent resolve o texto do segmento na string table. Segmento
// zerado (offset==0 && key==0) não tem texto: devolve nil (os bytes originais
// do segmento são preservados — mesma regra do fluxo legado).
//
// O ponteiro é DIREÇÃO, não fato: fora de uma fronteira de bloco (ou fora
// da tabela) não existe texto ali — o sufixo não vira frase. O segmento
// original é mantido e o campo expõe "".
func readTextContent(strtab []byte, seg models.Segment, languageCode string, version common.GameVersion) *TextContent {
	if strtab == nil || (seg.Offset == 0 && seg.Key == 0) {
		return nil
	}
	status := objectsfile.PointerStatusAt(strtab, int(seg.Offset))
	content := &TextContent{
		Charset: ffxencoding.GetCharsetForLanguage(languageCode),
		Version: version,
		Segment: seg,
		Status:  status,
	}
	if status == objectsfile.PointerOK {
		content.Bytes = converter.GetStringBytesAtLookupOffset(strtab, int(seg.Offset))
	}
	return content
}

// TextSlot é o campo de texto de um chunk mapeado: um conteúdo por idioma.
// A chave identifica o campo nos DTOs (snake_case do nome do campo da struct).
type TextSlot struct {
	key      string
	contents map[string]datastore.IGlobalKeyedString
}

// newTextSlot cria o slot vazio para a chave.
func newTextSlot(key string) *TextSlot {
	return &TextSlot{key: key, contents: make(map[string]datastore.IGlobalKeyedString)}
}

// Key devolve a chave (DTO) do campo.
func (s *TextSlot) Key() string { return s.key }

// ReadAndSetLocalizedContent resolve o texto do segmento contra strtab no
// idioma (contrato datastore). É o caminho da CARGA: grava direto, sem
// passar pela guarda de sobrescrita.
func (s *TextSlot) ReadAndSetLocalizedContent(languageCode string, bytes []byte, offset models.Offset, key models.Key, version common.GameVersion) {
	if bytes == nil {
		return
	}
	if content := readTextContent(bytes, models.Segment{Offset: offset, Key: key}, languageCode, version); content != nil {
		s.contents[languageCode] = content
	}
}

// SetLocalizedContent grava o conteúdo do idioma (contrato datastore) — o
// caminho das EDIÇÕES pela ponte compartilhada (builders). Conteúdo vazio
// NUNCA sobrescreve um conteúdo existente (mesma regra do fluxo legado).
func (s *TextSlot) SetLocalizedContent(languageCode string, content datastore.IGlobalKeyedString) {
	if content == nil {
		return
	}
	if _, ok := s.contents[languageCode]; ok && content.IsEmpty() {
		return
	}
	s.contents[languageCode] = content
}

func (s *TextSlot) GetLocalizedContent(languageCode string) datastore.IGlobalKeyedString {
	return s.contents[languageCode]
}

func (s *TextSlot) GetLocalizedString(languageCode string) string {
	if c := s.GetLocalizedContent(languageCode); c != nil {
		return c.GetString()
	}
	return ""
}

func (s *TextSlot) GetDefaultContent() datastore.IGlobalKeyedString {
	return s.GetLocalizedContent(common.DefaultLocalization)
}

func (s *TextSlot) GetDefaultString() string {
	return s.GetLocalizedString(common.DefaultLocalization)
}

// CopyInto copia todos os conteúdos para outro objeto (contrato datastore —
// é o caminho do SetLocalizations entre objetos).
func (s *TextSlot) CopyInto(other datastore.IGlobalLocalizedKeyedStringObject) {
	for loc, content := range s.contents {
		other.SetLocalizedContent(loc, content)
	}
}

func (s *TextSlot) String() string {
	return s.GetDefaultString()
}

// SetTexts grava os textos por idioma no slot — o equivalente próprio do
// fluxo novo ao applyLocalizedText/updateOrCreateSegment do legado. Texto
// vazio e idioma não suportado são ignorados; texto igual ao atual não
// re-codifica.
func (s *TextSlot) SetTexts(texts map[string]string, version common.GameVersion) {
	for languageCode, newText := range texts {
		if newText == "" {
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			continue
		}
		if newText == s.GetLocalizedString(languageCode) {
			continue
		}
		charset := common.LanguageCodeToCharset(languageCode)
		if content := s.GetLocalizedContent(languageCode); content != nil {
			content.SetString(newText, charset)
			continue
		}
		encoded, err := converter.StringToBytes(newText, charset, version)
		if err != nil {
			common.LogError("TextSlot.SetTexts(%s): %v", s.key, err)
			continue
		}
		s.SetLocalizedContent(languageCode, &TextContent{
			Charset: charset,
			Version: version,
			Bytes:   encoded,
		})
	}
}
