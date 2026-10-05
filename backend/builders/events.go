// Package builders é o intermediário глобal entre binário/store e DTO.
//
// Responsabilidade: montar o DTO pronto (texto + metadata + hash xxHash64)
// a partir dos dados brutos e aplicar o DTO de volta no binário/store.
// Os formatters recebem apenas o DTO pronto e nunca tocam no domínio.
package builders

import (
	"fmt"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/formatters/hash"
)

// SortedLocalizationKeys devolve as chaves de idioma ordenadas,
// para extração de texto determinística.
func SortedLocalizationKeys() []string {
	keys := make([]string, 0, len(common.SupportedLanguages))
	for k := range common.SupportedLanguages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ResolveEventTargets implementa o fluxo estrito de seleção:
//   - len(ids) == 0 → todos os eventos da versão, ordenados;
//   - len(ids) > 0 → extrai somente os pedidos, loga não encontrados
//     e retorna erro se zero for encontrado. Nunca cai para "tudo".
func ResolveEventTargets(version common.GameVersion, ids []string) ([]string, error) {
	if len(ids) == 0 {
		all := event.GetAllEventIDs(version)
		if len(all) == 0 {
			return nil, fmt.Errorf("no events loaded for version %s", version)
		}
		sorted := make([]string, len(all))
		copy(sorted, all)
		sort.Strings(sorted)
		return sorted, nil
	}
	var targets []string
	for _, id := range ids {
		if event.GetEvent(version, id) == nil {
			common.LogVerbose("event %s not found, skipping", id)
			continue
		}
		targets = append(targets, id)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("none of the %d requested event(s) found", len(ids))
	}
	sort.Strings(targets)
	return targets, nil
}

// BuildEventsDTO monta a Collection pronta (texto + metadata + hash)
// a partir dos dados brutos do store, para os ids dados.
// ids vazio = tudo. Chave da Collection = eventID (basename sem extensão).
func BuildEventsDTO(version common.GameVersion, ids []string) (dto.Collection, error) {
	targets, err := ResolveEventTargets(version, ids)
	if err != nil {
		return nil, err
	}
	langs := SortedLocalizationKeys()
	out := make(dto.Collection, len(targets))
	for _, id := range targets {
		ev := event.GetEvent(version, id)
		if ev == nil || len(ev.Strings) == 0 {
			common.LogVerbose("No strings found for event %s, skipping", id)
			continue
		}
		entry, ok := buildEventEntry(id, version, ev.Strings, langs)
		if !ok {
			continue
		}
		out[id] = entry
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no events with string data found")
	}
	// TODO: deletar quando colisão xxHash64 for considerada segura —
	// guarda temporária de desencargo: reprova DTO com mesmo hash para textos diferentes.
	if err := hash.ValidateNoCollision(out); err != nil {
		return nil, err
	}
	return out, nil
}

// BuildEventEntryDTOFrom monta UMA entrada de evento a partir das strings
// lidas direto dos arquivos (fonte explícita), sem tocar no store global.
// É o caminho do ORIGINAL: event.ReadLocalizedEventStringsFrom(SourceData).
// ok=false quando não há nenhuma row com texto.
func BuildEventEntryDTOFrom(id string, version common.GameVersion, strings []*event.LocalizedFieldStringObject) (dto.FileEntry, bool) {
	if len(strings) == 0 {
		return dto.FileEntry{}, false
	}
	return buildEventEntry(id, version, strings, SortedLocalizationKeys())
}

// buildEventEntry converte as strings localizadas de um evento em rows do
// DTO. O Index é a posição física no binário (reconstrução posicional).
func buildEventEntry(id string, version common.GameVersion, strings []*event.LocalizedFieldStringObject, langs []string) (dto.FileEntry, bool) {
	entry := dto.FileEntry{
		Metadata: dto.NewEventMetadata(id, version),
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
		common.LogVerbose("No strings found for event %s, skipping", id)
		return dto.FileEntry{}, false
	}
	dto.SortRows(entry.Rows)
	entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
	return entry, true
}

// ApplyEventsDTO aplica o DTO de volta no store/binário (parse DTO → binário).
// Só o applier faz isso; o formatter nunca toca no binário.
// Se ids vazio, aplica todas as entradas da Collection; senão, só as pedidas.
//
// O escopo é o store de eventos da versão (data/): TODOS os eventos
// carregados participam da propagação (índice de texto pré-estado) e só os
// binários tocados são recompilados. O motor é o mesmo do .vbf (sessão) —
// ver tableapply.go.
func ApplyEventsDTO(version common.GameVersion, c dto.Collection, ids []string) error {
	return applyTextDTO("event", c, ids, TableApplyScope{
		Ids: event.GetAllEventIDs(version),
		StringsFor: func(id string) []*event.LocalizedFieldStringObject {
			ev := event.GetEvent(version, id)
			if ev == nil {
				return nil
			}
			return ev.Strings
		},
		Save: func(id string) error {
			return event.ExportEventStringsToLocalizations(version, id)
		},
	})
}
