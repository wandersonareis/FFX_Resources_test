// Painéis de ajuda (kind help): montar o DTO pronto a partir dos binários
// .sps2 carregados no store e aplicar o DTO de volta. Segue o mesmo fluxo dos
// events: builder monta dto.Collection, formatter serializa, applier aplica.
package builders

import (
	"fmt"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/formatters/hash"
)

// ResolveHelpTargets implementa o fluxo estrito de seleção (padrão events):
//   - ids vazio → todos os painéis registrados, em ordem de registro;
//   - ids dado → somente os pedidos; id desconhecido é erro (ajuda o usuário
//     a perceber typo, diferente dos events que logam e seguem).
func ResolveHelpTargets(version common.GameVersion, ids []string) ([]string, error) {
	if err := helpfile.EnsureHelpLoaded(version); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		var all []string
		for _, name := range helpfile.HelpEntryNames() {
			if helpfile.GetHelp(version, name) != nil {
				all = append(all, name)
			}
		}
		if len(all) == 0 {
			return nil, fmt.Errorf("no help files found for version %s", version)
		}
		return all, nil
	}
	targets := make([]string, 0, len(ids))
	for _, id := range ids {
		if !helpfile.IsHelpEntry(id) {
			return nil, fmt.Errorf("painel de ajuda desconhecido: %s", id)
		}
		if helpfile.GetHelp(version, id) == nil {
			return nil, fmt.Errorf("painel de ajuda não carregado: %s", id)
		}
		targets = append(targets, id)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("none of the requested help file(s) found")
	}
	return targets, nil
}

// BuildHelpDTO monta a Collection pronta (texto + metadata + hash) a partir
// dos binários carregados no store. Chave da Collection = id do painel
// (now_help, now_help_page, mon_boku, s_monitor, dvdcopy, dvdcopy_page).
func BuildHelpDTO(version common.GameVersion, ids []string) (dto.Collection, error) {
	targets, err := ResolveHelpTargets(version, ids)
	if err != nil {
		return nil, err
	}
	langs := SortedLocalizationKeys()
	out := make(dto.Collection, len(targets))
	for _, name := range targets {
		panel := helpfile.GetHelp(version, name)
		if panel == nil {
			continue
		}
		entry, ok := buildHelpEntry(name, version, panel, langs)
		if !ok {
			continue
		}
		out[name] = entry
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no help panels with string data found")
	}
	if err := hash.ValidateNoCollision(out); err != nil {
		return nil, err
	}
	return out, nil
}

// BuildHelpEntryDTOFrom monta UMA entrada de painel a partir de um painel já
// montado de uma árvore explícita, sem tocar no store. É o caminho do
// ORIGINAL: helpfile.ReadHelpPanelFrom(version, name, common.SourceData).
// ok=false quando não há nenhuma row com texto.
func BuildHelpEntryDTOFrom(name string, version common.GameVersion, panel *helpfile.HelpKeyedStringFile) (dto.FileEntry, bool) {
	if panel == nil {
		return dto.FileEntry{}, false
	}
	return buildHelpEntry(name, version, panel, SortedLocalizationKeys())
}

// buildHelpEntry converte os segmentos localizados do painel em rows do DTO.
// O Index é a posição do ponteiro (reconstrução posicional).
func buildHelpEntry(name string, version common.GameVersion, panel *helpfile.HelpKeyedStringFile, langs []string) (dto.FileEntry, bool) {
	segments := panel.LocalizedStrings()
	entry := dto.FileEntry{
		Metadata: dto.NewHelpMetadata(name, "help/"+helpfile.HelpEntryDir(name), version),
		Rows:     make([]dto.TextRow, 0, len(segments)),
	}
	for i, seg := range segments {
		if seg == nil {
			continue
		}
		text := make(map[string]string, len(langs))
		for _, lang := range langs {
			text[lang] = seg.GetLocalizedString(lang)
		}
		entry.Rows = append(entry.Rows, dto.TextRow{
			Index: i,
			Hash:  hash.Texts(text),
			Text:  text,
		})
	}
	if len(entry.Rows) == 0 {
		common.LogVerbose("No strings found for help %s, skipping", name)
		return dto.FileEntry{}, false
	}
	dto.SortRows(entry.Rows)
	entry.Metadata = entry.Metadata.WithRowCount(len(entry.Rows))
	return entry, true
}

// DedupHelpDTO é o wrapper de help para o DedupDTO compartilhado: escopo
// global dos 6 painéis na ordem SortedKeys, igual à posição de def do
// artefato exportado (arquivo único).
func DedupHelpDTO(c dto.Collection) dto.Collection {
	return DedupDTO(c)
}

// HelpApplyScope é o conjunto de painéis CARREGADOS onde o apply atua — o
// mesmo papel de TableApplyScope para events/eventtable: o escopo é quem
// separa data/ (store global) de .vbf (sessão de cliques).
type HelpApplyScope struct {
	// Ids são os painéis carregados no escopo (índice de propagação).
	Ids []string
	// Panel devolve o painel carregado; nil = não carregado.
	Panel func(name string) *helpfile.HelpKeyedStringFile
	// Save grava o painel tocado em mods/ (intocado = no-op).
	Save func(name string) error
}

// StoreHelpScope monta o escopo do store de data/: os painéis que
// EnsureHelpLoaded carregou, gravados pelo próprio Save de cada painel.
func StoreHelpScope(version common.GameVersion) HelpApplyScope {
	return HelpApplyScope{
		Ids: helpfile.HelpEntryNames(),
		Panel: func(name string) *helpfile.HelpKeyedStringFile {
			return helpfile.GetHelp(version, name)
		},
		Save: func(name string) error {
			p := helpfile.GetHelp(version, name)
			if p == nil {
				return nil
			}
			return p.Save()
		},
	}
}

// ApplyHelpDTO aplica o DTO de volta nos binários (parse DTO → binário).
//
// Com dedup, o algoritmo é único para UI e import, em 4 fases — todas as
// decisões tomadas contra o estado PRÉ-aplicação, para o resultado não
// depender da ordem das entradas do lote:
//  1. defs do lote: hash → texto literal de 'us' (refs resolvem por aqui;
//     def fora do lote ⇒ identidade — a ref mantém o texto atual);
//  2. índice pré-estado: texto 'us' → todos os segmentos iguais nos 6
//     painéis (a alavanca da propagação anti-desalinhamento);
//  3. coleta: valida a estrutura de cada entrada e agenda as mudanças —
//     um 'us' que difere do estado atual agenda TODAS as cópias dele;
//  4. escrita + save de todo painel tocado (lote ou propagação). Sem
//     alteração, nada é gravado.
func ApplyHelpDTO(version common.GameVersion, c dto.Collection, ids []string) error {
	return ApplyHelpDTOWithScope(version, c, ids, StoreHelpScope(version))
}

// ApplyHelpDTOWithScope é o motor de apply com um escopo montado pelo
// chamador (data/ usa o store; o .vbf usa os painéis da sessão de cliques).
// version documenta a árvore de origem — quem decide o que está carregado é
// o escopo.
func ApplyHelpDTOWithScope(version common.GameVersion, c dto.Collection, ids []string, scope HelpApplyScope) error {
	filter := make(map[string]bool)
	if len(ids) > 0 {
		for _, id := range ids {
			filter[id] = true
		}
	}
	inBatch := func(name string) bool { return len(filter) == 0 || filter[name] }

	// Fase 1 — defs do lote: hash → texto literal de 'us'. Montadas antes
	// de qualquer escrita porque uma ref pode vir antes da def na ordem das
	// entradas. Ref não vira def (o valor dela é $hash, não texto).
	defs := make(map[string]string)
	for _, name := range c.SortedKeys() {
		if !inBatch(name) {
			continue
		}
		for _, row := range c[name].Rows {
			if row.Divergent {
				// A cópia divergida não é definição do ponteiro: as refs
				// devem continuar resolvendo para a def canônica do lote.
				continue
			}
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

	// Fase 2 — índice pré-estado: texto 'us' → todas as cópias (os 6
	// painéis). É por ele que cada edição enxerga as cópias ANTES de
	// qualquer escrita: um texto novo igual ao velho de outra entrada não
	// arrasta alvos alheios nem reverte rascunho de terceiros.
	index := make(map[string][]helpTarget)
	for _, name := range scope.Ids {
		p := scope.Panel(name)
		if p == nil {
			continue
		}
		file := p.Files[common.DefaultLocalization]
		if file == nil {
			continue
		}
		for _, seg := range file.Segments {
			if seg == nil {
				continue
			}
			index[seg.Text] = append(index[seg.Text], helpTarget{panel: name, seg: seg})
		}
	}

	// Fase 2½ — divergentes do lote: cada segmento 'us' com edição
	// DIVERGENTE é self-target — grava só nele. Anotados no pré-estado
	// para a expansão de grupo das edições normais filtrá-los: a def
	// editada com A não pode tocar a cópia que divergiu com B, em
	// nenhuma ordem de escrita do lote.
	divergent := divergentHelpSegments(c, filter, scope, defs)

	// Fase 3 — coleta: valida tudo do lote antes de escrever qualquer coisa
	// (um DTO malformado não deixa meio painel aplicado).
	var failed []string
	var applied []string
	var usChanges, dirChanges []helpChange
	for _, name := range c.SortedKeys() {
		if !inBatch(name) {
			continue
		}
		panel := scope.Panel(name)
		if panel == nil {
			common.LogError("failed to update help %s: painel não carregado", name)
			failed = append(failed, name)
			continue
		}
		us, dir, err := collectHelpChanges(name, panel, c[name], defs, index, divergent)
		if err != nil {
			common.LogError("failed to update help %s: %v", name, err)
			failed = append(failed, name)
			continue
		}
		usChanges = append(usChanges, us...)
		dirChanges = append(dirChanges, dir...)
		applied = append(applied, name)
	}
	if len(applied) == 0 && len(failed) == 0 {
		return fmt.Errorf("no matching help panels in DTO to apply")
	}

	// Fase 4 — escrita: alvos congelados no pré-estado; conflitos (mesmo
	// texto velho editado de formas diferentes) resolvem por último da ordem
	// do lote (SortedKeys), mantendo o grupo inteiro alinhado.
	saveSet := make(map[string]bool)
	for _, ch := range dirChanges {
		for _, t := range ch.targets {
			if ch.newText == t.seg.Text {
				continue
			}
			t.seg.SetText(ch.newText)
			saveSet[t.panel] = true
		}
	}
	for _, ch := range usChanges {
		for _, t := range ch.targets {
			if ch.newText == t.seg.Text {
				continue
			}
			t.seg.SetText(ch.newText)
			saveSet[t.panel] = true
		}
	}

	// Save: só o que recebeu texto novo (lote ou propagação). Save de cada
	// painel grava apenas as localizações alteradas — intocados são no-op.
	if len(saveSet) > 0 {
		for _, name := range scope.Ids {
			if !saveSet[name] {
				continue
			}
			if err := scope.Save(name); err != nil {
				common.LogError("failed to save help %s: %v", name, err)
				failed = append(failed, name)
			}
		}
	}
	if len(failed) > 0 {
		sort.Strings(failed)
		return fmt.Errorf("failed to update %d help panel(s): %v", len(failed), failed)
	}
	common.LogVerbose("Help panels processed successfully! (%d panels)", len(applied))
	return nil
}

// helpTarget é um segmento 'us' alvo de escrita, com seu painel.
type helpTarget struct {
	panel string
	seg   *helpfile.HelpSegment
}

// helpChange é uma mudança agendada: todos os segmentos que devem passar a
// ter newText (em 'us', as cópias idênticas do estado pré-aplicação).
type helpChange struct {
	targets []helpTarget
	newText string
}

// collectHelpChanges valida a entrada e agenda as mudanças dela contra o
// estado pré-aplicação (index). Não escreve nada.
func collectHelpChanges(name string, panel *helpfile.HelpKeyedStringFile, entry dto.FileEntry, defs map[string]string, index map[string][]helpTarget, divergent map[*helpfile.HelpSegment]bool) (usChanges, dirChanges []helpChange, err error) {
	dto.SortRows(entry.Rows)
	count := panel.SegmentCount()

	// Fase 1 — validação estrutural.
	for _, row := range entry.Rows {
		if row.Index < 0 || row.Index >= count {
			return nil, nil, fmt.Errorf("string index out of range for help %s: %d", name, row.Index)
		}
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				common.LogVerbose("unsupported localization %s for help %s[%d]", lang, name, row.Index)
				continue
			}
			file := panel.Files[lang]
			if file == nil {
				return nil, nil, fmt.Errorf("help %s sem arquivo para %s", name, lang)
			}
			if row.Index >= len(file.Segments) || file.Segments[row.Index] == nil {
				return nil, nil, fmt.Errorf("segmento %d ausente no help %s", row.Index, name)
			}
		}
	}

	// Fase 2 — agendamento contra o pré-estado.
	for _, row := range entry.Rows {
		for lang, newText := range row.Text {
			if newText == "" {
				continue
			}
			if _, ok := common.SupportedLanguages[lang]; !ok {
				continue
			}
			seg := panel.Files[lang].Segments[row.Index]

			if lang != common.DefaultLocalization {
				// Idiomas ≠ us nunca participam do dedup nem da
				// propagação: escrevem direto quando mudam.
				if newText != seg.Text {
					dirChanges = append(dirChanges, helpChange{
						targets: []helpTarget{{panel: name, seg: seg}},
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
			if newText == seg.Text {
				continue // sem mudança no segmento desta row: nada a agendar
			}
			if row.Divergent {
				// Divergência: grava só NESTE segmento, sem expandir às
				// cópias iguais e sem ser sobrescrita pela def no mesmo lote.
				usChanges = append(usChanges, helpChange{
					targets: []helpTarget{{panel: name, seg: seg}},
					newText: newText,
				})
				continue
			}
			targets := index[seg.Text]
			if len(targets) == 0 {
				targets = []helpTarget{{panel: name, seg: seg}}
			} else if len(divergent) > 0 {
				targets = excludeDivergentHelp(targets, divergent)
				if len(targets) == 0 {
					continue
				}
			}
			usChanges = append(usChanges, helpChange{targets: targets, newText: newText})
		}
	}
	return usChanges, dirChanges, nil
}

// divergentHelpSegments anota os segmentos 'us' com edição divergente no
// lote. As edições normais filtram esses alvos da expansão por texto para
// que uma mudança na def não sobrescreva uma divergência A/B.
func divergentHelpSegments(c dto.Collection, filter map[string]bool, scope HelpApplyScope, defs map[string]string) map[*helpfile.HelpSegment]bool {
	inBatch := func(id string) bool { return len(filter) == 0 || filter[id] }
	var divergent map[*helpfile.HelpSegment]bool
	for _, id := range c.SortedKeys() {
		if !inBatch(id) {
			continue
		}
		panel := scope.Panel(id)
		if panel == nil {
			continue
		}
		file := panel.Files[common.DefaultLocalization]
		if file == nil {
			continue
		}
		for _, row := range c[id].Rows {
			if !row.Divergent || row.Index < 0 || row.Index >= len(file.Segments) {
				continue
			}
			seg := file.Segments[row.Index]
			if seg == nil {
				continue
			}
			newText := row.Text[common.DefaultLocalization]
			if newText == "" {
				continue
			}
			if bare, isRef := refBare(row, newText); isRef {
				resolved, ok := defs[bare]
				if !ok {
					continue
				}
				newText = resolved
			}
			if newText == seg.Text {
				continue
			}
			if divergent == nil {
				divergent = make(map[*helpfile.HelpSegment]bool)
			}
			divergent[seg] = true
		}
	}
	return divergent
}

func excludeDivergentHelp(targets []helpTarget, divergent map[*helpfile.HelpSegment]bool) []helpTarget {
	out := make([]helpTarget, 0, len(targets))
	for _, target := range targets {
		if !divergent[target.seg] {
			out = append(out, target)
		}
	}
	return out
}
