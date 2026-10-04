package chunkmap

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"ffxresources/backend/models"
)

// Kind diz o que um campo da struct representa para o fluxo de texto.
type Kind uint8

const (
	// KindText é uma referência para a string table (models.Segment).
	KindText Kind = iota
	// KindData é qualquer outro campo: hoje só é preservado byte-a-byte;
	// a exportação de dados fica para depois.
	KindData
)

// Field descreve um campo da struct T na ordem binária:
//
//   - Key: a chave de exportação, em snake_case, derivada do caminho do
//     campo — que é o nome do campo no struct C# (ver "Convenções" dos
//     pacotes schema/ffx e schema/ffx2);
//   - Offset/Size: posição e comprimento dentro do chunk;
//   - Kind: texto ou dado — a struct é quem sabe (o tipo do campo decide);
//   - Path: caminho de índice até o campo (get/set via reflect; negativo =
//     índice de array).
//
// O resultado é somente leitura (Path é compartilhado entre chamadas).
type Field struct {
	Key    string
	Offset int
	Size   int
	Kind   Kind
	Path   []int
}

// segmentType é o tipo folha de texto. models.Segment = {Offset u16, Key u16}.
var segmentType = reflect.TypeOf(models.Segment{})

type fieldsResult struct {
	fields []Field
	err    error
}

var fieldsCache sync.Map // reflect.Type -> fieldsResult

// Fields devolve, cacheado por tipo, TODOS os campos da struct T na ordem
// binária (a mesma ordem em que aparecem no arquivo), texto e dado.
//
// As chaves vêm do nome do campo — não de tag:
//
//   - models.Segment folha -> `snake(caminho)`;
//   - TextPair (Standard/Simplified) -> `{campo}` e `{campo}_simplified`;
//   - sub-struct com nome -> prefixo `{campo}_`; embutida anônima achata
//     (CommandBody.Anim1 -> `anim1`);
//   - array -> `_{indice}` 0-based;
//   - escalar -> dado, com a mesma regra de chave.
//
// Chave vazia ou repetida é erro (pegam colisão de nomes).
//
// O nome do campo é o identificador Go idiomático da grafia do struct C#
// (PascalCase, inicialismos em maiúsculas), de modo que snake(id) reproduz o
// nome do campo C#. Onde isso não é possível (anim_1, _u, damage_9999,
// typos do C#), o campo carrega um comentário `// C# ...` com a origem — a
// lista completa está em "Divergências de nome" nos docs dos pacotes
// schema/ffx e schema/ffx2.
//
// O resultado é somente leitura (Path compartilhado entre chamadas).
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

// SegmentFields devolve só os campos KindText de Fields, na ordem do arquivo.
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
	if err := walkStruct(t, 0, "", nil, c); err != nil {
		return nil, err
	}
	return c.out, nil
}

type collector struct {
	out  []Field
	seen map[string]struct{}
}

func (c *collector) add(key string, kind Kind, off, size int, path []int) error {
	if key == "" {
		return fmt.Errorf("campo sem nome: struct embutida sem nome nao gera chave")
	}
	if _, dup := c.seen[key]; dup {
		return fmt.Errorf("chave %q repetida", key)
	}
	c.seen[key] = struct{}{}
	c.out = append(c.out, Field{Key: key, Offset: off, Size: size, Kind: kind, Path: path})
	return nil
}

// walkStruct percorre os campos escopo a escopo acumulando offset.
// prefix é o caminho já resolvido até aqui ("" no topo, "creature_data_"
// dentro de um campo CreatureData).
func walkStruct(rt reflect.Type, base int, prefix string, path []int, c *collector) error {
	off := base
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			// binary.Read escreve via reflect; campo não-exportado quebraria
			// o Decode em runtime. Toda a struct precisa ser exportada.
			return fmt.Errorf("%s.%s: campo deve ser exportado", rt.Name(), f.Name)
		}
		size := typeSize(f.Type)
		if size < 0 {
			return fmt.Errorf("%s.%s: tipo %s nao suportado por encoding/binary", rt.Name(), f.Name, f.Type)
		}
		name := prefix
		if !flattens(f) {
			name += snake(f.Name)
		}
		if err := walkValue(f.Type, off, name, appendStep(path, i), c); err != nil {
			return fmt.Errorf("%s.%s: %w", rt.Name(), f.Name, err)
		}
		off += size
	}
	return nil
}

// flattens diz se o campo embutido anônimo só contribui com seus próprios
// campos (sem nomear a sub-árvore) — o caso do CommandBody de FFX.
// TextRef e TextPair embutidos continuam nomeados pelo campo que os declara.
func flattens(f reflect.StructField) bool {
	return f.Anonymous &&
		f.Type.Kind() == reflect.Struct &&
		f.Type != segmentType &&
		!isTextPair(f.Type)
}

// walkValue cobre folha de texto, par standard/simplified, sub-struct, array
// e escalares. O offset dos filhos é o do pai mais o deslocamento relativo.
func walkValue(rt reflect.Type, base int, name string, path []int, c *collector) error {
	segSize := typeSize(segmentType)
	switch {
	case rt == segmentType:
		return c.add(name, KindText, base, segSize, path)
	case isTextPair(rt):
		// As duas chaves saem do nome do campo declarante, não dos nomes
		// internos do par (Standard não vira `{campo}_standard`).
		if err := c.add(name, KindText, base, segSize, appendStep(path, 0)); err != nil {
			return err
		}
		return c.add(name+"_simplified", KindText, base+segSize, segSize, appendStep(path, 1))
	}
	switch rt.Kind() {
	case reflect.Struct:
		next := name
		if next != "" {
			next += "_"
		}
		return walkStruct(rt, base, next, path, c)
	case reflect.Array:
		elemSize := typeSize(rt.Elem())
		if elemSize < 0 {
			return fmt.Errorf("array de %s nao suportado", rt.Elem())
		}
		for e := 0; e < rt.Len(); e++ {
			child := fmt.Sprintf("%s_%d", name, e)
			if err := walkValue(rt.Elem(), base+e*elemSize, child, appendStep(path, -(e+1)), c); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		// Tamanho dinâmico não serializa em encoding/binary; sobras devem ir
		// em Tail (após a struct) ou em arrays fixos.
		return fmt.Errorf("slice nao suportado (use array fixo ou Tail)")
	default:
		return c.add(name, KindData, base, typeSize(rt), path)
	}
}

// isTextPair reconhece o par standard/simplified (schema.TextPair): uma
// struct de exatamente dois models.Segment chamados Standard e Simplified,
// nessa ordem. Assim o chunkmap não precisa importar os pacotes de schema.
func isTextPair(rt reflect.Type) bool {
	if rt.Kind() != reflect.Struct || rt.NumField() != 2 {
		return false
	}
	for i, want := range []string{"Standard", "Simplified"} {
		f := rt.Field(i)
		if f.Name != want || f.Type != segmentType {
			return false
		}
	}
	return true
}

// appendStep devolve um caminho novo (sem aliasing do backing array).
func appendStep(path []int, step int) []int {
	out := make([]int, len(path), len(path)+1)
	copy(out, path)
	return append(out, step)
}

// snake converte um identificador Go (PascalCase) para snake_case — a forma
// das chaves dos arquivos JSON do projeto. Trata corrida de maiúsculas
// (inicialismos Go) e dígitos:
//
//	HPMax              -> hp_max
//	CostATB            -> cost_atb
//	BTLSequence        -> btl_sequence
//	EffectDescription  -> effect_description
//	ProhibitConsump2MP -> prohibit_consump_2mp
//	SubMenuCat2        -> sub_menu_cat2
//	Reserve1           -> reserve1
//
// É uma função total sobre ASCII (todos os nomes de campo são ASCII).
func snake(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(name) + 8)
	for i := 0; i < len(name); i++ {
		if separaAntes(name, i) {
			b.WriteByte('_')
		}
		if c := name[i]; c >= 'A' && c <= 'Z' {
			b.WriteByte(c + ('a' - 'A'))
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// separaAntes reporta se um '_' deve ser inserido logo antes de name[i].
func separaAntes(name string, i int) bool {
	c := name[i]
	switch {
	case isDigitByte(c):
		// Um dígito seguido de maiúscula inicia palavra nova:
		// ProhibitConsump2MP -> prohibit_consump_2mp (C# `prohibit_consump_2mp`).
		// Dígito no fim da palavra não separa: Reserve1 -> reserve1,
		// Dummy10 -> dummy10, SubMenuCat2 -> sub_menu_cat2.
		if i > 0 && isDigitByte(name[i-1]) {
			return false // só o primeiro dígito do grupo separa
		}
		j := i
		for j < len(name) && isDigitByte(name[j]) {
			j++
		}
		return j < len(name) && isUpperByte(name[j])
	case isUpperByte(c):
		if i == 0 || name[i-1] == '_' {
			return false
		}
		prev := name[i-1]
		switch {
		case isLowerByte(prev):
			return true
		case isDigitByte(prev):
			// já separamos antes do grupo de dígitos: não separar de novo
			j := i - 1
			for j > 0 && isDigitByte(name[j-1]) {
				j--
			}
			return !separaAntes(name, j)
		default: // maiúscula anterior: quebra só se inicia palavra nova
			return i+1 < len(name) && isLowerByte(name[i+1])
		}
	default:
		return false
	}
}

func isLowerByte(c byte) bool { return c >= 'a' && c <= 'z' }
func isUpperByte(c byte) bool { return c >= 'A' && c <= 'Z' }
func isDigitByte(c byte) bool { return c >= '0' && c <= '9' }

// typeSize devolve binary.Size do tipo (-1 se não suportado).
// O .Interface() é obrigatório: passar o reflect.Value como any faria o
// binary.Size refletir sobre o próprio wrapper (campo flag/uintptr) e
// devolver -1 mesmo para tipos válidos.
//
// binary.Size (e não Type.Size()) é o que vale: encoding/binary escreve os
// campos contíguos, sem o padding de alinhamento que o Go adiciona.
func typeSize(rt reflect.Type) int {
	return binary.Size(reflect.New(rt).Elem().Interface())
}
