package objectsfile

import (
	"fmt"
	"slices"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// SegmentField define um campo de segmento: nome (chave JSON) e gap.
// Gap = bytes pulados DEPOIS deste segmento (0 = contíguo, >0 = pula bytes após o segmento).
type SegmentField struct {
	Name string
	Gap  int
}

// LayoutSet mapeia versão -> lista de campos (ordem = ordem binária).
// A incompatibilidade entre versões é dado, não código.
type LayoutSet map[common.GameVersion][]SegmentField

// StringFormatter formata o objeto para debug/log (usado por ToString).
type StringFormatter func(f *KeyedStringFile, languageCode string) string

// KeyedStringFile é um wrapper genérico que substitui os tipos concretos
// que só variam no número de campos e no layout posicional dos segmentos.
type KeyedStringFile struct {
	Bytes        []byte
	HeaderLength int
	Version      common.GameVersion

	fields    []SegmentField
	offsets   []int
	segments  []datastore.IGlobalLocalizedKeyedStringObject
	byName    map[string]datastore.IGlobalLocalizedKeyedStringObject
	formatter StringFormatter
}

// NewKeyedStringFile cria o wrapper genérico a partir de um layout.
// typeName aparece apenas nas mensagens de erro.
// start é o offset absoluto do primeiro segmento (0 para contíguos,
// 4 para Dress/Command LastMission que pulam os primeiros 4 bytes).
func NewKeyedStringFile(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version common.GameVersion,
	layouts LayoutSet,
	typeName string,
) (*KeyedStringFile, error) {
	return NewKeyedStringFileAt(bytes, stringBytes, headerLength, languageCode, version, layouts, typeName, 0)
}

// NewKeyedStringFileAt é a variante com offset inicial explícito.
func NewKeyedStringFileAt(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version common.GameVersion,
	layouts LayoutSet,
	typeName string,
	start int,
) (*KeyedStringFile, error) {
	v := version.Normalize()
	fields, ok := layouts[v]
	if !ok {
		return nil, fmt.Errorf("%s is not compatible with game version %s", typeName, v)
	}
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create %s: have %d bytes, need at least %d", typeName, len(bytes), headerLength)
	}

	f := &KeyedStringFile{
		Bytes:        bytes,
		HeaderLength: headerLength,
		Version:      version,
		fields:       fields,
		segments:     make([]datastore.IGlobalLocalizedKeyedStringObject, len(fields)),
		byName:       make(map[string]datastore.IGlobalLocalizedKeyedStringObject, len(fields)),
	}

	// offsets calculados uma única vez, compartilhados entre leitura e escrita.
	f.offsets = segmentOffsets(start, gapsOf(fields), len(fields))

	for i := range fields {
		f.segments[i] = NewLocalizedKeyedStringObject()
		f.byName[fields[i].Name] = f.segments[i]
	}

	if err := readStringSegmentsAt(f.Bytes, f.offsets, stringBytes, languageCode, version, f.segments...); err != nil {
		return nil, fmt.Errorf("reading segments: %w", err)
	}
	return f, nil
}

// gapsOf converte os campos em uma lista de gaps compatível com segmentOffsets
// (gaps[i] = bytes pulados após o segmento i).
func gapsOf(fields []SegmentField) []int {
	if len(fields) == 0 {
		return nil
	}
	gaps := make([]int, len(fields))
	for i := range fields {
		// O último gap não tem efeito nas offsets; segmentOffsets ignora i >= len(gaps)-1.
		gaps[i] = fields[i].Gap
	}
	return gaps
}

// SetFormatter define um formatter customizado para ToString.
func (f *KeyedStringFile) SetFormatter(fn StringFormatter) {
	f.formatter = fn
}

// FormatWith formata usando um formatter ad-hoc sem alterar o estado.
func (f *KeyedStringFile) FormatWith(languageCode string, fn StringFormatter) string {
	if fn != nil {
		return fn(f, languageCode)
	}
	return f.defaultString(languageCode)
}

// GetName retorna o campo "name" se existir; senão o primeiro campo não-vazio.
func (f *KeyedStringFile) GetName(languageCode string) string {
	if seg := f.byName["name"]; seg != nil {
		return seg.GetLocalizedString(languageCode)
	}
	for _, seg := range f.segments {
		if s := seg.GetLocalizedString(languageCode); s != "" {
			return s
		}
	}
	return ""
}

// GetKeyedString retorna o segmento pelo título.
func (f *KeyedStringFile) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	return f.byName[title]
}

// GetLocalizedKeyedStrings retorna os conteúdos localizados na ordem do layout.
func (f *KeyedStringFile) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	result := make([]datastore.IGlobalKeyedString, 0, len(f.segments))
	for _, seg := range f.segments {
		if c := seg.GetLocalizedContent(localization); c != nil {
			result = append(result, c)
		}
	}
	return result
}

// GetHeaderLength retorna o comprimento do cabeçalho.
func (f *KeyedStringFile) GetHeaderLength() int {
	return f.HeaderLength
}

// GetTextObject retorna o próprio objeto.
func (f *KeyedStringFile) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return f
}

// SetLocalizations copia as localizações de outro objeto por interseção de nomes.
func (f *KeyedStringFile) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	o, ok := other.(*KeyedStringFile)
	if !ok {
		return
	}
	for name, seg := range f.byName {
		if otherSeg, exists := o.byName[name]; exists {
			otherSeg.CopyInto(seg)
		}
	}
}

// ToBytes serializa para binário usando os offsets pré-calculados.
func (f *KeyedStringFile) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(f.Bytes)
	if err := writeStringSegmentsAt(data, f.offsets, languageCode, f.segments...); err != nil {
		return nil, err
	}
	return data, nil
}

// ToString formata para debug/log: usa formatter customizado ou defaultString.
func (f *KeyedStringFile) ToString(languageCode string) string {
	if f.formatter != nil {
		return f.formatter(f, languageCode)
	}
	return f.defaultString(languageCode)
}

// String implementa IGlobalLocalizedTextObject.
func (f *KeyedStringFile) String() string {
	return f.ToString(common.DefaultLocalization)
}

// defaultString une os valores não-vazios na ordem do layout com " | ".
// É o fallback sempre que nenhum formatter foi registrado.
func (f *KeyedStringFile) defaultString(languageCode string) string {
	parts := make([]string, 0, len(f.segments))
	for _, seg := range f.segments {
		if s := seg.GetLocalizedString(languageCode); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " | ")
}

// FieldString retorna o valor de um campo pelo nome (nil-safe, para formatters legados).
func FieldString(f *KeyedStringFile, name, languageCode string) string {
	if seg := f.byName[name]; seg != nil {
		return seg.GetLocalizedString(languageCode)
	}
	return ""
}