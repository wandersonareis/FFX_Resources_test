// Package json implementa o formato JSON usado na exportação/importação
// dos arquivos de localização. É o único ponto público do formato JSON:
// os pacotes de domínio (event, objectsfile, macrodic) desconhecem JSON.
//
// Os formatters recebem apenas DTO pronto (backend/dto) e devolvem DTO:
// quem monta o DTO é backend/builders e quem aplica o DTO de volta no
// binário é o applier do domínio.
package json

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// extensionJSON é a extensão produzida pelos formatters JSON.
const extensionJSON = ".json"

// ExtensionJSON retorna a extensão produzida pelos formatters JSON.
func ExtensionJSON() string {
	return extensionJSON
}

// marshalNoEscape serializa v com indentação de 2 espaços e sem escapar HTML
// (`&`, `<`, `>` saem literais).
func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// minDedupRunes é o tamanho mínimo (em runes) para um texto virar
// referência `$hash`. Partículas curtas como "ok" nunca viram ref:
// repetem-se muito e cada ocorrência pode ter tradução distinta por
// contexto. A contagem é em runes para não miscaracterizar CJK
// (ex.: "你好" tem 2 runes mas 6 bytes).
const minDedupRunes = 5

// marshalCollection serializa a Collection com chaves e rows ordenadas
// (saída determinística), com todos os idiomas. Recebe apenas DTO pronto.
func marshalCollection(c dto.Collection) ([]byte, error) {
	return marshalCollectionLangs(c, nil)
}

// marshalCollectionLangs serializa a Collection contendo só os idiomas
// pedidos (nil/vazio = todos).
//
// Na saída, todo hash ganha o prefixo `$` e, só no idioma default, textos
// repetidos (len >= 5 runes) viram referência `$hash` — dedup por arquivo.
// O DTO de entrada nunca é mutado.
func marshalCollectionLangs(c dto.Collection, langs []string) ([]byte, error) {
	filter := langSet(langs)
	seen := make(map[string]string) // hashHex bare -> texto (só default lang)
	ordered := make(map[string]dto.FileEntry, len(c))
	for _, k := range c.SortedKeys() {
		entry := c[k]
		rows := make([]dto.TextRow, 0, len(entry.Rows))
		for _, row := range entry.Rows {
			r := dto.TextRow{Index: row.Index, Name: row.Name}
			if row.Hash != nil {
				r.Hash = make(map[string]string, len(row.Hash))
				for lang, h := range row.Hash {
					if filter != nil && !filter[lang] {
						continue
					}
					r.Hash[lang] = hash.Prefix(h)
				}
			}
			if row.Text != nil {
				r.Text = make(map[string]string, len(row.Text))
				for lang, t := range row.Text {
					if filter != nil && !filter[lang] {
						continue
					}
					r.Text[lang] = t
				}
			}
			if t := r.Text[common.DefaultLocalization]; utf8.RuneCountInString(t) >= minDedupRunes {
				if h, _ := hash.Strip(r.Hash[common.DefaultLocalization]); h != "" {
					if _, ok := seen[h]; ok {
						r.Text[common.DefaultLocalization] = hash.Prefix(h)
					} else {
						seen[h] = t
					}
				}
			}
			rows = append(rows, r)
		}
		dto.SortRows(rows)
		entry.Rows = rows
		ordered[k] = entry
	}
	return marshalNoEscape(ordered)
}

// langSet normaliza o filtro de idiomas (nil = todos).
func langSet(langs []string) map[string]bool {
	if len(langs) == 0 {
		return nil
	}
	set := make(map[string]bool, len(langs))
	for _, l := range langs {
		if l = strings.TrimSpace(l); l != "" {
			set[l] = true
		}
	}
	return set
}

// unmarshalCollection parseia o JSON de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
//
// O import nunca recalcula nem valida hash: o hash é ponteiro opaco.
// Remove o `$` dos hashes (DTO volta a hex puro) e expande referências
// `$hash` pelo valor atual da tabela hash→texto do próprio arquivo —
// seja qual for o valor ali (inclusive texto editado sob hash antigo).
// `$` órfão (fora da tabela = JSON danificado) é mantido + LogVerbose.
func unmarshalCollection(data []byte) (dto.Collection, error) {
	var c dto.Collection
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	table := make(map[string]string)
	for _, k := range c.SortedKeys() {
		entry := c[k]
		for i := range entry.Rows {
			row := &entry.Rows[i]
			for lang, h := range row.Hash {
				bare, _ := hash.Strip(h)
				row.Hash[lang] = bare
			}
			t := row.Text[common.DefaultLocalization]
			ref, isRef := hash.Strip(t)
			if isRef && ref == row.Hash[common.DefaultLocalization] {
				continue // referência: resolve na segunda passada
			}
			if own := row.Hash[common.DefaultLocalization]; own != "" && t != "" {
				table[own] = t
			}
		}
		c[k] = entry
	}
	for _, k := range c.SortedKeys() {
		entry := c[k]
		for i := range entry.Rows {
			row := &entry.Rows[i]
			t := row.Text[common.DefaultLocalization]
			ref, isRef := hash.Strip(t)
			if !isRef || ref != row.Hash[common.DefaultLocalization] {
				continue
			}
			if v, ok := table[ref]; ok {
				row.Text[common.DefaultLocalization] = v
			} else {
				common.LogVerbose("orphan dedup reference %q kept as-is", t)
			}
		}
		c[k] = entry
	}
	for k, entry := range c {
		dto.SortRows(entry.Rows)
		c[k] = entry
	}
	return c, nil
}
