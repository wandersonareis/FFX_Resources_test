package services

import (
	"fmt"
	"path/filepath"
	"sync"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/fileFormats/lockit"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/hash"
)

// ORIGINAL: a coluna "Original" da tabela vem SEMPRE da árvore data/
// (fonte da verdade, imutável); a coluna "Traduzido" (Text) continua
// mods-first (último save). Este arquivo concentra a leitura e a
// mesclagem do original em GetEntry.
//
// Regras:
//   - presença (arquivo/linha/idioma) é definida por data/;
//   - o que existe só em mods/ é IGNORADO (LogWarning) — não há original
//     contra o qual revisar;
//   - falha de leitura de data/ nunca quebra o view: loga e entrega a
//     entrada sem `original` (o frontend mostra "—");
//   - o ponteiro (Hash) da row casada passa a ser o hash do ORIGINAL de
//     data/ (imutável, uma vez na leitura pristine): o Traduzido (Text)
//     reusa o MESMO ponteiro — hash(Text) == hash(Original) ⇔ célula
//     ainda não traduzida; divergência de estrutura degrada por row
//     (sem reescrita nessa row).

// originalSource é a fonte de TODA leitura do original: data/ puro.
const originalSource = common.SourceData

type originalValue struct {
	// exists: o binário da entrada EXISTE em data/ (presença).
	exists bool
	// entry: conteúdo pristine de data/ (vazio quando !exists).
	entry dto.FileEntry
}

var (
	originalMu    sync.Mutex
	originalCache = map[string]originalValue{}
	originalMacro = map[common.GameVersion]dto.Collection{}
)

func originalCacheKey(kind string, version common.GameVersion, id string) string {
	return kind + "|" + version.String() + "|" + id
}

// originalRelPath devolve o caminho (na raiz usada pelo respectivo leitor)
// do binário pristine da entrada. ok=false significa "a entrada não tem
// arquivo associado" (id desconhecido) — quem trata é o chamador.
func originalRelPath(kind, id string, version common.GameVersion) (string, bool) {
	loc := common.DefaultLocalization
	switch kind {
	case KindEvents:
		rel, err := event.EventRelPath(id)
		if err != nil {
			return "", false
		}
		return filepath.Join(common.GetLocalizationRootForVersion(version, loc), rel), true
	case KindObjects:
		key, ok := objectKeyForID(version, id)
		if !ok {
			return "", false
		}
		layout, ok := objectsfile.FileLayoutFor(version, key)
		if !ok {
			layout = objectsfile.FileLayouts[key]
		}
		return filepath.Join(
			common.GetLocalizationRootForVersion(version, loc),
			layout.PatternPath(),
		), true
	case KindMacro:
		if version == common.GameVersionLastMiss {
			return "", false
		}
		return filepath.Join(
			common.GetLocalizationRootForVersion(version, loc),
			macrodic.MacroRelPath(),
		), true
	case KindLockit:
		for _, l := range lockit.LayoutsForVersion(version) {
			if l.ID() == id {
				return l.RelPath(loc), true
			}
		}
		return "", false
	case KindHelp:
		if version != common.GameVersionFFX || !helpfile.IsHelpEntry(id) {
			return "", false
		}
		return helpfile.HelpPathForVersion(version, loc, id), true
	}
	return "", false
}

// originalTrees informa em qual árvore o binário da ENTRADA está.
//
//	inData  → data/ tem o arquivo: há original (fonte da verdade);
//	inMods  → só mods/ tem: regra 4, a entrada é ignorada;
//	nenhum  → o arquivo não está em árvore nenhuma (store sintético ou
//	          entrada sem arquivo associado): entrega sem `original`,
//	          sem quebrar o view.
func originalTrees(kind, id string, version common.GameVersion) (inData, inMods bool) {
	// events: a ACHAGEM do evento já é feita em data/ (o store só recebe o
	// que veio da árvore original), então o registro no store É a presença —
	// stat adicional aqui só tornaria o view frágil a atraso de disco.
	if kind == KindEvents {
		return event.GetEvent(version, id) != nil, false
	}

	rel, ok := originalRelPath(kind, id, version)
	if !ok {
		return false, false
	}
	if acc, err := common.NewFileAccessorFrom(rel, common.SourceData); err == nil && acc.Exists {
		return true, false
	}
	if acc, err := common.NewFileAccessorFrom(rel, common.SourceMods); err == nil && acc.Exists {
		return false, true
	}
	return false, false
}

// hasModsFile informa se o binário da entrada existe em mods/. É o gatilho
// do custo de leitura do original na construção da ordem global de ponteiros:
// arquivo sem mods não está traduzido, e o ponteiro de mods JÁ É o ponteiro
// do original (mesmo conteúdo ⇒ mesmo XXH64).
func hasModsFile(kind, id string, version common.GameVersion) bool {
	rel, ok := originalRelPath(kind, id, version)
	if !ok {
		return false
	}
	acc, err := common.NewFileAccessorFrom(rel, common.SourceMods)
	return err == nil && acc.Exists
}

// originalExists é o teste de presença usado na árvore lateral: só data/
// define quem aparece (regra 4 esconde o que só existe em mods/).
func originalExists(kind, id string, version common.GameVersion) bool {
	inData, _ := originalTrees(kind, id, version)
	return inData
}

// originalFor devolve a entrada pristine de data/ para (kind, version, id),
// com cache. Inexistência é exists=false, sem erro; erro de leitura/parse
// NÃO é memoizado (pode ser transitória) e o chamador degrada.
func (s *MetadataService) originalFor(kind, id string, version common.GameVersion) (dto.FileEntry, bool, error) {
	if kind == KindMacro {
		return s.originalMacroEntry(id, version)
	}

	key := originalCacheKey(kind, version, id)
	originalMu.Lock()
	if v, ok := originalCache[key]; ok {
		originalMu.Unlock()
		return v.entry, v.exists, nil
	}
	originalMu.Unlock()

	entry, exists, err := s.loadOriginal(kind, id, version)
	if err == nil {
		originalMu.Lock()
		originalCache[key] = originalValue{exists: exists, entry: entry}
		originalMu.Unlock()
	}
	return entry, exists, err
}

// originalMacroEntry lê o macrodic de data/ uma vez por versão (o binário é
// único: todos os chunks saem do mesmo arquivo) e recorta a entrada pedida.
func (s *MetadataService) originalMacroEntry(id string, version common.GameVersion) (dto.FileEntry, bool, error) {
	originalMu.Lock()
	if c, ok := originalMacro[version]; ok {
		originalMu.Unlock()
		entry, found := c[id]
		return entry, found, nil
	}
	originalMu.Unlock()

	c, err := builders.BuildMacroDTOFromSource(version, originalSource)
	if err != nil {
		// Falha de leitura/interpretação: não memoiza (pode ser transitória).
		return dto.FileEntry{}, false, err
	}
	originalMu.Lock()
	originalMacro[version] = c
	originalMu.Unlock()

	entry, found := c[id]
	return entry, found, nil
}

// loadOriginal monta a entrada pristine da árvore data/ para o kind dado,
// sem registrar nada em store nenhum (os stores continuam refletindo a
// visão traduzida/normal).
func (s *MetadataService) loadOriginal(kind, id string, version common.GameVersion) (dto.FileEntry, bool, error) {
	switch kind {
	case KindEvents:
		strs, err := event.ReadLocalizedEventStringsFrom(id, version, originalSource)
		if err != nil {
			return dto.FileEntry{}, false, err
		}
		if len(strs) == 0 {
			return dto.FileEntry{}, false, nil
		}
		entry, ok := builders.BuildEventEntryDTOFrom(id, version, strs)
		if !ok {
			return dto.FileEntry{}, false, nil
		}
		return entry, true, nil

	case KindObjects:
		layouts, err := s.resolveObjectLayouts(version, []string{id})
		if err != nil {
			return dto.FileEntry{}, false, fmt.Errorf("objects %s: %w", id, err)
		}
		if len(layouts) == 0 {
			return dto.FileEntry{}, false, fmt.Errorf("objects %s: layout não encontrado", id)
		}
		layout := layouts[0]
		key := objectsfile.FileLayoutKey(version, layout.PatternPath())
		binFile, err := objectsfile.LoadObjectFileFrom(layout, originalSource)
		if err != nil {
			return dto.FileEntry{}, false, fmt.Errorf("objects %s: %w", id, err)
		}
		if binFile == nil || binFile.GetObjects() == nil || binFile.GetObjects().IsEmpty() {
			return dto.FileEntry{}, false, nil
		}
		c, err := builders.BuildObjectsDTO(binFile.GetObjects(), layout, key)
		if err != nil {
			return dto.FileEntry{}, false, fmt.Errorf("objects %s: %w", id, err)
		}
		entry, ok := c[id]
		if !ok {
			return dto.FileEntry{}, false, nil
		}
		return entry, true, nil

	case KindLockit:
		for _, l := range lockit.LayoutsForVersion(version) {
			if l.ID() != id {
				continue
			}
			f, err := lockit.LoadFrom(l, originalSource)
			if err != nil {
				return dto.FileEntry{}, false, fmt.Errorf("lockit %s: %w", id, err)
			}
			c, err := builders.BuildLockitDTOFrom([]*lockit.LockitFile{f})
			if err != nil {
				return dto.FileEntry{}, false, fmt.Errorf("lockit %s: %w", id, err)
			}
			entry, ok := c[id]
			if !ok {
				return dto.FileEntry{}, false, nil
			}
			return entry, true, nil
		}
		return dto.FileEntry{}, false, nil

	case KindHelp:
		panel := helpfile.ReadHelpPanelFrom(version, id, originalSource)
		if panel == nil {
			return dto.FileEntry{}, false, nil
		}
		entry, ok := builders.BuildHelpEntryDTOFrom(id, version, panel)
		if !ok {
			return dto.FileEntry{}, false, nil
		}
		return entry, true, nil
	}
	return dto.FileEntry{}, false, fmt.Errorf("unknown kind: %s", kind)
}

// withOriginal devolve uma CÓPIA de current com `Original` preenchido e o
// PONTEIRO de dupe (Hash) reescrito a partir de data/, casando as rows por
// (Index, Name) — e devolve o relatório de divergência de estrutura.
//
// O hash de exibição é o hash do ORIGINAL (imutável, criado uma vez na
// leitura pristine): coluna Traduzido (Text) reusa o MESMO ponteiro, então
// hash(Text) == hash(Original) ⇔ célula ainda não traduzida — a comparação
// dos dois é o flag "pendente de revisão" e decide o dupe.
//
// Rows sem contraparte ficam sem `original` (o frontend exibe "—") e com o
// hash de mods; o texto traduzido nunca vaza para a coluna Original. A
// emissão do aviso é do CHAMADOR (logDivergence), em ordem — os workers
// paralelos apenas coletam.
func withOriginal(ref entryRef, current, orig dto.FileEntry) (dto.FileEntry, divergeDiag) {
	if len(current.Rows) == 0 || len(orig.Rows) == 0 {
		return current, divergeDiag{ref: ref, dataRows: len(orig.Rows), modsRows: len(current.Rows)}
	}

	type rowKey struct {
		index int
		name  string
	}
	byKey := make(map[rowKey]map[string]string, len(orig.Rows))
	seenOrig := make(map[rowKey]bool, len(orig.Rows))
	for _, r := range orig.Rows {
		byKey[rowKey{r.Index, r.Name}] = r.Text
	}

	rows := make([]dto.TextRow, len(current.Rows))
	copy(rows, current.Rows)

	matched := 0
	var onlyInMods []DiagRow
	for i := range rows {
		k := rowKey{rows[i].Index, rows[i].Name}
		src, ok := byKey[k]
		if !ok || len(src) == 0 {
			onlyInMods = append(onlyInMods, diagRowOf(k.index, k.name, rows[i].Text[common.DefaultLocalization]))
			continue
		}
		cp := make(map[string]string, len(src))
		for kk, v := range src {
			cp[kk] = v
		}
		rows[i].Original = cp
		seenOrig[k] = true
		matched++
	}

	var onlyInData []DiagRow
	for _, r := range orig.Rows {
		k := rowKey{r.Index, r.Name}
		if seenOrig[k] {
			continue
		}
		onlyInData = append(onlyInData, diagRowOf(k.index, k.name, r.Text[common.DefaultLocalization]))
	}

	diag := divergeDiag{
		ref:        ref,
		dataRows:   len(orig.Rows),
		modsRows:   len(rows),
		onlyInMods: onlyInMods,
		onlyInData: onlyInData,
	}

	// Ponteiro estrutural por row casada: hash do original por idioma.
	// Idioma sem texto original mantém o hash atual (nada para agrupar por
	// ali). Idempotente: célula não traduzida já tem os hashes coincidentes.
	for i := range rows {
		if rows[i].Original == nil {
			continue
		}
		for lang, o := range rows[i].Original {
			if o == "" {
				continue
			}
			if rows[i].Hash == nil {
				rows[i].Hash = map[string]string{}
			}
			rows[i].Hash[lang] = hash.Sum64Hex(o)
		}
	}
	current.Rows = rows
	return current, diag
}

// clearOriginalCache descarta o original em memória (a árvore data/ mudou —
// ou os caches de view estão sendo invalidados por higiene).
func clearOriginalCache() {
	originalMu.Lock()
	originalCache = map[string]originalValue{}
	originalMacro = map[common.GameVersion]dto.Collection{}
	originalMu.Unlock()
}

// InvalidateViewCaches descarta tudo que foi DERIVADO da árvore de
// gamefiles (DTOs deduplicados, original pristine e os stores de formato).
// Chamado ao trocar GameFilesLocation: os binários carregados pertencem à
// árvore anterior e sem isso a view serviria texto de outro diretório.
func InvalidateViewCaches() {
	treeGeneration.Add(1)
	preloadMu.Lock()
	preloadedVersions = map[string]bool{}
	preloadMu.Unlock()
	clearDedupViewCache()
	event.ClearAllEvents()
	helpfile.ClearAllHelp()
	objectsfile.ObjectFileDataStore.Clear()
	lockit.DataStore.Clear()
	datastore.Instance.ClearAllMacros()
	// A preparação por versão (charsets + macros) leu a árvore anterior:
	// zera para que a próxima chamada recarregue do novo diretório.
	readyMu.Lock()
	readyVersions = map[common.GameVersion]error{}
	readyMu.Unlock()
}
