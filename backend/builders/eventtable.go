// Famílias "eventtable" (tabelas no formato dos events: stride 8, offset u16
// por lado + flags/choices): battle/btl, cloudsave, tutorial.msb — ver
// backend/fileFormats/eventtable. Este pacote monta/aplica o DTO sobre o
// codec compartilhado com event (EncodeLocalizedStrings/RebuildFieldStrings).
package builders

import (
	"fmt"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	"ffxresources/backend/formatters/hash"
)

// ResolveTableTargets implementa o fluxo estrito de seleção (padrão events):
//   - ids vazio → todos os ids disponíveis na versão, ordenados;
//   - ids dado → somente os pedidos; id desconhecido é erro.
func ResolveTableTargets(kind string, version common.GameVersion, ids []string) ([]string, error) {
	if err := eventtable.Validate(kind); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		all := eventtable.IDs(kind, version)
		if len(all) == 0 {
			return nil, fmt.Errorf("no %s files found for version %s", kind, version)
		}
		sorted := make([]string, len(all))
		copy(sorted, all)
		sort.Strings(sorted)
		return sorted, nil
	}
	targets := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := eventtable.RelPath(kind, id); !ok {
			common.LogVerbose("%s: id %q fora da família, pulando", kind, id)
			continue
		}
		targets = append(targets, id)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("none of the %d requested %s file(s) found", len(ids), kind)
	}
	sort.Strings(targets)
	return targets, nil
}

// BuildTableDTO monta a Collection pronta (texto + metadata + hash) a
// partir dos binários carregados no store. ids vazio = tudo.
func BuildTableDTO(kind string, version common.GameVersion, ids []string) (dto.Collection, error) {
	targets, err := ResolveTableTargets(kind, version, ids)
	if err != nil {
		return nil, err
	}
	langs := SortedLocalizationKeys()
	out := make(dto.Collection, len(targets))
	for _, id := range targets {
		f, err := eventtable.LoadFromStore(kind, version, id)
		if err != nil {
			common.LogVerbose("%s: skip %s: %v", kind, id, err)
			continue
		}
		entry, ok := buildTableEntry(kind, id, version, f.Strings, langs)
		if !ok {
			continue
		}
		out[id] = entry
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no %s files with string data found", kind)
	}
	if err := hash.ValidateNoCollision(out); err != nil {
		return nil, err
	}
	return out, nil
}

// BuildTableEntryDTOFrom monta UMA entrada a partir das strings lidas direto
// da árvore (caminho do ORIGINAL/.vbf), sem tocar no store.
func BuildTableEntryDTOFrom(kind, id string, version common.GameVersion, strings []*event.LocalizedFieldStringObject) (dto.FileEntry, bool) {
	if len(strings) == 0 {
		return dto.FileEntry{}, false
	}
	return buildTableEntry(kind, id, version, strings, SortedLocalizationKeys())
}

// buildTableEntry converte as strings localizadas em rows do DTO. O Index é
// a posição física no binário (reconstrução posicional).
func buildTableEntry(kind, id string, version common.GameVersion, strings []*event.LocalizedFieldStringObject, langs []string) (dto.FileEntry, bool) {
	key, ok := eventtable.Key(version, kind, id)
	if !ok {
		return dto.FileEntry{}, false
	}
	entry := dto.FileEntry{
		Metadata: dto.Metadata{Key: key, ID: id},
		Rows:     make([]dto.TextRow, 0, len(strings)),
	}
	for i, str := range strings {
		if str == nil {
			continue
		}
		text := make(map[string]string, len(langs))
		for _, lang := range langs {
			text[lang] = str.GetLocalizedString(lang)
		}
		entry.Rows = append(entry.Rows, dto.TextRow{
			Index: i,
			Hash:  hash.Texts(text),
			Text:  text,
		})
	}
	if len(entry.Rows) == 0 {
		return dto.FileEntry{}, false
	}
	dto.SortRows(entry.Rows)
	entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
	return entry, true
}

// ApplyTableDTO aplica o DTO de volta nos binários pelo MESMO motor do
// events (4 fases — ver tableapply.go): o escopo é o store do kind
// (data/ — limitado apenas pela presença de arquivo), então a edição de
// uma def se espalha por índice de texto para TODAS as cópias presentes e
// cada binário tocado é recompilado em mods/. Tradução divergente não
// participa (dedupe é de original para original). Refs "$hash" do payload
// (view dedupada) resolvem contra as defs do lote.
func ApplyTableDTO(kind string, version common.GameVersion, c dto.Collection) error {
	if err := eventtable.Validate(kind); err != nil {
		return err
	}
	return applyTextDTO(kind, c, nil, TableApplyScope{
		Ids: eventtable.List(kind, version),
		StringsFor: func(id string) []*event.LocalizedFieldStringObject {
			f := eventtable.Get(kind, version, id)
			if f == nil {
				return nil
			}
			return f.Strings
		},
		Save: func(id string) error {
			f := eventtable.Get(kind, version, id)
			if f == nil {
				return fmt.Errorf("%s %s: não carregado", kind, id)
			}
			return f.Save()
		},
	})
}
