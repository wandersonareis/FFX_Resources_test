// Package strings implementa o formato Strings (java-like) de exportação/
// importação dos textos de localização. É um formato texto-para-humanos
// que vive ao lado do JSON: os builders montam o mesmo DTO
// (backend/dto) e cada formatter o serializa à sua maneira.
//
// Gramática:
//
//	<entry-block> := <header> <line>+
//	<header>      := "/*key=" <key> " row_count=" <n> "*/"
//	<line>        := <keyfields> "║" "$" <hash16> " " "=" " " <value>
//	<keyfields>   := [<prefix> ":"]... <id> [":" <name>] ":" <index> ":" <lang>
//	<value>       := <literal> | "$" <hash16>          // def | ref
//
// Âncora semântica à direita: último campo = lang, penúltimo = index;
// campos novos (versão, path) crescem pela esquerda sem quebrar o parser.
// "║" liga a chave ao hash de identidade; "=" liga ao valor com split no
// primeiro ("=" dentro do texto passa intacto; a chave nunca tem espaço).
// def: literal cujo xxh64 == hash da chave (verificação grátis na leitura).
// ref: "$hash" resolvido pela tabela hash→texto do próprio arquivo.
// Name é omitido quando vazio (events); objects/macro emitem
// <id>:<name>:<index>:<lang>.
package strings

import (
	"sort"
	"strconv"
	stdstrings "strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// extensionStrings é a extensão produzida pelo formatter Strings.
const extensionStrings = ".strings"

// ExtensionStrings retorna a extensão produzida pelo formatter Strings.
func ExtensionStrings() string {
	return extensionStrings
}

// keyValueSep liga a chave ao hash de identidade (U+2566).
const keyValueSep = "║"

// minDedupRunes replica a regra do JSON (formatters/json): só textos do
// idioma default com ao menos 5 runes viram referência $hash.
const minDedupRunes = 5

// Marshal organiza a Collection pronta no formato Strings (todos os idiomas).
// Recebe apenas DTO pronto; não toca no domínio nem muta a entrada.
func Marshal(c dto.Collection) ([]byte, error) {
	return MarshalLangs(c, nil)
}

// MarshalLangs organiza a Collection pronta no formato Strings contendo
// só os idiomas pedidos (nil/vazio = todos). Uma linha por (row, idioma)
// não-vazio, com chaves e rows ordenadas (saída determinística).
func MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	filter := langSet(langs)
	var out stdstrings.Builder
	seen := make(map[string]string) // hashHex bare -> texto (só default lang)
	for _, k := range c.SortedKeys() {
		entry := c[k]
		out.WriteString("/*key=" + entry.Metadata.Key + " row_count=" + strconv.Itoa(len(entry.Rows)) + "*/\n")
		rows := make([]dto.TextRow, len(entry.Rows))
		copy(rows, entry.Rows)
		dto.SortRows(rows)
		version, id := versionAndID(entry.Metadata, k)
		for _, row := range rows {
			for _, lang := range rowLangs(row, filter) {
				text := row.Text[lang]
				if text == "" {
					continue
				}
				h := row.Hash[lang]
				if h == "" {
					h = hash.Sum64Hex(text)
				}
				value := escapeValue(text)
				if lang == common.DefaultLocalization {
					if runeCount(text) >= minDedupRunes {
						if _, dup := seen[h]; dup {
							value = hash.Prefix(h)
						} else {
							seen[h] = text
						}
					}
				}
				if version != "" {
					out.WriteString(version + ":")
				}
				out.WriteString(id)
				if row.Name != "" {
					out.WriteString(":" + row.Name)
				}
				out.WriteString(":" + strconv.Itoa(row.Index) + ":" + lang + keyValueSep + hash.Prefix(h) + " = " + value + "\n")
			}
		}
	}
	return []byte(out.String()), nil
}

// versionAndID resolve o prefixo esquerdo da linha a partir da metadata
// (fallback: chave do mapa como id, sem versão).
func versionAndID(m dto.Metadata, mapKey string) (version, id string) {
	id = m.ID
	if id == "" {
		id = mapKey
	}
	if p, ok := dto.ParseKey(m.Key); ok {
		version = p.Version
	}
	return version, id
}

// langSet normaliza o filtro de idiomas (nil = todos).
func langSet(langs []string) map[string]bool {
	if len(langs) == 0 {
		return nil
	}
	set := make(map[string]bool, len(langs))
	for _, l := range langs {
		if l = stdstrings.TrimSpace(l); l != "" {
			set[l] = true
		}
	}
	return set
}

// rowLangs devolve os idiomas da row com texto, ordenados e filtrados.
func rowLangs(row dto.TextRow, filter map[string]bool) []string {
	langs := make([]string, 0, len(row.Text))
	for lang, text := range row.Text {
		if text == "" {
			continue
		}
		if filter != nil && !filter[lang] {
			continue
		}
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// escapeValue escapa valor literal para uma linha: backslash, CR, LF e ║.
func escapeValue(s string) string {
	s = stdstrings.ReplaceAll(s, `\`, `\\`)
	s = stdstrings.ReplaceAll(s, "\r", `\r`)
	s = stdstrings.ReplaceAll(s, "\n", `\n`)
	s = stdstrings.ReplaceAll(s, keyValueSep, `\║`)
	return s
}

// unescapeValue desfaz escapeValue (leniente: sequência desconhecida
// mantém backslash + char).
func unescapeValue(s string) string {
	var out stdstrings.Builder
	out.Grow(len(s))
	esc := false
	for _, r := range s {
		if esc {
			switch r {
			case 'n':
				out.WriteRune('\n')
			case 'r':
				out.WriteRune('\r')
			case '\\':
				out.WriteRune('\\')
			case '║':
				out.WriteString(keyValueSep)
			default:
				out.WriteRune('\\')
				out.WriteRune(r)
			}
			esc = false
			continue
		}
		if r == '\\' {
			esc = true
			continue
		}
		out.WriteRune(r)
	}
	if esc {
		out.WriteRune('\\')
	}
	return out.String()
}

func runeCount(s string) int {
	return len([]rune(s))
}
