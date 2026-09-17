package objectsfile

import (
	"fmt"
	"slices"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// Cópias com sufixo Store do caminho genérico de keyed strings.
// Espelham byte-a-byte a lógica usada pelos Read* (start sempre 0,
// offsets = 4 bytes sequenciais + gap após cada segmento).
// Os originais em keyed_string_file.go / data_objects_binary_reader.go
// estão congelados como oráculo; este arquivo pertence ao store novo.

// StringFormatterStore formata um KeyedStringFileStore para debug/log.
type StringFormatterStore func(f *KeyedStringFileStore, languageCode string) string

// KeyedStringFileStore é a cópia Store de KeyedStringFile.
type KeyedStringFileStore struct {
	Bytes        []byte
	HeaderLength int
	Version      common.GameVersion

	fields    []SegmentField
	offsets   []int
	segments  []datastore.IGlobalLocalizedKeyedStringObject
	byName    map[string]datastore.IGlobalLocalizedKeyedStringObject
	formatter StringFormatterStore
}

// NewKeyedStringFileStore cria o wrapper genérico a partir de um layout.
// O primeiro segmento sempre começa em 0; saltos são expressos via Gap
// (0 = contíguo, >0 = bytes pulados após o segmento, ex. 0x1C, 0xA4).
func NewKeyedStringFileStore(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version common.GameVersion,
	layouts LayoutSet,
	typeName string,
) (*KeyedStringFileStore, error) {
	fields, ok := layouts[version]
	if !ok {
		return nil, fmt.Errorf("%s is not compatible with game version %s", typeName, version)
	}
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create %s: have %d bytes, need at least %d", typeName, len(bytes), headerLength)
	}

	f := &KeyedStringFileStore{
		Bytes:        bytes,
		HeaderLength: headerLength,
		Version:      version,
		fields:       fields,
		segments:     make([]datastore.IGlobalLocalizedKeyedStringObject, len(fields)),
		byName:       make(map[string]datastore.IGlobalLocalizedKeyedStringObject, len(fields)),
	}

	f.offsets = segmentOffsetsStore(gapsOfStore(fields), len(fields))

	for i := range fields {
		f.segments[i] = NewLocalizedKeyedStringObject()
		f.byName[fields[i].Name] = f.segments[i]
	}

	if err := readStringSegmentsAtStore(f.Bytes, f.offsets, stringBytes, languageCode, version, f.segments...); err != nil {
		return nil, fmt.Errorf("reading segments: %w", err)
	}
	return f, nil
}

// gapsOfStore converte os campos em gaps compatíveis com segmentOffsetsStore.
func gapsOfStore(fields []SegmentField) []int {
	if len(fields) == 0 {
		return nil
	}
	gaps := make([]int, len(fields))
	for i := range fields {
		gaps[i] = fields[i].Gap
	}
	return gaps
}

// SetFormatter define um formatter customizado para ToString.
func (f *KeyedStringFileStore) SetFormatter(fn StringFormatterStore) {
	f.formatter = fn
}

// FormatWith formata usando um formatter ad-hoc sem alterar o estado.
func (f *KeyedStringFileStore) FormatWith(languageCode string, fn StringFormatterStore) string {
	if fn != nil {
		return fn(f, languageCode)
	}
	return f.defaultString(languageCode)
}

// GetName retorna o campo "name" se existir; senão o primeiro campo não-vazio.
func (f *KeyedStringFileStore) GetName(languageCode string) string {
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
func (f *KeyedStringFileStore) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	return f.byName[title]
}

// GetLocalizedKeyedStrings retorna os conteúdos localizados na ordem do layout.
func (f *KeyedStringFileStore) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	result := make([]datastore.IGlobalKeyedString, 0, len(f.segments))
	for _, seg := range f.segments {
		if c := seg.GetLocalizedContent(localization); c != nil {
			result = append(result, c)
		}
	}
	return result
}

// GetHeaderLength retorna o comprimento do cabeçalho.
func (f *KeyedStringFileStore) GetHeaderLength() int {
	return f.HeaderLength
}

// GetTextObject retorna o próprio objeto.
func (f *KeyedStringFileStore) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return f
}

// SetLocalizations copia as localizações de outro objeto por interseção de nomes.
func (f *KeyedStringFileStore) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	o, ok := other.(*KeyedStringFileStore)
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
func (f *KeyedStringFileStore) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(f.Bytes)
	if err := writeStringSegmentsAtStore(data, f.offsets, languageCode, f.segments...); err != nil {
		return nil, err
	}
	return data, nil
}

// ToString formata para debug/log: usa formatter customizado ou defaultString.
func (f *KeyedStringFileStore) ToString(languageCode string) string {
	if f.formatter != nil {
		return f.formatter(f, languageCode)
	}
	return f.defaultString(languageCode)
}

// String implementa IGlobalLocalizedTextObject.
func (f *KeyedStringFileStore) String() string {
	return f.ToString(common.DefaultLocalization)
}

// defaultString une os valores não-vazios na ordem do layout com " | ".
func (f *KeyedStringFileStore) defaultString(languageCode string) string {
	parts := make([]string, 0, len(f.segments))
	for _, seg := range f.segments {
		if s := seg.GetLocalizedString(languageCode); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " | ")
}

// FieldStringStore retorna o valor de um campo pelo nome (nil-safe, para formatters).
func FieldStringStore(f *KeyedStringFileStore, name, languageCode string) string {
	if seg := f.byName[name]; seg != nil {
		return seg.GetLocalizedString(languageCode)
	}
	return ""
}

// Formatters do store (cópias dos legados, operando no tipo Store).
var (
	commandLegacyFmtStore = func(f *KeyedStringFileStore, lang string) string {
		return fmt.Sprintf("%s %s - %s %s",
			FieldStringStore(f, "name", lang),
			FieldStringStore(f, "simplifiedName", lang),
			FieldStringStore(f, "description", lang),
			FieldStringStore(f, "simplifiedDescription", lang))
	}

	nameOnlyLegacyFmtStore = func(f *KeyedStringFileStore, lang string) string {
		return fmt.Sprintf("%s %s",
			FieldStringStore(f, "name", lang),
			FieldStringStore(f, "simplifiedName", lang))
	}

	threePartLegacyFmtStore = func(f *KeyedStringFileStore, lang string) string {
		return fmt.Sprintf("%s - %s - %s",
			FieldStringStore(f, "name", lang),
			FieldStringStore(f, "description", lang),
			FieldStringStore(f, "effect", lang))
	}
)

// newKeyedStringStore cria a instância aplicando um formatter opcional.
// Espelha newKeyedString do caminho Read.
func newKeyedStringStore(bytes, stringBytes []byte, headerLength int, languageCode string, version common.GameVersion, layouts LayoutSet, typeName string, formatter StringFormatterStore) (*KeyedStringFileStore, error) {
	f, err := NewKeyedStringFileStore(bytes, stringBytes, headerLength, languageCode, version, layouts, typeName)
	if err != nil {
		return nil, err
	}
	if formatter != nil {
		f.SetFormatter(formatter)
	}
	return f, nil
}
