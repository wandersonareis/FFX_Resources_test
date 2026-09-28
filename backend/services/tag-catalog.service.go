// Catálogo de tags nomeadas (PC, MCR, BUTTON, ICON) para o autocomplete do
// editor de texto. Serve os valores válidos do estágio 2 e a ordem de exibição
// (ranking de uso gerado por scripts/tagfreq — só reordena, nunca exclui).
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
)

// TagSuggestion é um valor completo de tag nomeada, pronto para inserir como
// chip atômico no editor.
type TagSuggestion struct {
	// Tag é a tag completa e canônica, ex.: {PC:01:YUNA}.
	Tag string `json:"tag"`
	// Label é o texto exibido no popup, ex.: YUNA.
	Label string `json:"label"`
	// Key é a identidade estável do valor sem o texto humano, usada na
	// validação de tags digitadas e no ranking de frequência:
	// "01" (PC/BUTTON/ICON) ou "s01:l05" (MCR).
	Key string `json:"key"`
}

// TagCatalogEntry é uma tag nomeada: o prefixo inserido no estágio 1 do
// autocomplete e os valores do estágio 2.
type TagCatalogEntry struct {
	// Tag é o nome canônico (PC, MCR, BUTTON, ICON).
	Tag string `json:"tag"`
	// Prefix é o texto inserido ao escolher no estágio 1, ex.: "{PC:".
	Prefix string `json:"prefix"`
	// Hint é a descrição exibida no popup do estágio 1.
	Hint string `json:"hint"`
	// Values são os valores possíveis (estágio 2), ordenados por uso.
	Values []TagSuggestion `json:"values"`
}

// TagCatalog é o catálogo servido ao frontend via binding GetTagCatalog.
type TagCatalog struct {
	Version string            `json:"version"`
	Tags    []TagCatalogEntry `json:"tags"`
}

// tagFrequencyPath resolve mods/edits/tag_frequency_<v>.json (mesmo diretório
// dos artefatos de export).
func tagFrequencyPath(version common.GameVersion) string {
	name := common.WithVersionSuffixFor("tag_frequency.json", version)
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", name)
}

// loadTagFrequency lê o ranking de uso. Ausente ou ilegível → mapa vazio
// (ordem fixa por índice; o ranking é opcional).
func loadTagFrequency(version common.GameVersion) map[string]int {
	counts := map[string]int{}
	raw, err := os.ReadFile(tagFrequencyPath(version))
	if err != nil {
		return counts
	}
	if err := json.Unmarshal(raw, &counts); err != nil {
		common.LogError("tag catalog: ranking de frequência inválido em %s: %v", tagFrequencyPath(version), err)
		return map[string]int{}
	}
	return counts
}

// sortByFrequency reordena por uso decrescente (estável: empates mantêm a
// ordem de índice). Só reordena — nunca remove valores.
func sortByFrequency(values []TagSuggestion, tag string, freq map[string]int) {
	sort.SliceStable(values, func(i, j int) bool {
		return freq[tag+":"+values[i].Key] > freq[tag+":"+values[j].Key]
	})
}

// byteTableValues monta as sugestões de uma tabela estática do encoding
// (índice → nome) em ordem crescente de índice, no formato exato que o
// decoder emite ({TAG:XX:Nome}).
func byteTableValues(prefix string, table map[byte]string) []TagSuggestion {
	idxs := make([]int, 0, len(table))
	for idx := range table {
		idxs = append(idxs, int(idx))
	}
	sort.Ints(idxs)
	out := make([]TagSuggestion, 0, len(idxs))
	for _, idx := range idxs {
		name := table[byte(idx)]
		key := fmt.Sprintf("%02X", idx)
		out = append(out, TagSuggestion{
			Tag:   fmt.Sprintf("{%s:%s:%s}", prefix, key, name),
			Label: name,
			Key:   key,
		})
	}
	return out
}

// MCR hash pattern: conteúdo não traduzido/em falta vira o hash de 16 hex
// ($c655a394619027cb) do placeholder — inútil como sugestão de texto.
var mcrUnresolvedHash = regexp.MustCompile(`^\$[0-9a-fA-F]{16}$`)

// mcrValues monta as sugestões de MCR a partir do macrodic em memória. A tag
// reproduz o formato exato do decoder: {MCR:sXX:lYY:"texto"} com aspas quando
// há macro e {MCR:sXX:lYY:<Missing>} sem aspas quando não há — assim o
// round-trip byte-idêntico se mantém e a tag fecha certo no "}".
func mcrValues(version common.GameVersion) []TagSuggestion {
	macros := datastore.GetMacros(version)
	if macros == nil || macros.Count() == 0 {
		return nil
	}
	type keyed struct {
		idx int
		val TagSuggestion
	}
	list := make([]keyed, 0, macros.Count())
	macros.ForEach(func(idx int, m datastore.IGlobalLocalizedMacroStringObject) {
		if m == nil {
			return
		}
		section := idx / 0x100
		line := idx % 0x100
		// A sessão 0 é inalcançável pela tag MCR (o byte 0x13 é PC), então
		// esses chunks nunca podem ser sugestão válida.
		if section == 0 || section > 0x0F {
			return
		}
		key := fmt.Sprintf("s%02X:l%02X", section, line)
		text := "<Missing>"
		if content, ok := m.GetLocalizedContent(common.DefaultLocalization); ok && content != nil {
			text = content.String()
		}
		// Texto com aspas/chaves quebraria o parse da tag; não sugere.
		if strings.ContainsAny(text, `"{}`) {
			return
		}
		// Hash não resolvido não serve de sugestão (texto de verdade existe
		// em outra localization ou o slot é placeholder); fica fora.
		if mcrUnresolvedHash.MatchString(strings.TrimSpace(text)) {
			return
		}
		label := strings.TrimSpace(text)
		if label == "" {
			label = key
		}
		// O decoder fecha a tag direto no "}"; aspas só envolvem o texto
		// quando há macro — "<Missing>" sai sem aspas (string_bytes.go).
		value := `"` + text + `"`
		if text == "<Missing>" {
			value = text
		}
		list = append(list, keyed{idx, TagSuggestion{
			Tag:   "{MCR:" + key + ":" + value + "}",
			Label: label,
			Key:   key,
		}})
	})
	sort.Slice(list, func(i, j int) bool { return list[i].idx < list[j].idx })
	out := make([]TagSuggestion, len(list))
	for i, item := range list {
		out[i] = item.val
	}
	return out
}

// BuildTagCatalog monta o catálogo de tags nomeadas da versão. Garante a
// versão pronta (charsets + macrodic) antes de ler; os valores vêm dos mapas
// estáticos do encoding (PC/BUTTON/ICON) e do macrodic em memória (MCR).
func BuildTagCatalog(version common.GameVersion) (TagCatalog, error) {
	if err := ensureVersionReady(version); err != nil {
		return TagCatalog{}, fmt.Errorf("tag catalog: %w", err)
	}
	freq := loadTagFrequency(version)
	entries := []TagCatalogEntry{
		{Tag: "PC", Prefix: "{PC:", Hint: "Personagem", Values: byteTableValues("PC", ffxencoding.PlayerCharTable(version))},
		{Tag: "MCR", Prefix: "{MCR:", Hint: "Macro (NPC, nomes, locais — filtre por sXX ou texto)", Values: mcrValues(version)},
		{Tag: "BUTTON", Prefix: "{BUTTON:", Hint: "Botão", Values: byteTableValues("BUTTON", ffxencoding.ButtonTable())},
		{Tag: "ICON", Prefix: "{ICON:", Hint: "Ícone", Values: byteTableValues("ICON", ffxencoding.IconTable())},
	}
	for i := range entries {
		sortByFrequency(entries[i].Values, entries[i].Tag, freq)
	}
	return TagCatalog{Version: version.String(), Tags: entries}, nil
}
