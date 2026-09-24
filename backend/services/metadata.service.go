package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
)

// Kinds lógicos de texto servidos ao frontend.
const (
	KindEvents  = "events"
	KindObjects = "objects"
	KindMacro   = "macro"
)

// LastMissionShortened é o prefixo (eventID[:2]) do grupo de events da
// Last Mission, que mora na árvore do ffx2: "lm".
const LastMissionShortened = "lm"

// EntrySummary é o índice leve para sidebar/tree: id + key, sem rows.
// row_count aparece nas entradas (GetEntry/GetCollection), nunca aqui.
type EntrySummary struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// MetadataService é o acesso do frontend (via Wails) a metadata e texto.
// Devolve DTO estruturado (dto.Collection/FileEntry) direto da memória —
// sem exportar para disco e reler o arquivo. O formato JSON de export
// continua intocado; este service é o "formato frontend".
type MetadataService struct {
	notifier INotificationService
}

func NewMetadataService(notifier INotificationService) *MetadataService {
	return &MetadataService{notifier: notifier}
}

// Cache de inicialização por versão: charsets + macros carregados uma única
// vez (lazy), sem depender de chamada manual de InitializeInternals.
var (
	readyMu       sync.Mutex
	readyVersions = map[common.GameVersion]error{}
)

// ensureVersionReady garante charsets e macros da versão antes de qualquer
// leitura de texto. lastmiss reaproveita o bucket de ffx2 (CharsetVersion).
func ensureVersionReady(version common.GameVersion) error {
	v := common.CharsetVersion(version)
	readyMu.Lock()
	defer readyMu.Unlock()
	if err, done := readyVersions[v]; done {
		return err
	}
	err := reader.PrepareVersion(v)
	readyVersions[v] = err
	return err
}

// ---- metadata por key/id/path ---------------------------------------------

// GetMetadata resolve qualquer uma das 3 formas e devolve a metadata nova
// (key, id, is_dir; row_count omitido — só existe no export).
// Formas: metadata.key (ffx/...), id de collection (azit0000, command,
// chunk_00) ou caminho em disco (absoluto ou com ffx_ps2/).
func (s *MetadataService) GetMetadata(query string) (dto.Metadata, error) {
	key, id, isDir, err := s.Resolve(query)
	if err != nil {
		return dto.Metadata{}, err
	}
	return dto.Metadata{Key: key, ID: id, IsDir: isDir}, nil
}

// Resolve devolve (key, id, isDir) para as 3 formas de consulta.
func (s *MetadataService) Resolve(query string) (key, id string, isDir bool, err error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return "", "", false, fmt.Errorf("empty query")
	}
	slash := filepath.ToSlash(filepath.Clean(q))
	// Caminho em disco: absoluto, com ffx_ps2/, com backslash (Windows)
	// ou relativo que existe no filesystem.
	if strings.Contains(slash, "ffx_ps2/") || filepath.IsAbs(q) ||
		strings.Contains(q, "\\") || pathExists(q) {
		return s.resolvePath(q)
	}
	if strings.Contains(slash, "/") {
		return s.resolveKey(slash)
	}
	return s.resolveID(slash)
}

// pathExists informa se o caminho existe (arquivo ou diretório).
func pathExists(q string) bool {
	clean := filepath.Clean(q)
	if _, err := os.Stat(clean); err == nil {
		return true
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return false
	}
	_, err = os.Stat(abs)
	return err == nil
}

// resolveKey trata metadata.key canônica (version/pattern).
func (s *MetadataService) resolveKey(slash string) (string, string, bool, error) {
	p, ok := dto.ParseKey(slash)
	if !ok {
		return "", "", false, fmt.Errorf("invalid key: %s", slash)
	}
	id := p.Stem
	if strings.EqualFold(p.FileName, "macrodic.dcp") {
		// Key compartilhada por todos os chunks: sem id.
		id = ""
	}
	return slash, id, false, nil
}

// resolveID trata id de collection usando a versão ativa como default.
func (s *MetadataService) resolveID(id string) (string, string, bool, error) {
	cur := common.CurrentGameVersion()
	if _, ok := dto.ChunkIndexFromID(id); ok {
		return common.VersionPathName(cur) + "/menu/macrodic.dcp", id, false, nil
	}
	if key, ok := objectKeyForID(cur, id); ok {
		return key, id, false, nil
	}
	if len(id) < 2 {
		return "", "", false, fmt.Errorf("unknown id: %s", id)
	}
	return dto.NewEventMetadata(id, cur).Key, id, false, nil
}

// resolvePath trata caminho em disco: deriva key + id + is_dir.
func (s *MetadataService) resolvePath(q string) (string, string, bool, error) {
	clean := filepath.Clean(q)
	abs, err := filepath.Abs(clean)
	if err != nil {
		abs = clean
	}
	isDir := false
	if st, serr := os.Stat(abs); serr == nil {
		isDir = st.IsDir()
	}
	if isDir {
		// Diretório: sem key de arquivo; id = basename.
		return "", filepath.Base(clean), true, nil
	}
	version := common.CurrentGameVersion()
	if v, verr := common.CheckFFXPath(abs); verr == nil {
		version = v
	}
	slash := filepath.ToSlash(abs)
	locPattern, ok := localizationPatternFromPath(slash)
	if !ok {
		return "", "", false, fmt.Errorf("path is not under a localization root: %s", q)
	}
	key := common.VersionPathName(version) + "/" + locPattern
	id := strings.TrimSuffix(filepath.Base(locPattern), filepath.Ext(locPattern))
	if strings.EqualFold(filepath.Base(locPattern), "macrodic.dcp") {
		id = ""
	}
	return key, id, false, nil
}

// localizationPatternFromPath extrai o trecho após ffx_ps2/<v>/master/new_*pc/.
func localizationPatternFromPath(slashAbs string) (string, bool) {
	idx := strings.Index(slashAbs, "/master/")
	if idx < 0 {
		return "", false
	}
	rest := strings.Trim(slashAbs[idx+len("/master/"):], "/")
	segs := strings.Split(rest, "/")
	if len(segs) < 2 || !strings.HasPrefix(segs[0], "new_") {
		return "", false
	}
	return strings.Join(segs[1:], "/"), true
}

// objectKeyForID procura nos layouts a key do basename (determinístico).
func objectKeyForID(version common.GameVersion, id string) (string, bool) {
	want := id + ".bin"
	var matches []string
	for key, layout := range objectsfile.FileLayouts {
		if layout.Version == version && strings.EqualFold(layout.FileName, want) {
			matches = append(matches, key)
		}
	}
	if len(matches) == 0 {
		return "", false
	}
	sort.Strings(matches)
	return matches[0], true
}

// ExportEntry monta o DTO da entrada (kind/id/version) e escreve os artefatos
// JSON e .strings em mods/edits (caminhos padrão dos formatters). Devolve os
// caminhos escritos. langs nil/vazio = todos os idiomas.
func (s *MetadataService) ExportEntry(kind string, version common.GameVersion, id string, langs []string) ([]string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	c, err := s.GetCollection(kind, version, []string{id})
	if err != nil {
		return nil, err
	}

	var paths []string
	switch kind {
	case KindEvents:
		jp, jerr := jsonfmt.NewJSONEventsFormatter().WriteEvents(c, version, langs)
		if jerr != nil {
			return paths, jerr
		}
		sp, serr := strfmt.NewStringsFormatter().WriteEvents(c, version, langs)
		if serr != nil {
			return paths, serr
		}
		paths = append(paths, jp, sp)
	case KindObjects:
		jps, jerr := jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
		if jerr != nil {
			return paths, jerr
		}
		sps, serr := strfmt.NewStringsFormatter().WriteObjects(c, version, langs)
		if serr != nil {
			return paths, serr
		}
		paths = append(append(paths, jps...), sps...)
	case KindMacro:
		jp, jerr := jsonfmt.NewJSONMacroFormatter().WriteMacro(c, version, langs)
		if jerr != nil {
			return paths, jerr
		}
		sp, serr := strfmt.NewStringsFormatter().WriteMacro(c, version, langs)
		if serr != nil {
			return paths, serr
		}
		paths = append(paths, jp, sp)
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
	return paths, nil
}

// ImportEntry lê o artefato JSON padrão da entrada (mods/edits) e aplica o DTO
// de volta no binário, persistindo. Devolve o caminho lido.
func (s *MetadataService) ImportEntry(kind string, version common.GameVersion, id string) ([]string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if err := ensureVersionReady(version); err != nil {
		return nil, err
	}

	switch kind {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return nil, err
		}
		c, err := s.GetCollection(kind, version, []string{id})
		if err != nil {
			return nil, err
		}
		path, err := jsonfmt.EventsJSONPath(c, version)
		if err != nil {
			return nil, err
		}
		read, err := jsonfmt.NewJSONEventsFormatter().ReadEvents(path)
		if err != nil {
			return nil, err
		}
		if err := builders.ApplyEventsDTO(version, read, []string{id}); err != nil {
			return nil, err
		}
		return []string{path}, nil

	case KindObjects:
		path, err := jsonfmt.ObjectsJSONPath(id, version)
		if err != nil {
			return nil, err
		}
		read, err := jsonfmt.NewJSONObjectFormatter().ReadObjects(path)
		if err != nil {
			return nil, err
		}
		for _, key := range read.SortedKeys() {
			if err := s.applyObjectsEntry(version, key, read[key]); err != nil {
				return nil, err
			}
		}
		return []string{path}, nil

	case KindMacro:
		path, err := jsonfmt.DefaultMacroJSONPath(version)
		if err != nil {
			return nil, err
		}
		read, err := jsonfmt.NewJSONMacroFormatter().ReadMacro(path)
		if err != nil {
			return nil, err
		}
		if err := builders.ApplyMacroDTO(version, read); err != nil {
			return nil, err
		}
		return []string{path}, nil

	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

// ApplyEntry aplica uma entrada editada (DTO) de volta no binário e persiste.
// É o "salvar" do editor in-memory (Ver/Editar).
func (s *MetadataService) ApplyEntry(kind string, version common.GameVersion, id string, entry dto.FileEntry) error {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if err := ensureVersionReady(version); err != nil {
		return err
	}

	switch kind {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return err
		}
		return builders.ApplyEventsDTO(version, dto.Collection{id: entry}, []string{id})
	case KindObjects:
		return s.applyObjectsEntry(version, id, entry)
	case KindMacro:
		// Aplica sobre o DTO completo para não perder os outros chunks.
		c, err := builders.BuildMacroDTO(version)
		if err != nil {
			return err
		}
		c[id] = entry
		return builders.ApplyMacroDTO(version, c)
	default:
		return fmt.Errorf("unknown kind: %s", kind)
	}
}

// macroChunkHasText informa se alguma row do chunk tem texto em algum idioma.
func macroChunkHasText(entry dto.FileEntry) bool {
	for _, row := range entry.Rows {
		for _, text := range row.Text {
			if strings.TrimSpace(text) != "" {
				return true
			}
		}
	}
	return false
}

// ApplyTextCollection aplica um lote de entradas editadas (DTO) e persiste.
// É o "salvar" do editor do frontend: recebe só as entradas com edição, mas
// agrupa por arquivo para reconstruir cada binário uma única vez.
func (s *MetadataService) ApplyTextCollection(kind string, version common.GameVersion, c dto.Collection) error {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if len(c) == 0 {
		return nil
	}
	if err := ensureVersionReady(version); err != nil {
		return err
	}

	switch kind {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return err
		}
		return builders.ApplyEventsDTO(version, c, c.SortedKeys())

	case KindObjects:
		// Aplica as rows de todas as entradas antes de salvar cada binário:
		// entradas repetidas (mesmo arquivo) recebem todas as suas rows.
		byKey := make(map[string][]dto.FileEntry, len(c))
		order := make([]string, 0, len(c))
		for _, id := range c.SortedKeys() {
			key, ok := objectKeyForID(version, id)
			if !ok {
				return fmt.Errorf("unknown object id: %s", id)
			}
			if _, seen := byKey[key]; !seen {
				order = append(order, key)
			}
			byKey[key] = append(byKey[key], c[id])
		}
		for _, key := range order {
			layout, ok := objectsfile.FileLayouts[key]
			if !ok {
				layout, ok = objectsfile.FileLayoutFor(version, key)
				if !ok {
					return fmt.Errorf("no object layout for key: %s", key)
				}
			}
			binFile, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
			if err != nil {
				return err
			}
			for _, entry := range byKey[key] {
				if err := builders.ApplyObjectsEntry(binFile.GetObjects(), version, key, entry); err != nil {
					return err
				}
			}
			if err := binFile.SaveToBinary(layout.PatternPath()); err != nil {
				return err
			}
		}
		return nil

	case KindMacro:
		// Aplica sobre o DTO completo para não perder os outros chunks.
		full, err := builders.BuildMacroDTO(version)
		if err != nil {
			return err
		}
		for id, entry := range c {
			full[id] = entry
		}
		return builders.ApplyMacroDTO(version, full)

	default:
		return fmt.Errorf("unknown kind: %s", kind)
	}
}

// applyObjectsEntry carrega o binário do layout, aplica a entrada e salva.
func (s *MetadataService) applyObjectsEntry(version common.GameVersion, id string, entry dto.FileEntry) error {
	key, ok := objectKeyForID(version, id)
	if !ok {
		return fmt.Errorf("unknown object id: %s", id)
	}
	layout, ok := objectsfile.FileLayouts[key]
	if !ok {
		layout, ok = objectsfile.FileLayoutFor(version, id+".bin")
		if !ok {
			return fmt.Errorf("no object layout for id: %s", id)
		}
	}
	binFile, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
	if err != nil {
		return err
	}
	if err := builders.ApplyObjectsEntry(binFile.GetObjects(), version, key, entry); err != nil {
		return err
	}
	return binFile.SaveToBinary(layout.PatternPath())
}

// ---- texto em memória (DTO estruturado, sem disco) --------------------------

// filterEventIDsForVersion aplica a régua de events por versão, porque o
// lastmiss é expansão do ffx2 e lê a MESMA árvore de eventos:
//   - ffx2: esconde o grupo "lm" — é conteúdo de Last Mission;
//   - lastmiss: mostra só o grupo "lm" (lmdn*/lmev*/lmtuto*/lmys*);
//   - ffx: inalterado (o "lm" do FFX, lmyt*, é texto dessa versão).
func filterEventIDsForVersion(ids []string, version common.GameVersion) []string {
	if version != common.GameVersionFFX2 && version != common.GameVersionLastMiss {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		isLM := len(id) >= 2 && strings.EqualFold(id[:2], LastMissionShortened)
		if version == common.GameVersionLastMiss {
			if isLM {
				out = append(out, id)
			}
			continue
		}
		if !isLM {
			out = append(out, id)
		}
	}
	return out
}

// ListEntries devolve o índice leve (id + key, sem rows) para montar
// sidebar/tree. Barato por design: não carrega binários de objects.
func (s *MetadataService) ListEntries(kind string, version common.GameVersion) ([]EntrySummary, error) {
	if err := ensureVersionReady(version); err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return nil, err
		}
		ids := filterEventIDsForVersion(event.GetAllEventIDs(version), version)
		sort.Strings(ids)
		out := make([]EntrySummary, 0, len(ids))
		for _, id := range ids {
			out = append(out, EntrySummary{ID: id, Key: dto.NewEventMetadata(id, version).Key})
		}
		return out, nil
	case KindObjects:
		var keys []string
		for key, layout := range objectsfile.FileLayouts {
			if layout.Version == version {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		out := make([]EntrySummary, 0, len(keys))
		for _, key := range keys {
			layout := objectsfile.FileLayouts[key]
			id := strings.TrimSuffix(layout.FileName, filepath.Ext(layout.FileName))
			out = append(out, EntrySummary{ID: id, Key: key})
		}
		return out, nil
	case KindMacro:
		if version == common.GameVersionLastMiss {
			// Last Mission não tem dicionário próprio: não servir os chunks
			// do macrodic de FFX-2 (apenas retorna vazio, sem erro).
			return []EntrySummary{}, nil
		}
		c, err := builders.BuildMacroDTO(version)
		if err != nil {
			return nil, err
		}
		out := make([]EntrySummary, 0, len(c))
		for _, k := range c.SortedKeys() {
			if !macroChunkHasText(c[k]) {
				// Chunk sem texto: não interessa ao tradutor; oculto no frontend.
				continue
			}
			out = append(out, EntrySummary{ID: k, Key: c[k].Metadata.Key})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

// GetEntry devolve uma entrada completa (metadata + rows) por demanda.
func (s *MetadataService) GetEntry(kind, id string, version common.GameVersion) (dto.FileEntry, error) {
	c, err := s.GetCollection(kind, version, []string{id})
	if err != nil {
		return dto.FileEntry{}, err
	}
	entry, ok := c[id]
	if !ok {
		return dto.FileEntry{}, fmt.Errorf("%s entry not found: %s", kind, id)
	}
	return entry, nil
}

// ListLanguages devolve os idiomas disponíveis em formato chave/valor
// (Code para arquivos, Name para exibição). Todas as versões usam
// os mesmos idiomas.
func (s *MetadataService) ListLanguages() []common.Language {
	return common.AvailableLanguages()
}

// normalizeIDs aceita ids ou keys: item com "/" ou "\" (metadata.key ou
// caminho) é resolvido para id via Resolve; id puro passa direto.
// Uma metadata.key de macro (sem chunk) não seleciona nada e é ignorada.
func (s *MetadataService) normalizeIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		q := strings.TrimSpace(id)
		if q == "" {
			continue
		}
		if strings.Contains(q, "/") || strings.Contains(q, "\\") {
			_, rid, _, err := s.Resolve(q)
			if err != nil {
				out = append(out, q)
				continue
			}
			if rid == "" {
				continue
			}
			out = append(out, rid)
			continue
		}
		out = append(out, q)
	}
	return out
}

// GetCollection monta o DTO em memória via builders (sem gravar arquivo).
// ids vazio = tudo (bulk: 20MB FFX / 45MB FFX2+lastmiss em events —
// uso excepcional; o padrão do frontend é ListEntries + GetEntry).
// ids aceita ids ou keys (metadata.key/caminho resolvidos para id).
func (s *MetadataService) GetCollection(kind string, version common.GameVersion, ids []string) (dto.Collection, error) {
	if err := ensureVersionReady(version); err != nil {
		return nil, err
	}
	ids = s.normalizeIDs(ids)
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case KindEvents:
		if err := ensureEventsLoaded(version); err != nil {
			return nil, err
		}
		return builders.BuildEventsDTO(version, ids)
	case KindObjects:
		return s.buildObjectsCollection(version, ids)
	case KindMacro:
		if version == common.GameVersionLastMiss {
			// Last Mission não tem dicionário próprio (macrodic do ffx2).
			return dto.Collection{}, nil
		}
		c, err := builders.BuildMacroDTO(version)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return c, nil
		}
		filter := make(map[string]bool, len(ids))
		for _, id := range ids {
			filter[id] = true
		}
		out := make(dto.Collection, len(ids))
		for _, k := range c.SortedKeys() {
			if filter[k] {
				out[k] = c[k]
			}
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("no matching macro chunks in DTO")
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

// ExportStrings monta o DTO em memória e escreve arquivos .strings,
// ao lado dos .json (mesmo diretório, mesmo basename).
// ids vazio = tudo; langs nil/vazio = todos os idiomas.
func (s *MetadataService) ExportStrings(kind string, version common.GameVersion, ids, langs []string) ([]string, error) {
	c, err := s.GetCollection(kind, version, ids)
	if err != nil {
		return nil, err
	}
	f := strfmt.NewStringsFormatter()
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case KindEvents:
		p, err := f.WriteEvents(c, version, langs)
		if err != nil {
			return nil, err
		}
		return []string{p}, nil
	case KindObjects:
		return f.WriteObjects(c, version, langs)
	case KindMacro:
		p, err := f.WriteMacro(c, version, langs)
		if err != nil {
			return nil, err
		}
		return []string{p}, nil
	default:
		return nil, fmt.Errorf("unknown kind: %s", kind)
	}
}

// buildObjectsCollection carrega cada arquivo por demanda (store + fallback
// de leitura) e monta o DTO. ids vazio = todos os layouts da versão.
func (s *MetadataService) buildObjectsCollection(version common.GameVersion, ids []string) (dto.Collection, error) {
	layouts, err := s.resolveObjectLayouts(version, ids)
	if err != nil {
		return nil, err
	}
	out := make(dto.Collection, len(layouts))
	for _, layout := range layouts {
		key := objectsfile.FileLayoutKey(version, layout.PatternPath())
		binFile, ok := objectsfile.ObjectFileDataStore.Get(key)
		if !ok {
			var lerr error
			binFile, lerr = objectsfile.LoadObjectFileFromStoreByLayout(layout)
			if lerr != nil {
				common.LogVerbose("skip objects %s: %v", key, lerr)
				continue
			}
		}
		if binFile == nil || binFile.GetObjects() == nil || binFile.GetObjects().IsEmpty() {
			continue
		}
		single, berr := builders.BuildObjectsDTO(binFile.GetObjects(), layout, key)
		if berr != nil {
			common.LogVerbose("skip objects %s: %v", key, berr)
			continue
		}
		for k, v := range single {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no objects with text data found")
	}
	return out, nil
}

// resolveObjectLayouts traduz ids (basenames) em layouts; vazio = todos.
func (s *MetadataService) resolveObjectLayouts(version common.GameVersion, ids []string) ([]objectsfile.FileLayout, error) {
	if len(ids) == 0 {
		var all []objectsfile.FileLayout
		for _, layout := range objectsfile.FileLayouts {
			if layout.Version == version {
				all = append(all, layout)
			}
		}
		if len(all) == 0 {
			return nil, fmt.Errorf("no object layouts for version %s", version)
		}
		sort.Slice(all, func(a, b int) bool { return all[a].FileName < all[b].FileName })
		return all, nil
	}
	var out []objectsfile.FileLayout
	var missing []string
	for _, id := range ids {
		key, ok := objectKeyForID(version, id)
		if !ok {
			missing = append(missing, id)
			continue
		}
		layout, ok := objectsfile.FileLayoutFor(version, key)
		if !ok {
			// key veio do próprio registro: fallback pelo mapa direto.
			layout = objectsfile.FileLayouts[key]
		}
		out = append(out, layout)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("none of the %d requested object(s) found: %v", len(ids), missing)
	}
	return out, nil
}

// ensureEventsLoaded garante o store em memória (carga bulk única;
// chamadas seguintes usam o datastore). Sem isso, o DTO viria vazio
// onde o Extract direto do arquivo funcionaria.
func ensureEventsLoaded(version common.GameVersion) error {
	if event.HasEvents(version) {
		return nil
	}
	if err := event.NewEventsBinaryFile(version).LoadFromBinary(); err != nil {
		return fmt.Errorf("failed to load events for version %s: %w", version, err)
	}
	return nil
}

