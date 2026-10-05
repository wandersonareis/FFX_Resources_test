package builders

// MOTOR DE APPLY DE TABELAS DE TEXTO (events e eventtable) — 4 fases.
//
// Todas as decisões são tomadas contra o estado PRÉ-aplicação, para o
// resultado não depender da ordem das entradas do lote:
//
//  1. defs do lote: hash → texto literal de 'us' (refs resolvem por aqui;
//     def fora do lote ⇒ identidade — a ref mantém o texto atual);
//  2. índice pré-estado: texto 'us' → todos os FieldStrings iguais em TODOS
//     os arquivos do ESCOPO (a alavanca da propagação: a edição de uma def
//     alcança as cópias — mesmo fora do lote — e cada binário tocado é
//     recompilado). Tradução divergente não participa: dedupe é de
//     original para original, por conteúdo;
//  3. coleta: valida a estrutura de cada entrada e agenda as mudanças —
//     um 'us' que difere do estado atual agenda TODAS as cópias dele;
//  4. escrita + save só dos arquivos realmente tocados.
//
// O escopo é quem separa data/ de .vbf: os dois fluxos rodam o MESMO motor;
// data/ lista tudo que está presente na árvore, o .vbf lista o que foi
// aberto no clique (sessão). Nada aqui decide presença — quem monta o
// escopo sabe o que está carregado.

import (
	"fmt"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
)

// TableApplyScope é o conjunto de arquivos CARREGADOS onde o apply atua.
// Exportado porque o .vbf monta o escopo fora do pacote (a sessão de
// cliques); data/ monta o dele internamente.
type TableApplyScope struct {
	// Ids são os arquivos carregados no escopo (ordenados na saída).
	Ids []string
	// StringsFor devolve as strings do arquivo; nil/vazio = não carregado.
	StringsFor func(id string) []*event.LocalizedFieldStringObject
	// Save grava o binário do arquivo tocado em mods/.
	Save func(id string) error
}

// ApplyTextDTOWithScope roda o motor com um escopo montado pelo chamador
// (o .vbf usa o da sessão).
func ApplyTextDTOWithScope(label string, c dto.Collection, scope TableApplyScope) error {
	return applyTextDTO(label, c, nil, scope)
}

// applyTarget é um segmento 'us' alvo de escrita, com seu arquivo.
type applyTarget struct {
	fileID string
	fs     *event.FieldString
}

// applyChange é uma mudança agendada: todos os FieldStrings que devem
// passar a ter newText (em 'us', as cópias idênticas do estado pré-aplicação).
type applyChange struct {
	targets []applyTarget
	newText string
}

// applyTextDTO roda as 4 fases sobre o lote e o escopo. label é o nome do
// kind para os logs/erros.
func applyTextDTO(label string, c dto.Collection, ids []string, scope TableApplyScope) error {
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
	// em TODOS os arquivos do escopo. É por ele que cada edição enxerga as
	// cópias ANTES de qualquer escrita: um texto novo igual ao velho de
	// outro arquivo não arrasta alvos alheios nem reverte rascunho de
	// terceiros. Tradução divergente não entra em grupo nenhum: dedupe é
	// de original para original.
	index := make(map[string][]applyTarget)
	for _, id := range scope.Ids {
		for _, obj := range scope.StringsFor(id) {
			if obj == nil {
				continue
			}
			fs := obj.GetLocalizedContent(common.DefaultLocalization)
			if fs == nil {
				continue
			}
			text := fs.GetRegularString()
			index[text] = append(index[text], applyTarget{fileID: id, fs: fs})
		}
	}

	// Fase 3 — coleta: valida tudo do lote antes de escrever qualquer coisa
	// (um DTO malformado não deixa meio arquivo aplicado).
	var failed, applied []string
	var usChanges, dirChanges []applyChange
	for _, id := range c.SortedKeys() {
		if !inBatch(id) {
			continue
		}
		strings := scope.StringsFor(id)
		if len(strings) == 0 {
			common.LogError("failed to update %s %s: arquivo não carregado no escopo", label, id)
			failed = append(failed, id)
			continue
		}
		us, dir, err := collectTableChanges(id, strings, c[id], defs, index)
		if err != nil {
			common.LogError("failed to update %s %s: %v", label, id, err)
			failed = append(failed, id)
			continue
		}
		usChanges = append(usChanges, us...)
		dirChanges = append(dirChanges, dir...)
		applied = append(applied, id)
	}
	if len(applied) == 0 && len(failed) == 0 {
		return fmt.Errorf("no matching %s entries in DTO to apply", label)
	}

	// Fase 4 — escrita: alvos congelados no pré-estado; conflitos (mesmo
	// texto velho editado de formas diferentes) resolvem por último da ordem
	// do lote (SortedKeys), mantendo o grupo inteiro alinhado.
	saveSet := make(map[string]bool)
	write := func(changes []applyChange) {
		for _, ch := range changes {
			for _, t := range ch.targets {
				if ch.newText == t.fs.GetRegularString() {
					continue
				}
				t.fs.SetRegularString(ch.newText)
				saveSet[t.fileID] = true
			}
		}
	}
	write(dirChanges)
	write(usChanges)

	// Save: só o que recebeu texto novo (lote ou propagação) — o binário
	// compilado vai para mods/; a fonte (data/ ou container) fica intacta.
	var saveFailed []string
	for _, id := range scope.Ids {
		if !saveSet[id] {
			continue
		}
		if err := scope.Save(id); err != nil {
			common.LogVerbose("Error saving %s %s: %v", label, id, err)
			saveFailed = append(saveFailed, id)
		}
	}
	if len(failed) > 0 || len(saveFailed) > 0 {
		sort.Strings(failed)
		sort.Strings(saveFailed)
		return fmt.Errorf("failed to update %d %s(s): %v; failed to save %d %s(s): %v",
			len(failed), label, failed, len(saveFailed), label, saveFailed)
	}
	common.LogVerbose("%s processed successfully! (%d entries)", label, len(applied))
	return nil
}

// collectTableChanges valida a entrada e agenda as mudanças dela contra o
// estado pré-aplicação (index). Não escreve nada.
func collectTableChanges(fileID string, strings []*event.LocalizedFieldStringObject, entry dto.FileEntry, defs map[string]string, index map[string][]applyTarget) (usChanges, dirChanges []applyChange, err error) {
	dto.SortRows(entry.Rows)

	// Fase 1 — validação estrutural.
	for _, row := range entry.Rows {
		if row.Index < 0 || row.Index >= len(strings) {
			return nil, nil, fmt.Errorf("string index out of range for %s: %d", fileID, row.Index)
		}
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				common.LogVerbose("unsupported localization %s for %s[%d]", lang, fileID, row.Index)
				continue
			}
			obj := strings[row.Index]
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
		obj := strings[row.Index]
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
					dirChanges = append(dirChanges, applyChange{
						targets: []applyTarget{{fileID: fileID, fs: fs}},
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
				targets = []applyTarget{{fileID: fileID, fs: fs}}
			}
			usChanges = append(usChanges, applyChange{targets: targets, newText: newText})
		}
	}
	return usChanges, dirChanges, nil
}
