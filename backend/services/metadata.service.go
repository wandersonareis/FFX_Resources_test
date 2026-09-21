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
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/fileFormats/objectsfile"
	strfmt "ffxresources/backend/formatters/strings"
	"ffxresources/backend/interactions"
	"ffxresources/backend/spira"
)

// Kinds lógicos de texto servidos ao frontend.
const (
	KindEvents  = "events"
	KindObjects = "objects"
	KindMacro   = "macro"
)

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

// ResolveEntryLocation deriva os caminhos em disco de uma entrada
// (kind/id/version) a partir da metadata.key canônica, sem varrer a árvore de
// diretórios. Serve Ver/Extrair/Importar no frontend.
func (s *MetadataService) ResolveEntryLocation(kind, id string, version common.GameVersion) (dto.EntryLocation, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	key, err := entryKeyForID(kind, id, version)
	if err != nil {
		return dto.EntryLocation{}, err
	}
	p, ok := dto.ParseKey(key)
	if !ok {
		return dto.EntryLocation{}, fmt.Errorf("invalid key: %s", key)
	}

	sourcePath := filepath.Join(
		common.GameFilesRoot,
		filepath.FromSlash(common.PackRootForVersion(version, common.DefaultLocalization)),
		filepath.FromSlash(p.LocalizationPattern),
	)

	loc := dto.EntryLocation{Kind: kind, ID: id, Key: key, SourcePath: sourcePath}
	if !common.IsFileExists(sourcePath) {
		return loc, nil
	}

	formatter := interactions.NewInteractionService().TextFormatter()
	node, nerr := spira.BuildNode(sourcePath, formatter)
	if nerr != nil || node == nil || node.Data == nil {
		return loc, nil
	}
	if node.Data.Extract != nil {
		loc.ExtractTarget = node.Data.Extract.GetTargetFile()
		loc.IsExtracted = common.IsFileExists(loc.ExtractTarget)
	}
	if node.Data.Translate != nil {
		loc.TranslateTarget = node.Data.Translate.GetTargetFile()
		loc.IsTranslated = common.IsFileExists(loc.TranslateTarget)
	}
	return loc, nil
}

// entryKeyForID devolve a metadata.key canônica de uma entrada.
func entryKeyForID(kind, id string, version common.GameVersion) (string, error) {
	switch kind {
	case KindEvents:
		return dto.NewEventMetadata(id, version).Key, nil
	case KindObjects:
		if key, ok := objectKeyForID(version, id); ok {
			return key, nil
		}
		return "", fmt.Errorf("unknown object id: %s", id)
	case KindMacro:
		chunk, ok := dto.ChunkIndexFromID(id)
		if !ok {
			return "", fmt.Errorf("unknown macro chunk id: %s", id)
		}
		return dto.NewMacroMetadata(version, chunk).Key, nil
	default:
		return "", fmt.Errorf("unknown kind: %s", kind)
	}
}

// ---- texto em memória (DTO estruturado, sem disco) --------------------------

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
		ids := event.GetAllEventIDs(version)
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
		c, err := builders.BuildMacroDTO(version)
		if err != nil {
			return nil, err
		}
		out := make([]EntrySummary, 0, len(c))
		for _, k := range c.SortedKeys() {
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

// EnsureMacroContainers expõe a leitura dos containers (uso avançado).
func (s *MetadataService) macroContainers(version common.GameVersion) (map[string]*macrodic.MacroDictionaryBinaryFile, error) {
	return macrodic.ReadMacroDictionaryContainers(version)
}
