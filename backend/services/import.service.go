package services

// Importação em lote (.json/.strings): parse → detecção de kind/versão →
// store → validação por entrada → merge do 'us' → capacidade uint16.
//
// Divisão de erros:
//   - arquivo (ilegível, extensão, key ausente/inválida, kind ou versão
//     divergentes, kinds misturados) → error (toast no frontend);
//   - por entrada e de capacidade → summary.Errors (modal desabilita Importar).
//
// Capacidade uint16 SEMPRE após a conversão das tags de controle em bytes:
//   - events: rebuild real em cópias (RebuildFieldStrings deduplica offsets);
//   - objects: texto medido direto por converter.StringToByteSize (mesma
//     conversão do FillByteList do rebuild, sem instâncias auxiliares);
//   - macro: rebuild real do dicionário inteiro (layout do container).
//
// A store nunca é mutada: events trabalham em cópias de FieldString.

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/lockit"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/hash"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
)

// importUint16Limit é o teto de bytes por binário: os offsets são uint16
// (0..65535) e a string table começa em 0 → 65536 bytes é o máximo gravável.
const importUint16Limit = 65536

// importPlan é o resultado do prepareImport: entradas válidas mescladas e o
// resumo para o modal.
type importPlan struct {
	kind    string
	merged  dto.Collection // store clonada + 'us' do import (alvo do apply)
	summary dto.ImportSummary
}

// ---- helpers de rows ---------------------------------------------------------

// rowKeyOf é a identidade de row: (Index, Name) — mesmando a chave
// "index\x00name" do strings parser.
func rowKeyOf(r dto.TextRow) string {
	return fmt.Sprintf("%d\x00%s", r.Index, r.Name)
}

// usKeysOf mapeia rowKey → texto 'us' não vazio. O formato .strings descarta
// rows sem texto: a comparação é por este conjunto, nunca por contagem bruta
// de rows/índices.
func usKeysOf(rows []dto.TextRow) map[string]string {
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		if t := r.Text[common.DefaultLocalization]; t != "" {
			out[rowKeyOf(r)] = t
		}
	}
	return out
}

// rowIndexCount conta índices distintos (index_count do resumo).
func rowIndexCount(rows []dto.TextRow) int {
	seen := make(map[int]bool, len(rows))
	for _, r := range rows {
		seen[r.Index] = true
	}
	return len(seen)
}

// importLanguages devolve (ordem estável) os idiomas com algum texto no arquivo.
func importLanguages(c dto.Collection) []string {
	seen := make(map[string]bool)
	for _, id := range c.SortedKeys() {
		for _, r := range c[id].Rows {
			for lang, t := range r.Text {
				if t != "" {
					seen[lang] = true
				}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for lang := range seen {
		out = append(out, lang)
	}
	sort.Strings(out)
	return out
}

// validateImportEntry confere a entrada importada contra a store.
// Regras (qualquer erro bloqueia a importação inteira):
//   - row keys do import ⊆ row keys da store (índice novo não cabe no binário);
//   - conjuntos de rows com texto 'us' não vazio são iguais (.strings descarta
//     rows vazias, então a igualdade é por conteúdo, não por contagem);
//   - changed = rows com 'us' presente nos dois lados e texto diferente.
func validateImportEntry(importedEntry, storeEntry dto.FileEntry) (changed int, errs []string) {
	storeRows := make(map[string]dto.TextRow, len(storeEntry.Rows))
	for _, r := range storeEntry.Rows {
		storeRows[rowKeyOf(r)] = r
	}
	for _, r := range importedEntry.Rows {
		if _, ok := storeRows[rowKeyOf(r)]; !ok {
			errs = append(errs, fmt.Sprintf(
				"índice novo no arquivo (index %d, name %q) — o binário não aceita entradas novas",
				r.Index, r.Name))
		}
	}
	importUS := usKeysOf(importedEntry.Rows)
	storeUS := usKeysOf(storeEntry.Rows)
	missing := 0
	for rk := range storeUS {
		if _, ok := importUS[rk]; !ok {
			missing++
		}
	}
	extra := 0
	for rk := range importUS {
		if _, ok := storeUS[rk]; !ok {
			extra++
		}
	}
	if missing > 0 {
		errs = append(errs, fmt.Sprintf("%d texto(s) em inglês ausente(s) no arquivo", missing))
	}
	if extra > 0 {
		errs = append(errs, fmt.Sprintf("%d texto(s) em inglês no arquivo sem par na store", extra))
	}
	for rk, want := range importUS {
		if have, ok := storeUS[rk]; ok && have != want {
			changed++
		}
	}
	return changed, errs
}

// mergeImportUs clona a store e sobrepõe apenas o texto 'us' vindo do import.
// Idiomas não-us e placeholders vêm integralmente da store: o import só
// contribui com inglês (os demais idiomas do arquivo são descartados).
func mergeImportUs(store, imported dto.Collection) dto.Collection {
	out := make(dto.Collection, len(store))
	for id, se := range store {
		imp, ok := imported[id]
		if !ok {
			out[id] = se
			continue
		}
		impRows := make(map[string]dto.TextRow, len(imp.Rows))
		for _, r := range imp.Rows {
			impRows[rowKeyOf(r)] = r
		}
		rows := make([]dto.TextRow, len(se.Rows))
		for i, r := range se.Rows {
			nr := r
			if ir, ok := impRows[rowKeyOf(r)]; ok {
				if t := ir.Text[common.DefaultLocalization]; t != "" {
					text := make(map[string]string, len(r.Text)+1)
					for k, v := range r.Text {
						text[k] = v
					}
					h := make(map[string]string, len(r.Hash)+1)
					for k, v := range r.Hash {
						h[k] = v
					}
					text[common.DefaultLocalization] = t
					h[common.DefaultLocalization] = hash.Sum64Hex(t)
					nr.Text = text
					nr.Hash = h
				}
			}
			rows[i] = nr
		}
		out[id] = dto.FileEntry{Metadata: se.Metadata, Rows: rows}
	}
	return out
}

// ---- detecção ----------------------------------------------------------------

// kindFromKey decide o kind pela metadata.key: macrodic.dcp → macro,
// /event/ → events; /gamedata/ps3data/lockit/ → lockit; demais → objects.
func kindFromKey(key string) string {
	lower := strings.ToLower(key)
	if strings.HasSuffix(lower, "macrodic.dcp") {
		return KindMacro
	}
	if strings.Contains(lower, "/event/") {
		return KindEvents
	}
	if lockit.IsLockitKey(key) {
		return KindLockit
	}
	return KindObjects
}

// detectImportKind infere o kind e valida a versão pelas metadata.key.
// A versão passa por VersionPathName: ffx2 e lastmiss dividem a árvore
// ("ffx2"), então um export de lastmiss importa na aba lastmiss (e vice-versa).
// Falha aqui é de arquivo inteiro (toast); erros por entrada vão no resumo.
func detectImportKind(active common.GameVersion, imported dto.Collection) (string, error) {
	kind := ""
	for _, id := range imported.SortedKeys() {
		key := strings.TrimSpace(imported[id].Metadata.Key)
		if key == "" {
			return "", fmt.Errorf("entrada %q sem metadata.key (arquivo não gerado por este app?)", id)
		}
		p, ok := dto.ParseKey(key)
		if !ok {
			return "", fmt.Errorf("metadata.key inválida na entrada %q: %q", id, key)
		}
		fv, verr := common.ParseGameVersionStrict(p.Version)
		if verr != nil {
			return "", fmt.Errorf("versão desconhecida no arquivo: %q", p.Version)
		}
		if common.VersionPathName(fv) != common.VersionPathName(active) {
			return "", fmt.Errorf("arquivo é da versão %s, mas a aba ativa é %s", p.Version, active.String())
		}
		k := kindFromKey(key)
		if kind == "" {
			kind = k
		} else if k != kind {
			return "", fmt.Errorf("arquivo mistura kinds diferentes (%s e %s)", kind, k)
		}
	}
	if kind == "" {
		return "", fmt.Errorf("arquivo sem entradas")
	}
	return kind, nil
}

// importKnownIDs filtra os ids importados que existem estruturalmente nesta
// versão (régua de events, layout de objects, chunks do macro). Devolve também
// o DTO completo de macro (a capacidade mede o arquivo inteiro).
func (s *MetadataService) importKnownIDs(kind string, version common.GameVersion, imported dto.Collection) (map[string]bool, dto.Collection, error) {
	known := make(map[string]bool)
	var macroFull dto.Collection
	switch kind {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return nil, nil, err
		}
		visible := make(map[string]bool)
		for _, id := range filterEventIDsForVersion(event.GetAllEventIDs(version), version) {
			visible[id] = true
		}
		for _, id := range imported.SortedKeys() {
			if visible[id] && event.GetEvent(version, id) != nil {
				known[id] = true
			}
		}
	case KindObjects:
		for _, id := range imported.SortedKeys() {
			if _, ok := objectKeyForID(version, id); ok {
				known[id] = true
			}
		}
	case KindLockit:
		for _, id := range imported.SortedKeys() {
			if _, ok := lockit.LayoutForID(version, id); ok {
				known[id] = true
			}
		}
	case KindMacro:
		if version == common.GameVersionLastMiss {
			// Last Mission não tem dicionário próprio: store vazia e todas
			// as entradas importadas entram no resumo como desconhecidas.
			return known, nil, nil
		}
		full, err := builders.BuildMacroDTO(version)
		if err != nil {
			return nil, nil, err
		}
		macroFull = full
		for _, id := range imported.SortedKeys() {
			if _, ok := full[id]; ok {
				known[id] = true
			}
		}
	}
	return known, macroFull, nil
}

// loadImportStore carrega a Collection da store só para os ids conhecidos
// (ids vazio significaria "tudo" no GetCollection — hence o guarda).
func (s *MetadataService) loadImportStore(kind string, version common.GameVersion, known map[string]bool, macroFull dto.Collection) (dto.Collection, error) {
	if len(known) == 0 {
		return dto.Collection{}, nil
	}
	ids := make([]string, 0, len(known))
	for id := range known {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if kind == KindMacro {
		out := make(dto.Collection, len(ids))
		for _, id := range ids {
			if e, ok := macroFull[id]; ok {
				out[id] = e
			}
		}
		return out, nil
	}
	return s.GetCollection(kind, version, ids)
}

// ---- capacidade uint16 (pós-rebuild, só leitura) ------------------------------

// importEventsUsage mede o binário do evento em cópias: espelha
// extractFieldStringsForLanguage (pula obj nil, placeholder p/ us nil), aplica
// o 'us' importado nas cópias e SEMPRE roda RebuildFieldStrings — que converte
// tags em bytes e deduplica offsets — só descartando o resultado. A store
// nunca é tocada.
func importEventsUsage(version common.GameVersion, id string, imp dto.FileEntry) (dto.ImportUsage, error) {
	usage := dto.ImportUsage{ID: id, Kind: KindEvents, Limit: importUint16Limit}
	ev := event.GetEvent(version, id)
	if ev == nil {
		return usage, fmt.Errorf("evento não encontrado: %s", id)
	}
	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
	importUS := make(map[int]string, len(imp.Rows))
	for _, r := range imp.Rows {
		if r.Name != "" {
			continue // events não têm Name; linha estranha já é bloqueada na validação
		}
		if t := r.Text[common.DefaultLocalization]; t != "" {
			importUS[r.Index] = t
		}
	}
	copies := make([]*event.FieldString, 0, len(ev.Strings))
	for i, obj := range ev.Strings {
		if obj == nil {
			continue
		}
		var c *event.FieldString
		if fs := obj.GetLocalizedContent(common.DefaultLocalization); fs != nil {
			cp := *fs
			c = &cp
		} else {
			c = event.NewEmptyFieldString(charset, version)
		}
		if t, ok := importUS[i]; ok {
			c.SetRegularString(t)
		}
		copies = append(copies, c)
	}
	content := event.RebuildFieldStrings(copies, charset, version)
	header := len(copies) * 8 // first = count*8, precisa caber em uint16
	usage.Used = header + len(content)
	over := header > 65535 || header+len(content) > importUint16Limit
	for _, c := range copies {
		if c.RegularOffset > 65535 || c.SimplifiedOffset > 65535 {
			over = true
			break
		}
	}
	usage.Over = over
	return usage, nil
}

// importObjectsUsage mede a string table só com o texto: para cada segmento
// 'us', o texto do import (override por ponteiro de conteúdo) senão o da
// store, medido por converter.StringToByteSize — que é byte a byte o que o
// FillByteList do RebuildKeyedStrings gravaria (terminador incluso; paridade
// coberta por teste no converter). Sem wrappers de IGlobalKeyedString, sem
// buffer auxiliar e sem gravar arquivo. A iteração espelha o collect do
// rebuild: RangeIndex + GetLocalizedKeyedStrings("us"), pulando nil.
func importObjectsUsage(version common.GameVersion, id string, imp dto.FileEntry) (dto.ImportUsage, error) {
	usage := dto.ImportUsage{ID: id, Kind: KindObjects, Limit: importUint16Limit}
	key, ok := objectKeyForID(version, id)
	if !ok {
		return usage, fmt.Errorf("objeto desconhecido: %s", id)
	}
	layout, ok := objectsfile.FileLayouts[key]
	if !ok {
		return usage, fmt.Errorf("layout não encontrado: %s", key)
	}
	binFile, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		return usage, err
	}
	objects := binFile.GetObjects()
	if objects == nil || objects.Len() == 0 {
		return usage, fmt.Errorf("sem objetos: %s", id)
	}

	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
	overrides := make(map[datastore.IGlobalKeyedString]string, len(imp.Rows))
	for _, r := range imp.Rows {
		t := r.Text[common.DefaultLocalization]
		if t == "" {
			continue
		}
		obj := objects.Get(r.Index)
		if obj == nil {
			return usage, fmt.Errorf("objeto %d ausente", r.Index)
		}
		seg := obj.GetKeyedString(r.Name)
		if seg == nil {
			return usage, fmt.Errorf("segmento %q ausente no objeto %d", r.Name, r.Index)
		}
		content := seg.GetLocalizedContent(common.DefaultLocalization)
		if content == nil {
			return usage, fmt.Errorf("segmento %q sem conteúdo 'us' no objeto %d", r.Name, r.Index)
		}
		overrides[content] = t
	}

	var merr error
	pos := 0
	objects.RangeIndex(func(_ int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil || merr != nil {
			return
		}
		for _, ks := range obj.GetLocalizedKeyedStrings(common.DefaultLocalization) {
			if ks == nil {
				continue
			}
			text, has := overrides[ks]
			if !has {
				text = ks.GetString()
			}
			n, err := converter.StringToByteSize(text, charset, version)
			if err != nil {
				merr = err
				return
			}
			pos += n
		}
	})
	if merr != nil {
		return usage, merr
	}
	usage.Used = pos
	// Offset de cada string é sua posição inicial < Used; Used ≤ 65536 ⟺
	// todo offset ≤ 65535 (cada string tem ao menos o terminador).
	usage.Over = pos > importUint16Limit
	return usage, nil
}

// importMacroUsage reconstrói o dicionário inteiro (todos os chunks, todos os
// idiomas) com o 'us' mesclado e mede o container 'us'. Um uso só para o
// arquivo: o limite é por binário. O layout (header + chunks) não é derivável
// do texto puro, por isso o rebuild real.
func importMacroUsage(version common.GameVersion, macroFull, valid dto.Collection) (dto.ImportUsage, error) {
	usage := dto.ImportUsage{Kind: KindMacro, Limit: importUint16Limit}
	containers, err := builders.RebuildMacroContainers(version, mergeImportUs(macroFull, valid))
	if err != nil {
		return usage, err
	}
	us := containers[common.DefaultLocalization]
	if us == nil || len(us.Bytes) == 0 {
		return usage, fmt.Errorf("rebuild do dicionário sem conteúdo 'us'")
	}
	usage.Used = len(us.Bytes)
	usage.Over = usage.Used > importUint16Limit
	return usage, nil
}

// ---- pipeline -----------------------------------------------------------------

// prepareImport executa o pipeline SEM aplicar: parse → detecção → store →
// validação por entrada → merge do 'us' → capacidades. Erros de arquivo voltam
// como error (toast); erros por entrada/capacidade ficam em summary.Errors.
func (s *MetadataService) prepareImport(path string, version common.GameVersion) (*importPlan, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("nenhum arquivo selecionado")
	}
	raw, err := common.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("arquivo vazio: %s", filepath.Base(path))
	}

	var imported dto.Collection
	format := ""
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".json":
		format = "json"
		imported, err = jsonfmt.NewJSONEventsFormatter().Unmarshal(raw)
	case ".strings":
		format = "strings"
		imported, err = strfmt.NewStringsFormatter().Unmarshal(raw)
	default:
		return nil, fmt.Errorf("extensão não suportada: %s (use .json ou .strings)", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar o arquivo: %w", err)
	}
	if len(imported) == 0 {
		return nil, fmt.Errorf("arquivo sem entradas: %s", filepath.Base(path))
	}

	kind, err := detectImportKind(version, imported)
	if err != nil {
		return nil, err
	}
	if err := ensureVersionReady(version); err != nil {
		return nil, err
	}
	known, macroFull, err := s.importKnownIDs(kind, version, imported)
	if err != nil {
		return nil, err
	}
	store, err := s.loadImportStore(kind, version, known, macroFull)
	if err != nil {
		return nil, err
	}

	summary := dto.ImportSummary{
		Path:      path,
		Format:    format,
		Kind:      kind,
		Version:   version.String(),
		Languages: importLanguages(imported),
		// Toda importação confirmada reconstrói e grava os binários em mods/;
		// o aviso no modal deixa o efeito explícito antes de aplicar.
		SavesBinary: true,
		Entries:     make([]dto.ImportEntryInfo, 0, len(imported)),
		Usages:      make([]dto.ImportUsage, 0, len(known)),
		Errors:      []string{},
	}

	valid := make(dto.Collection, len(known))
	var totalChanged, totalIndices int
	for _, id := range imported.SortedKeys() {
		imp := imported[id]
		info := dto.ImportEntryInfo{
			ID:         id,
			Key:        imp.Metadata.Key,
			IndexCount: rowIndexCount(imp.Rows),
		}
		totalIndices += info.IndexCount
		if se, ok := store[id]; ok {
			info.StoreIndexCount = rowIndexCount(se.Rows)
			changed, errs := validateImportEntry(imp, se)
			info.ChangedTexts = changed
			totalChanged += changed
			if len(errs) > 0 {
				info.Error = strings.Join(errs, "; ")
				for _, e := range errs {
					summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %s", id, e))
				}
			} else {
				valid[id] = se
			}
		} else {
			info.Error = "entrada não existe nesta versão"
			summary.Errors = append(summary.Errors, fmt.Sprintf("%s: entrada não existe nesta versão", id))
		}
		summary.Entries = append(summary.Entries, info)
	}
	summary.EntryCount = len(imported)
	summary.TotalIndices = totalIndices
	summary.ChangedTexts = totalChanged

	merged := mergeImportUs(valid, imported)

	// Capacidades sobre as entradas válidas; estouro bloqueia a importação.
	if kind == KindMacro {
		if len(merged) > 0 && macroFull != nil {
			usage, uerr := importMacroUsage(version, macroFull, valid)
			if uerr != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("dicionário: %s", uerr.Error()))
			} else {
				summary.Usages = append(summary.Usages, usage)
				if usage.Over {
					summary.Errors = append(summary.Errors, fmt.Sprintf(
						"dicionário: texto excede o limite binário (%d > %d bytes)",
						usage.Used, usage.Limit))
				}
			}
		}
	} else if kind == KindLockit {
		// O lockit é uma lista CRLF sem offsets uint16 nem cabeçalho; não há
		// limite de capacidade a medir (o import grava apenas o us).
	} else {
		for _, id := range merged.SortedKeys() {
			var usage dto.ImportUsage
			var uerr error
			if kind == KindEvents {
				usage, uerr = importEventsUsage(version, id, imported[id])
			} else {
				usage, uerr = importObjectsUsage(version, id, imported[id])
			}
			if uerr != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %s", id, uerr.Error()))
				continue
			}
			summary.Usages = append(summary.Usages, usage)
			if usage.Over {
				summary.Errors = append(summary.Errors, fmt.Sprintf(
					"%s: texto excede o limite binário (%d > %d bytes)",
					id, usage.Used, usage.Limit))
			}
		}
	}

	return &importPlan{kind: kind, merged: merged, summary: summary}, nil
}

// PreviewImport parseia e valida o arquivo (sem aplicar) e devolve o resumo
// para o modal de confirmação.
func (s *MetadataService) PreviewImport(path string, version common.GameVersion) (dto.ImportSummary, error) {
	plan, err := s.prepareImport(path, version)
	if err != nil {
		return dto.ImportSummary{}, err
	}
	return plan.summary, nil
}

// ImportFile reexecuta o pipeline e aplica o 'us' mesclado no binário.
// Qualquer erro de validação/capacidade bloqueia a importação inteira.
// Devolve quantos textos em inglês mudaram em relação à store.
func (s *MetadataService) ImportFile(path string, version common.GameVersion) (int, error) {
	plan, err := s.prepareImport(path, version)
	if err != nil {
		return 0, err
	}
	if len(plan.summary.Errors) > 0 {
		return 0, fmt.Errorf("importação bloqueada por %d erro(s) de validação", len(plan.summary.Errors))
	}
	if len(plan.merged) == 0 {
		return 0, fmt.Errorf("nada a importar")
	}
	if err := s.ApplyTextCollection(plan.kind, version, plan.merged); err != nil {
		return 0, err
	}
	return plan.summary.ChangedTexts, nil
}

// ExportJSON escreve os artefatos JSON do lote em mods/edits (ids vazio =
// tudo; langs nil/vazio = todos os idiomas). O caminho do arquivo é o mesmo
// usado na importação (seletor nativo).
func (s *MetadataService) ExportJSON(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	c, err := s.GetCollection(kind, version, ids)
	if err != nil {
		return nil, err
	}
	switch kind {
	case KindEvents:
		p, jerr := jsonfmt.NewJSONEventsFormatter().WriteEvents(c, version, langs)
		if jerr != nil {
			return nil, jerr
		}
		return []string{p}, nil
	case KindObjects:
		return jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
	case KindMacro:
		p, jerr := jsonfmt.NewJSONMacroFormatter().WriteMacro(c, version, langs)
		if jerr != nil {
			return nil, jerr
		}
		return []string{p}, nil
	case KindLockit:
		return jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

// CountChangedTexts devolve quantas rows têm 'us' diferente entre import e
// store (utilitário de conferência para o frontend).
func (s *MetadataService) CountChangedTexts(imported, store dto.Collection) int {
	total := 0
	for _, id := range imported.SortedKeys() {
		se, ok := store[id]
		if !ok {
			continue
		}
		changed, _ := validateImportEntry(imported[id], se)
		total += changed
	}
	return total
}
