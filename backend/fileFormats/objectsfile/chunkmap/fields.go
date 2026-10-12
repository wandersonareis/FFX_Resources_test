package chunkmap

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"ffxresources/backend/models"
)

// Kind describes the role a Go struct field plays in a mapped chunk.
type Kind uint8

const (
	KindText Kind = iota
	KindData
)

// Field describes one mapped binary value. Name is the Go struct field path
// for display only; Path and Component identify text without name lookups.
type Field struct {
	Name      string
	Offset    int
	Size      int
	Kind      Kind
	Path      []int
	Component int
	Part      models.TextPart
	SameAs    models.TextPart
}

type fieldsResult struct {
	fields []Field
	err    error
}

var (
	fieldsCache      sync.Map // reflect.Type -> fieldsResult
	textExporterType = reflect.TypeOf((*models.TextSegmentsExporter)(nil)).Elem()
	textImporterType = reflect.TypeOf((*models.TextSegmentsImporter)(nil)).Elem()
)

// Fields returns all fields in binary order. Text fields are identified only
// through the explicit models.TextSegmentsExporter/Importer contracts.
func Fields[T any]() ([]Field, error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if cached, ok := fieldsCache.Load(t); ok {
		r := cached.(fieldsResult)
		if r.err != nil {
			return nil, r.err
		}
		return slices.Clone(r.fields), nil
	}
	r := fieldsResult{}
	r.fields, r.err = collectFields(t)
	fieldsCache.Store(t, r)
	if r.err != nil {
		return nil, r.err
	}
	return slices.Clone(r.fields), nil
}

// SegmentFields returns only text references, in their binary order.
func SegmentFields[T any]() ([]Field, error) {
	all, err := Fields[T]()
	if err != nil {
		return nil, err
	}
	text := make([]Field, 0, len(all))
	for _, f := range all {
		if f.Kind == KindText {
			text = append(text, f)
		}
	}
	return text, nil
}

func collectFields(t reflect.Type) ([]Field, error) {
	c := &collector{seen: make(map[string]struct{})}
	if err := c.walkStruct(t, 0, "", nil); err != nil {
		return nil, err
	}
	return c.fields, nil
}

type collector struct {
	fields []Field
	seen   map[string]struct{}
}

func (c *collector) add(name string, kind Kind, offset, size int, path []int, component int, part, sameAs models.TextPart) error {
	if name == "" {
		return fmt.Errorf("campo sem nome: struct embutida sem nome nao gera caminho")
	}
	if _, exists := c.seen[name]; exists {
		return fmt.Errorf("caminho de campo %q repetido", name)
	}
	c.seen[name] = struct{}{}
	c.fields = append(c.fields, Field{
		Name: name, Offset: offset, Size: size, Kind: kind,
		Path: path, Component: component, Part: part, SameAs: sameAs,
	})
	return nil
}

func (c *collector) walkStruct(rt reflect.Type, base int, prefix string, path []int) error {
	offset := base
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			return fmt.Errorf("%s.%s: campo deve ser exportado", rt.Name(), field.Name)
		}
		size := typeSize(field.Type)
		if size < 0 {
			return fmt.Errorf("%s.%s: tipo %s nao suportado por encoding/binary", rt.Name(), field.Name, field.Type)
		}

		name := prefix
		if !flattens(field) {
			name = joinFieldPath(name, field.Name)
		}
		if err := c.walkValue(field.Type, offset, name, appendStep(path, i)); err != nil {
			return fmt.Errorf("%s.%s: %w", rt.Name(), field.Name, err)
		}
		offset += size
	}
	return nil
}

func (c *collector) walkValue(rt reflect.Type, base int, name string, path []int) error {
	if isTextField(rt) {
		value := reflect.New(rt).Elem().Interface().(models.TextSegmentsExporter).ExportTextSegments()
		segmentSize := typeSize(reflect.TypeOf(models.Segment{}))
		if len(value) == 0 || len(value)*segmentSize != typeSize(rt) {
			return fmt.Errorf("%s: text field %s declara %d segmentos para %d bytes", name, rt, len(value), typeSize(rt))
		}
		for i, text := range value {
			fieldName := name
			if text.Part != "" {
				fieldName += "." + string(text.Part)
			}
			if err := c.add(fieldName, KindText, base+i*segmentSize, segmentSize, path, i, text.Part, text.SameAs); err != nil {
				return err
			}
		}
		return nil
	}

	switch rt.Kind() {
	case reflect.Struct:
		return c.walkStruct(rt, base, name, path)
	case reflect.Array:
		elementSize := typeSize(rt.Elem())
		if elementSize < 0 {
			return fmt.Errorf("array de %s nao suportado", rt.Elem())
		}
		for i := 0; i < rt.Len(); i++ {
			if err := c.walkValue(rt.Elem(), base+i*elementSize,
				fmt.Sprintf("%s[%d]", name, i), appendStep(path, -(i+1))); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		return fmt.Errorf("slice nao suportado (use array fixo ou Tail)")
	default:
		return c.add(name, KindData, base, typeSize(rt), path, 0, "", "")
	}
}

func isTextField(rt reflect.Type) bool {
	return rt.Implements(textExporterType) && reflect.PointerTo(rt).Implements(textImporterType)
}

func flattens(field reflect.StructField) bool {
	return field.Anonymous && field.Type.Kind() == reflect.Struct && !isTextField(field.Type)
}

func joinFieldPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func appendStep(path []int, step int) []int {
	out := make([]int, len(path), len(path)+1)
	copy(out, path)
	return append(out, step)
}

func typeSize(rt reflect.Type) int {
	return binary.Size(reflect.New(rt).Elem().Interface())
}

// samePath reports whether two adjacent text components belong to the same
// schema field (for example Standard and Simplified of one TextPair).
func samePath(a, b []int) bool {
	return slices.Equal(a, b)
}
