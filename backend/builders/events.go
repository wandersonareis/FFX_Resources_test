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
		entry := dto.FileEntry{
			Metadata: dto.NewEventMetadata(id, version),
			Rows:     make([]dto.TextRow, 0, len(ev.Strings)),
		}
		for i, str := range ev.Strings {
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
			continue
		}
		dto.SortRows(entry.Rows)
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

// ApplyEventsDTO aplica o DTO de volta no store/binário (parse DTO → binário).
// Só o applier faz isso; o formatter nunca toca no binário.
// Se ids vazio, aplica todas as entradas da Collection; senão, só as pedidas.
func ApplyEventsDTO(version common.GameVersion, c dto.Collection, ids []string) error {
	filter := make(map[string]bool)
	if len(ids) > 0 {
		for _, id := range ids {
			filter[id] = true
		}
	}
	var failed []string
	var applied []string
	for _, key := range c.SortedKeys() {
		if len(filter) > 0 && !filter[key] {
			continue
		}
		entry := c[key]
		if err := applySingleEventDTO(version, key, entry); err != nil {
			common.LogError("failed to update event %s: %v", key, err)
			failed = append(failed, key)
			continue
		}
		applied = append(applied, key)
	}
	if len(applied) == 0 && len(failed) == 0 {
		return fmt.Errorf("no matching events in DTO to apply")
	}
	// Persiste os aplicados de volta no binário.
	var saveFailed []string
	for _, id := range applied {
		if err := event.ExportEventStringsToLocalizations(version, id); err != nil {
			common.LogVerbose("Error saving event %s: %v", id, err)
			saveFailed = append(saveFailed, id)
		}
	}
	if len(failed) > 0 || len(saveFailed) > 0 {
		sort.Strings(failed)
		sort.Strings(saveFailed)
		return fmt.Errorf("failed to update %d event(s): %v; failed to save %d event(s): %v",
			len(failed), failed, len(saveFailed), saveFailed)
	}
	common.LogVerbose("Events processed successfully! (%d events)", len(applied))
	return nil
}

func applySingleEventDTO(version common.GameVersion, eventID string, entry dto.FileEntry) error {
	ev := event.GetEvent(version, eventID)
	if ev == nil {
		return fmt.Errorf("event not found in memory: %s", eventID)
	}
	dto.SortRows(entry.Rows)
	for _, row := range entry.Rows {
		if row.Index < 0 || row.Index >= len(ev.Strings) {
			return fmt.Errorf("string index out of range for event %s: %d", eventID, row.Index)
		}
		obj := ev.Strings[row.Index]
		if obj == nil {
			continue
		}
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				common.LogVerbose("unsupported localization %s for event %s[%d]", lang, eventID, row.Index)
				continue
			}
			if want, ok := row.Hash[lang]; ok && want != "" {
				if got := hash.Sum64Hex(newText); got != want {
					common.LogVerbose("hash mismatch for event %s[%d] lang %s: file %s vs text %s",
						eventID, row.Index, lang, want, got)
				}
			}
			fs := obj.GetLocalizedContent(lang)
			if fs == nil {
				return fmt.Errorf("failed to get localized content for %s", lang)
			}
			if fs.GetRegularString() == newText {
				continue
			}
			fs.SetRegularString(newText)
		}
	}
	event.SetEvent(version, eventID, ev)
	return nil
}
