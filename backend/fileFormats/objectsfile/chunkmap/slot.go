package chunkmap

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/strings"
	"ffxresources/backend/models"
)

// TextContent implements the low-level keyed-string contract consumed by the
// shared string-table compiler. TextSlot is new-flow-only localized storage.

// TextContent é o conteúdo de texto de um idioma num slot: o models.Segment
// (offset na string table + key) e os bytes codificados no charset do idioma.
type TextContent struct {
	Charset string
	Version common.GameVersion
	Segment models.Segment
	Bytes   []byte
	// Status é a classificação do Segment.Offset contra a tabela lida — a
	// validação comum da string table. Refs quebrados ficam sem texto e mantêm
	// o segmento original no save.
	Status strings.PointerStatus
	// edited marca texto trocado depois da leitura (SetString).
	edited bool
}

// PointerStatus devolve a classificação do ref lido (contrato tableRef do
// objectsfile — é o que permite o rebuild compartilhar a mesma guarda).
func (c *TextContent) PointerStatus() strings.PointerStatus { return c.Status }

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
// do segmento são preservados).
//
// O ponteiro é direção, não fato: fora de uma fronteira de bloco (ou fora da
// tabela) não existe texto ali. O segmento original é mantido e o campo expõe "".
func readTextContent(strtab []byte, seg models.Segment, languageCode string, version common.GameVersion) *TextContent {
	if strtab == nil || (seg.Offset == 0 && seg.Key == 0) {
		return nil
	}
	status := strings.PointerStatusAt(strtab, int(seg.Offset))
	content := &TextContent{
		Charset: ffxencoding.GetCharsetForLanguage(languageCode),
		Version: version,
		Segment: seg,
		Status:  status,
	}
	if status == strings.PointerOK {
		content.Bytes = converter.GetStringBytesAtLookupOffset(strtab, int(seg.Offset))
	}
	return content
}

// TextSlot stores the contents of one text reference by localization.
type TextSlot struct {
	contents map[string]*TextContent
}

func newTextSlot() *TextSlot {
	return &TextSlot{contents: make(map[string]*TextContent)}
}

func (s *TextSlot) read(languageCode string, table []byte, segment models.Segment, version common.GameVersion) {
	if content := readTextContent(table, segment, languageCode, version); content != nil {
		s.contents[languageCode] = content
	}
}

func (s *TextSlot) get(languageCode string) *TextContent { return s.contents[languageCode] }

func (s *TextSlot) GetLocalizedString(languageCode string) string {
	if content := s.contents[languageCode]; content != nil {
		return content.GetString()
	}
	return ""
}

// clone devolve um TextSlot com cópias rasas dos TextContent: edições no
// clone (SetString via copy em prepareTexts) não afetam o original.
func (s *TextSlot) clone() *TextSlot {
	c := newTextSlot()
	for lang, content := range s.contents {
		if content == nil {
			continue
		}
		copied := *content
		c.contents[lang] = &copied
	}
	return c
}

type textUpdate struct {
	languageCode string
	content      *TextContent
}

func (s *TextSlot) prepareTexts(texts map[string]string, version common.GameVersion) ([]textUpdate, error) {
	updates := make([]textUpdate, 0, len(texts))
	for languageCode, newText := range texts {
		if newText == "" || newText == s.GetLocalizedString(languageCode) {
			continue
		}
		charset := common.LanguageCodeToCharset(languageCode)
		encoded, err := converter.StringToBytes(newText, charset, version)
		if err != nil {
			return nil, fmt.Errorf("codificar idioma %s: %w", languageCode, err)
		}
		var content *TextContent
		if existing := s.contents[languageCode]; existing != nil {
			updated := *existing
			updated.Charset = charset
			updated.Bytes = encoded
			updated.edited = true
			content = &updated
		} else {
			content = &TextContent{
				Charset: charset,
				Version: version,
				Bytes:   encoded,
				Status:  strings.PointerCreated,
				edited:  true,
			}
		}
		updates = append(updates, textUpdate{languageCode: languageCode, content: content})
	}
	return updates, nil
}

func (s *TextSlot) applyTexts(updates []textUpdate) {
	for _, update := range updates {
		s.contents[update.languageCode] = update.content
	}
}
