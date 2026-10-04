package chunkmap

import (
	"fmt"
	"reflect"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/models"
)

// Formatter formata o objeto para debug/log (usado por ToString).
type Formatter[T any] func(o *MappedTextObject[T], languageCode string) string

// MappedTextObject adapta um chunk mapeado pela struct T ao contrato
// datastore.IGlobalLocalizedTextObject — o contrato compartilhado da camada
// Wails (builders/formatters), não o fluxo legado. A struct é quem declara o
// que é texto (campos TextRef/TextPair) e quais são as chaves — derivadas do
// nome do campo, que é o do struct C#.
//
//   - Leitura: models.Segment de cada campo -> TextSlot (slot.go), que
//     resolve o texto contra a string table pelo converter.
//   - Escrita (ToBytes): conteúdo de volta no campo do segmento + struct +
//     Tail (sem patch dos bytes originais).
//
// Campos não-texto são SOMENTE leitura: nunca são serializados de volta
// exceto como os bytes exatos lidos (struct preserva o valor decodificado).
type MappedTextObject[T any] struct {
	headerLength int
	version      common.GameVersion
	typeName     string
	// languageCode é o idioma resolvido na carga — a mesma chave do snapshot.
	languageCode string

	chunk  *Chunk[T]
	fields []Field
	langs  []string
	slots  []*TextSlot
	byName map[string]*TextSlot
	// initial é o snapshot do texto (idioma da carga) por slot — a base da
	// detecção de edição do ToBytes do arquivo.
	initial   []string
	formatter Formatter[T]
}

// garante o contrato de texto (9 métodos) para qualquer T.
var _ datastore.IGlobalLocalizedTextObject = (*MappedTextObject[models.Segment])(nil)

// NewMappedTextObject decodifica chunkBytes em T e instancia um slot de
// texto por campo, resolvendo o texto contra stringBytes no idioma.
func NewMappedTextObject[T any](
	chunkBytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version common.GameVersion,
	typeName string,
) (*MappedTextObject[T], error) {
	fields, err := SegmentFields[T]()
	if err != nil {
		return nil, fmt.Errorf("%s: mapear campos: %w", typeName, err)
	}
	chunk, err := Decode[T](chunkBytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", typeName, err)
	}

	o := &MappedTextObject[T]{
		headerLength: headerLength,
		version:      version,
		typeName:     typeName,
		languageCode: languageCode,
		chunk:        chunk,
		fields:       fields,
		langs:        supportedLangs(),
		slots:        make([]*TextSlot, len(fields)),
		byName:       make(map[string]*TextSlot, len(fields)),
		initial:      make([]string, len(fields)),
	}

	value := reflect.ValueOf(&chunk.Value).Elem()
	for i, f := range fields {
		seg := segmentAt(value, f.Path).Interface().(models.Segment)
		slot := newTextSlot(f.Key)
		slot.ReadAndSetLocalizedContent(languageCode, stringBytes, seg.Offset, seg.Key, version)
		o.slots[i] = slot
		o.byName[f.Key] = slot
		o.initial[i] = slot.GetLocalizedString(languageCode)
	}
	return o, nil
}

// Edited reporta se algum slot mudou o texto no idioma da carga desde
// NewMappedTextObject — a detecção de edição usada pelo ToBytes do arquivo
// (cobre qualquer caminho de escrita, inclusive a ponte dos builders).
func (o *MappedTextObject[T]) Edited() bool {
	for i, slot := range o.slots {
		if slot.GetLocalizedString(o.languageCode) != o.initial[i] {
			return true
		}
	}
	return false
}

// segmentAt segue o caminho (negativo = índice de array) até o models.Segment.
func segmentAt(v reflect.Value, path []int) reflect.Value {
	for _, step := range path {
		if step < 0 {
			v = v.Index(-step - 1)
		} else {
			v = v.Field(step)
		}
	}
	return v
}

// SetFormatter define um formatter customizado para ToString.
func (o *MappedTextObject[T]) SetFormatter(fn Formatter[T]) {
	o.formatter = fn
}

// OrderedFieldKeys devolve as chaves na ordem do arquivo (ordem binária).
func (o *MappedTextObject[T]) OrderedFieldKeys() []string {
	keys := make([]string, 0, len(o.fields))
	for _, f := range o.fields {
		keys = append(keys, f.Key)
	}
	return keys
}

// FieldOffsets devolve (chave, offset) por campo de texto — para inspeção
// e validação contra o cabeçalho/binário.
func (o *MappedTextObject[T]) FieldOffsets() ([]string, []int) {
	keys := make([]string, len(o.fields))
	offsets := make([]int, len(o.fields))
	for i, f := range o.fields {
		keys[i] = f.Key
		offsets[i] = f.Offset
	}
	return keys, offsets
}

// GetName retorna o campo "name" se existir; senão o primeiro campo do
// arquivo (o primeiro campo é o título natural do chunk — `name`, ou o
// campo equivalente como `help` no btl_txt.bin).
func (o *MappedTextObject[T]) GetName(languageCode string) string {
	if slot := o.byName["name"]; slot != nil {
		return slot.GetLocalizedString(languageCode)
	}
	if len(o.slots) > 0 {
		return o.slots[0].GetLocalizedString(languageCode)
	}
	return ""
}

// GetKeyedString retorna o segmento pelo título (chave JSON).
func (o *MappedTextObject[T]) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	return o.byName[title]
}

// Texts implementa TextResolver: leitura do campo `key` na string table —
// os textos não-vazios por idioma (nil quando o campo não existe).
func (o *MappedTextObject[T]) Texts(key string) map[string]string {
	slot := o.byName[key]
	if slot == nil {
		return nil
	}
	var out map[string]string
	for _, loc := range o.langs {
		if text := slot.GetLocalizedString(loc); text != "" {
			if out == nil {
				out = make(map[string]string)
			}
			out[loc] = text
		}
	}
	return out
}

// SetTexts implementa TextResolver: grava os textos no slot de `key` e
// devolve o models.Segment resultante (o que ToBytes escreve na struct).
// A gravação é própria do fluxo novo (TextSlot.SetTexts) — codificação via
// converter, sem passar pela ponte do fluxo legado.
func (o *MappedTextObject[T]) SetTexts(key string, texts map[string]string) models.Segment {
	slot := o.byName[key]
	if slot == nil || len(texts) == 0 {
		return models.Segment{}
	}
	slot.SetTexts(texts, o.version)
	for _, loc := range o.langs {
		if c := slot.GetLocalizedContent(loc); c != nil {
			return models.Segment{Offset: c.GetOffset(), Key: c.GetKey()}
		}
	}
	return models.Segment{}
}

// ExportText devolve um FieldText por campo de texto da struct, na ordem
// binária do arquivo — o helper genérico (chunkmap.ExportText) sobre o
// próprio objeto como resolver.
func (o *MappedTextObject[T]) ExportText() ([]objectsfile.FieldText, error) {
	return ExportText[T](o)
}

// ImportText aplica os campos recebidos (DTO -> slot); chave desconhecida é
// erro (pega typo no JSON).
func (o *MappedTextObject[T]) ImportText(fields []objectsfile.FieldText) error {
	return ImportText[T](fields, o)
}

// GetLocalizedKeyedStrings retorna os conteúdos localizados na ordem do arquivo.
func (o *MappedTextObject[T]) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	result := make([]datastore.IGlobalKeyedString, 0, len(o.slots))
	for _, slot := range o.slots {
		if c := slot.GetLocalizedContent(localization); c != nil {
			result = append(result, c)
		}
	}
	return result
}

// SetLocalizations copia as localizações de outro objeto por interseção de
// chaves. Aceita *MappedTextObject[T] (mescla de idiomas no LoadFile) e o
// *KeyedStringFile legado: ambos expõem GetKeyedString(string).
func (o *MappedTextObject[T]) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	getter, ok := other.(interface {
		GetKeyedString(string) datastore.IGlobalLocalizedKeyedStringObject
	})
	if !ok {
		return
	}
	for _, f := range o.fields {
		slot := o.byName[f.Key]
		if slot == nil {
			continue
		}
		if otherSeg := getter.GetKeyedString(f.Key); otherSeg != nil {
			otherSeg.CopyInto(slot)
		}
	}
}

// GetTextObject retorna o próprio objeto.
func (o *MappedTextObject[T]) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return o
}

// GetHeaderLength retorna o comprimento do cabeçalho do arquivo.
func (o *MappedTextObject[T]) GetHeaderLength() int {
	return o.headerLength
}

// ToBytes grava cada conteúdo de volta no models.Segment correspondente e
// serializa struct + Tail. Slot sem conteúdo no idioma é pulado (bytes
// originais do segmento preservados).
func (o *MappedTextObject[T]) ToBytes(languageCode string) ([]byte, error) {
	value := reflect.ValueOf(&o.chunk.Value).Elem()
	for i, f := range o.fields {
		content := o.slots[i].GetLocalizedContent(languageCode)
		if content == nil {
			continue
		}
		segmentAt(value, f.Path).Set(reflect.ValueOf(models.Segment{
			Offset: content.GetOffset(),
			Key:    content.GetKey(),
		}))
	}
	out, err := o.chunk.Encode()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", o.typeName, err)
	}
	return out, nil
}

// ToString formata para debug/log: formatter customizado ou defaultString.
func (o *MappedTextObject[T]) ToString(languageCode string) string {
	if o.formatter != nil {
		return o.formatter(o, languageCode)
	}
	return o.defaultString(languageCode)
}

// String implementa IGlobalLocalizedTextObject.
func (o *MappedTextObject[T]) String() string {
	return o.ToString(common.DefaultLocalization)
}

// defaultString une os valores não-vazios na ordem do arquivo com " | ".
func (o *MappedTextObject[T]) defaultString(languageCode string) string {
	parts := make([]string, 0, len(o.slots))
	for _, slot := range o.slots {
		if s := slot.GetLocalizedString(languageCode); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " | ")
}
