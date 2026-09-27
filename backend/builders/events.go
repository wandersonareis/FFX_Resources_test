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
		entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
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
//
// Com dedup, o algoritmo é o mesmo do help (4 fases — todas as decisões
// tomadas contra o estado PRÉ-aplicação, para o resultado não depender da
// ordem das entradas do lote):
//  1. defs do lote: hash → texto literal de 'us' (refs resolvem por aqui;
//     def fora do lote ⇒ identidade — a ref mantém o texto atual);
//  2. índice pré-estado: texto 'us' → todos os segmentos iguais em TODOS os
//     eventos da versão (a alavanca da propagação anti-desalinhamento: a
//     edição de uma def alcança os gêmeos — mesmo fora do lote — e cada
//     binário tocado é recompilado);
//  3. coleta: valida a estrutura de cada entrada e agenda as mudanças —
//     um 'us' que difere do estado atual agenda TODAS as cópias dele;
//  4. escrita + save só dos eventos realmente tocados.
func ApplyEventsDTO(version common.GameVersion, c dto.Collection, ids []string) error {
	filter := make(map[string]bool)
	if len(ids) > 0 {
		for _, id := range ids {
			filter[id] = true
		}
	}
	inBatch := func(id string) bool { return len(filter) == 0 || filter[id] }

	// Fase 1 — defs do lote: hash → texto literal de 'us'. Montadas antes
	// de qualquer escrita porque uma ref pode vir antes da def na ordem das
	// entradas. Ref não vira def (o valor dela é $hash, não texto).
	defs := make(map[string]string)
	for _, id := range c.SortedKeys() {
		if !inBatch(id) {
			continue
		}
		for _, row := range c[id].Rows {
			t := row.Text[common.DefaultLocalization]
			if t == "" {
				continue
			}
			if _, isRef := refBare(row, t); isRef {
				continue
			}
			if h := row.Hash[common.DefaultLocalization]; h != "" {
				defs[h] = t
			}
		}
	}

	// Fase 2 — índice pré-estado: texto 'us' → todos os FieldStrings iguais
	// em TODOS os eventos da versão. É por ele que cada edição enxerga as
	// cópias ANTES de qualquer escrita: um texto novo igual ao velho de
	// outro evento não arrasta alvos alheios nem reverte rascunho de
	// terceiros.
	index := make(map[string][]eventTarget)
	for _, id := range event.GetAllEventIDs(version) {
		ev := event.GetEvent(version, id)
		if ev == nil {
			continue
		}
		for _, obj := range ev.Strings {
			if obj == nil {
				continue
			}
			fs := obj.GetLocalizedContent(common.DefaultLocalization)
			if fs == nil {
				continue
			}
			text := fs.GetRegularString()
			index[text] = append(index[text], eventTarget{eventID: id, fs: fs})
		}
	}

	// Fase 3 — coleta: valida tudo do lote antes de escrever qualquer coisa
	// (um DTO malformado não deixa meio evento aplicado).
	var failed []string
	var applied []string
	var usChanges, dirChanges []eventChange
	for _, id := range c.SortedKeys() {
		if !inBatch(id) {
			continue
		}
		ev := event.GetEvent(version, id)
		if ev == nil {
			common.LogError("failed to update event %s: evento não carregado", id)
			failed = append(failed, id)
			continue
		}
		us, dir, err := collectEventChanges(id, ev, c[id], defs, index)
		if err != nil {
			common.LogError("failed to update event %s: %v", id, err)
			failed = append(failed, id)
			continue
		}
		usChanges = append(usChanges, us...)
		dirChanges = append(dirChanges, dir...)
		applied = append(applied, id)
	}
	if len(applied) == 0 && len(failed) == 0 {
		return fmt.Errorf("no matching events in DTO to apply")
	}

	// Fase 4 — escrita: alvos congelados no pré-estado; conflitos (mesmo
	// texto velho editado de formas diferentes) resolvem por último da ordem
	// do lote (SortedKeys), mantendo o grupo inteiro alinhado.
	saveSet := make(map[string]bool)
	for _, ch := range dirChanges {
		for _, t := range ch.targets {
			if ch.newText == t.fs.GetRegularString() {
				continue
			}
			t.fs.SetRegularString(ch.newText)
			saveSet[t.eventID] = true
		}
	}
	for _, ch := range usChanges {
		for _, t := range ch.targets {
			if ch.newText == t.fs.GetRegularString() {
				continue
			}
			t.fs.SetRegularString(ch.newText)
			saveSet[t.eventID] = true
		}
	}

	// Save: só o que recebeu texto novo (lote ou propagação) — o binário
	// compilado vai para mods/; o original do gamefiles fica intacto.
	var saveFailed []string
	for _, id := range event.GetAllEventIDs(version) {
		if !saveSet[id] {
			continue
		}
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

// eventTarget é um segmento 'us' alvo de escrita, com seu evento.
type eventTarget struct {
	eventID string
	fs      *event.FieldString
}

// eventChange é uma mudança agendada: todos os FieldStrings que devem
// passar a ter newText (em 'us', as cópias idênticas do estado pré-aplicação).
type eventChange struct {
	targets []eventTarget
	newText string
}

// collectEventChanges valida a entrada e agenda as mudanças dela contra o
// estado pré-aplicação (index). Não escreve nada.
func collectEventChanges(eventID string, ev *event.EventFile, entry dto.FileEntry, defs map[string]string, index map[string][]eventTarget) (usChanges, dirChanges []eventChange, err error) {
	dto.SortRows(entry.Rows)

	// Fase 1 — validação estrutural.
	for _, row := range entry.Rows {
		if row.Index < 0 || row.Index >= len(ev.Strings) {
			return nil, nil, fmt.Errorf("string index out of range for event %s: %d", eventID, row.Index)
		}
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				common.LogVerbose("unsupported localization %s for event %s[%d]", lang, eventID, row.Index)
				continue
			}
			obj := ev.Strings[row.Index]
			if obj == nil {
				continue
			}
			if obj.GetLocalizedContent(lang) == nil {
				return nil, nil, fmt.Errorf("failed to get localized content for %s", lang)
			}
		}
	}

	// Fase 2 — agendamento contra o pré-estado.
	for _, row := range entry.Rows {
		obj := ev.Strings[row.Index]
		if obj == nil {
			continue
		}
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				continue
			}
			fs := obj.GetLocalizedContent(lang)

			if lang != common.DefaultLocalization {
				// Idiomas ≠ us nunca participam do dedup nem da
				// propagação: escrevem direto quando mudam.
				if newText != fs.GetRegularString() {
					dirChanges = append(dirChanges, eventChange{
						targets: []eventTarget{{eventID: eventID, fs: fs}},
						newText: newText,
					})
				}
				continue
			}

			// 'us': ref vira o texto da def do lote (sem def = identidade,
			// mantém o atual); literal editada agenda as cópias.
			if bare, isRef := refBare(row, newText); isRef {
				resolved, ok := defs[bare]
				if !ok {
					continue
				}
				newText = resolved
			}
			if newText == fs.GetRegularString() {
				continue // sem mudança no segmento desta row: nada a agendar
			}
			targets := index[fs.GetRegularString()]
			if len(targets) == 0 {
				targets = []eventTarget{{eventID: eventID, fs: fs}}
			}
			usChanges = append(usChanges, eventChange{targets: targets, newText: newText})
		}
	}
	return usChanges, dirChanges, nil
}
