package chunkmap

import (
	"fmt"
	"reflect"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/models"
)

// Formatter formats a mapped object for debug output.
type Formatter[T any] func(o *MappedTextObject[T], languageCode string) string

// MappedTextObject keeps the decoded bytes and text slots in binary order.
// Text field names are presentation metadata; slot indexes and reflect paths
// are the only identities used internally.
type MappedTextObject[T any] struct {
	headerLength int
	version      common.GameVersion
	typeName     string
	languageCode string
	chunk        *Chunk[T]
	fields       []Field
	slots        []*TextSlot
	initial      []string
	formatter    Formatter[T]
}

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
		return nil, fmt.Errorf("%s: mapear campos textuais: %w", typeName, err)
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
		slots:        make([]*TextSlot, len(fields)),
		initial:      make([]string, len(fields)),
	}
	value := reflect.ValueOf(&chunk.Value).Elem()
	for i, field := range fields {
		fieldValue := segmentAt(value, field.Path)
		exporter, ok := fieldValue.Interface().(models.TextSegmentsExporter)
		if !ok {
			return nil, fmt.Errorf("%s: campo %s nao exporta segmentos de texto", typeName, field.Name)
		}
		segments := exporter.ExportTextSegments()
		if field.Component < 0 || field.Component >= len(segments) {
			return nil, fmt.Errorf("%s: componente textual invalido em %s", typeName, field.Name)
		}
		slot := newTextSlot()
		slot.read(languageCode, stringBytes, segments[field.Component].Segment, version)
		o.slots[i] = slot
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

// segmentAt follows a reflect path (negative steps represent array indexes).
func segmentAt(value reflect.Value, path []int) reflect.Value {
	for _, step := range path {
		if step < 0 {
			value = value.Index(-step - 1)
		} else {
			value = value.Field(step)
		}
	}
	return value
}

func (o *MappedTextObject[T]) SetFormatter(fn Formatter[T]) { o.formatter = fn }

// Value expõe a struct decodificada do chunk — leitura de dados não textuais
// (como os model_ids de w_name.bin). Edições aqui entram no Encode.
func (o *MappedTextObject[T]) Value() *T { return &o.chunk.Value }

// OrderedFieldNames returns Go struct field paths in binary order.
func (o *MappedTextObject[T]) OrderedFieldNames() []string {
	names := make([]string, len(o.fields))
	for i, field := range o.fields {
		names[i] = field.Name
	}
	return names
}

// FieldOffsets returns each textual struct field path and its byte offset.
func (o *MappedTextObject[T]) FieldOffsets() ([]string, []int) {
	names := make([]string, len(o.fields))
	offsets := make([]int, len(o.fields))
	for i, field := range o.fields {
		names[i], offsets[i] = field.Name, field.Offset
	}
	return names, offsets
}

func (o *MappedTextObject[T]) ExportText(languageCode string) ([]TextField, error) {
	if err := validateLanguage(languageCode); err != nil {
		return nil, err
	}
	var out []TextField
	for start := 0; start < len(o.fields); {
		end := start + 1
		for end < len(o.fields) && samePath(o.fields[start].Path, o.fields[end].Path) {
			end++
		}

		partTexts := make(map[models.TextPart]string, end-start)
		for i := start; i < end; i++ {
			partTexts[o.fields[i].Part] = o.slots[i].GetLocalizedString(languageCode)
		}
		for i := start; i < end; i++ {
			text := partTexts[o.fields[i].Part]
			if text == "" || (o.fields[i].SameAs != "" && partTexts[o.fields[i].SameAs] == text) {
				continue
			}
			out = append(out, TextField{
				Index: i, Name: o.fields[i].Name,
				Texts: map[string]string{languageCode: text},
			})
		}
		start = end
	}
	return out, nil
}

func (o *MappedTextObject[T]) ImportText(fields []TextField) error {
	// Validate the entire input before mutating any slot.
	seen := make(map[int]struct{}, len(fields))
	byIndex := make(map[int]map[string]string, len(fields))
	for _, field := range fields {
		if field.Index < 0 || field.Index >= len(o.slots) {
			return fmt.Errorf("chunkmap: indice de campo textual fora da faixa: %d", field.Index)
		}
		if field.Name != "" && field.Name != o.fields[field.Index].Name {
			return fmt.Errorf("chunkmap: campo %d e %q, nao %q", field.Index, field.Name, o.fields[field.Index].Name)
		}
		if _, duplicate := seen[field.Index]; duplicate {
			return fmt.Errorf("chunkmap: campo textual %d repetido", field.Index)
		}
		seen[field.Index] = struct{}{}
		byIndex[field.Index] = field.Texts
		for languageCode := range field.Texts {
			if err := validateLanguage(languageCode); err != nil {
				return err
			}
		}
	}

	// Text field types declare SameAs relationships. If a related component is
	// omitted, the supplied source translation is applied to it as well.
	for start := 0; start < len(o.fields); {
		end := start + 1
		for end < len(o.fields) && samePath(o.fields[start].Path, o.fields[end].Path) {
			end++
		}
		for targetIndex := start; targetIndex < end; targetIndex++ {
			relation := o.fields[targetIndex].SameAs
			if relation == "" {
				continue
			}
			sourceIndex := -1
			for i := start; i < end; i++ {
				if o.fields[i].Part == relation {
					sourceIndex = i
					break
				}
			}
			if sourceIndex < 0 {
				return fmt.Errorf("%s: campo %s referencia parte textual ausente %q", o.typeName, o.fields[targetIndex].Name, relation)
			}
			if sourceTexts, supplied := byIndex[sourceIndex]; supplied {
				targetTexts := make(map[string]string, len(sourceTexts)+len(byIndex[targetIndex]))
				for languageCode, text := range sourceTexts {
					targetTexts[languageCode] = text
				}
				for languageCode, text := range byIndex[targetIndex] {
					if text != "" {
						targetTexts[languageCode] = text
					}
				}
				byIndex[targetIndex] = targetTexts
			}
		}
		start = end
	}

	updatesByIndex := make([][]textUpdate, len(o.slots))
	for index := range o.slots {
		texts, supplied := byIndex[index]
		if !supplied {
			continue
		}
		updates, err := o.slots[index].prepareTexts(texts, o.version)
		if err != nil {
			return fmt.Errorf("%s: campo %s: %w", o.typeName, o.fields[index].Name, err)
		}
		updatesByIndex[index] = updates
	}
	for index, updates := range updatesByIndex {
		o.slots[index].applyTexts(updates)
	}
	return nil
}

func (o *MappedTextObject[T]) SetLocalizations(other *MappedTextObject[T]) {
	if other == nil {
		return
	}
	for i := 0; i < min(len(o.slots), len(other.slots)); i++ {
		for localization, content := range other.slots[i].contents {
			if localization != o.languageCode {
				o.slots[i].contents[localization] = content
			}
		}
	}
}

func (o *MappedTextObject[T]) GetHeaderLength() int { return o.headerLength }

func (o *MappedTextObject[T]) ToBytes(languageCode string) ([]byte, error) {
	if err := validateLanguage(languageCode); err != nil {
		return nil, err
	}
	value := reflect.ValueOf(&o.chunk.Value).Elem()
	for start := 0; start < len(o.fields); {
		end := start + 1
		for end < len(o.fields) && samePath(o.fields[start].Path, o.fields[end].Path) {
			end++
		}

		field := o.fields[start]
		fieldValue := segmentAt(value, field.Path)
		exporter, ok := fieldValue.Interface().(models.TextSegmentsExporter)
		if !ok {
			return nil, fmt.Errorf("%s: campo %s nao exporta segmentos de texto", o.typeName, field.Name)
		}
		segments := exporter.ExportTextSegments()
		for i := start; i < end; i++ {
			component := o.fields[i].Component
			if component < 0 || component >= len(segments) {
				return nil, fmt.Errorf("%s: componente textual invalido em %s", o.typeName, o.fields[i].Name)
			}
			if content := o.slots[i].get(languageCode); content != nil {
				segments[component].Segment = content.Segment
			}
		}
		importer, ok := fieldValue.Addr().Interface().(models.TextSegmentsImporter)
		if !ok {
			return nil, fmt.Errorf("%s: campo %s nao importa segmentos de texto", o.typeName, field.Name)
		}
		if err := importer.ImportTextSegments(segments); err != nil {
			return nil, fmt.Errorf("%s: campo %s: %w", o.typeName, field.Name, err)
		}
		start = end
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

func (o *MappedTextObject[T]) String() string { return o.ToString(common.DefaultLocalization) }

func (o *MappedTextObject[T]) defaultString(languageCode string) string {
	parts := make([]string, 0, len(o.slots))
	for _, slot := range o.slots {
		if text := slot.GetLocalizedString(languageCode); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " | ")
}
