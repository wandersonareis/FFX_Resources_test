package services

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// TextSearchMinQueryRunes é o tamanho mínimo do termo para consultar o
// índice de conteúdo: abaixo disso quase tudo casa e o custo não compensa.
const TextSearchMinQueryRunes = 2

// TextSearchFreshCheckInterval limita a REVALIDAÇÃO do índice por versão
// entre consultas: o isFresh dá stat em todos os arquivos indexados, então
// consulta repetida reaproveita o último veredito. Escritas do app chamam
// clearTextSearchCache direto, então o intervalo só atrasa mudanças
// EXTERNAS (alguém mexeu nos gamefiles com o app aberto). Testes podem
// zerar para forçar a revalidação imediata.
var TextSearchFreshCheckInterval = 2 * time.Second

type indexedSourceStamps struct {
	data    common.FileStamp
	hasData bool
	mods    common.FileStamp
	hasMods bool
}

type textSearchIndexSlot struct {
	once  sync.Once
	index textSearchIndex
	err   error

	// Última verificação de frescor, com o resultado cacheado até o
	// próximo intervalo (protegido por próprio mutex: o slot é lido por
	// várias goroutines de IPC sem ordem definida).
	freshMu        sync.Mutex
	freshCheckedAt time.Time
	fresh          bool
}

// indexFresh aplica o throttle sobre o isFresh (stat de cada arquivo).
func (slot *textSearchIndexSlot) indexFresh() bool {
	slot.freshMu.Lock()
	defer slot.freshMu.Unlock()
	if time.Since(slot.freshCheckedAt) < TextSearchFreshCheckInterval {
		return slot.fresh
	}
	slot.fresh = slot.index.isFresh()
	slot.freshCheckedAt = time.Now()
	return slot.fresh
}

var textSearchCache = struct {
	sync.Mutex
	byVersion map[common.GameVersion]*textSearchIndexSlot
}{byVersion: make(map[common.GameVersion]*textSearchIndexSlot)}

// SearchText procura em data/ (Original) e mods/ (Traduzido), somente no
// idioma us. O índice de cada versão é criado na primeira consulta e mantido
// até que seus arquivos mudem ou os caches da árvore sejam invalidados.
func (s *MetadataService) SearchText(version common.GameVersion, query string) (TextSearchResponse, error) {
	if len([]rune(strings.TrimSpace(query))) < TextSearchMinQueryRunes {
		return TextSearchResponse{}, nil
	}

	for attempt := 0; attempt < 2; attempt++ {
		slot, err := s.textSearchSlot(version)
		if err != nil {
			return TextSearchResponse{}, err
		}
		if !textSearchSlotIsCurrent(version, slot) {
			continue
		}
		if slot.indexFresh() {
			return slot.index.search(query), nil
		}
		textSearchDropSlot(version, slot)
	}
	return TextSearchResponse{}, fmt.Errorf("gamefiles mudaram durante a indexação da busca; tente novamente")
}

// WarmTextSearch pré-constrói o índice da versão sem consulta (o slot é o
// mesmo do SearchText): a primeira busca da aba não paga a construção.
func (s *MetadataService) WarmTextSearch(version common.GameVersion) {
	_, _ = s.textSearchSlot(version)
}

// textSearchSlot devolve o slot da versão com o índice pronto. Um build que
// FALHOU não fica em cache para sempre: o slot é removido para a próxima
// consulta tentar de novo (a chamada atual ainda recebe o erro).
func (s *MetadataService) textSearchSlot(version common.GameVersion) (*textSearchIndexSlot, error) {
	slot := textSearchSlotFor(version)
	var buildErr error
	slot.once.Do(func() {
		slot.index, buildErr = s.buildTextSearchIndex(version)
		slot.err = buildErr
	})
	if slot.err != nil {
		textSearchDropSlot(version, slot)
		return nil, slot.err
	}
	return slot, nil
}

func textSearchSlotFor(version common.GameVersion) *textSearchIndexSlot {
	textSearchCache.Lock()
	defer textSearchCache.Unlock()
	slot := textSearchCache.byVersion[version]
	if slot == nil {
		slot = &textSearchIndexSlot{}
		textSearchCache.byVersion[version] = slot
	}
	return slot
}

func textSearchSlotIsCurrent(version common.GameVersion, slot *textSearchIndexSlot) bool {
	textSearchCache.Lock()
	defer textSearchCache.Unlock()
	return textSearchCache.byVersion[version] == slot
}

// textSearchDropSlot remove o slot SÓ se ainda for o corrente — descarte
// idempotente para o retry de build/consulta falhada.
func textSearchDropSlot(version common.GameVersion, slot *textSearchIndexSlot) {
	textSearchCache.Lock()
	if textSearchCache.byVersion[version] == slot {
		delete(textSearchCache.byVersion, version)
	}
	textSearchCache.Unlock()
}

func clearTextSearchCache() {
	textSearchCache.Lock()
	textSearchCache.byVersion = make(map[common.GameVersion]*textSearchIndexSlot)
	textSearchCache.Unlock()
}

func searchKindsFor(version common.GameVersion) []string {
	switch version {
	case common.GameVersionFFX:
		return []string{KindEvents, KindObjects, KindMacro, KindLockit, KindHelp, KindBattleText, KindCloud, KindMenuMain}
	case common.GameVersionFFX2:
		return []string{KindEvents, KindObjects, KindMacro, KindLockit, KindBattleText, KindCloud, KindTutorial}
	case common.GameVersionLastMiss:
		return []string{KindEvents, KindObjects}
	case common.GameVersionEternalCalm:
		return []string{KindEvents}
	default:
		return nil
	}
}

func (s *MetadataService) buildTextSearchIndex(version common.GameVersion) (textSearchIndex, error) {
	index := textSearchIndex{stamps: make(map[string]indexedSourceStamps)}
	for _, kind := range searchKindsFor(version) {
		if kind == KindMacro {
			s.indexMacroSearch(version, &index)
			continue
		}

		entries, err := s.ListEntries(kind, version)
		if err != nil {
			common.LogVerbose("busca de texto: sem índice de %s/%s: %v", version, kind, err)
			continue
		}
		for _, summary := range entries {
			var dataEntry, modsEntry *dto.FileEntry
			for _, source := range []common.FileSource{common.SourceData, common.SourceMods} {
				entry, exists, loadErr := s.loadOriginalFrom(kind, summary.ID, version, source)
				if loadErr != nil || !exists {
					continue
				}
				copy := entry
				if source == common.SourceData {
					dataEntry = &copy
				} else {
					modsEntry = &copy
				}
			}
			index.addFile(version, kind, summary.ID, summary.Key, dataEntry, modsEntry)
		}
	}
	return index, nil
}

func (s *MetadataService) indexMacroSearch(version common.GameVersion, index *textSearchIndex) {
	data, err := builders.BuildMacroDTOFromSource(version, common.SourceData)
	if err != nil {
		common.LogVerbose("busca de texto: sem índice de macro/%s em data/: %v", version, err)
		return
	}
	mods, _ := builders.BuildMacroDTOFromSource(version, common.SourceMods)
	for _, id := range data.SortedKeys() {
		dataEntry := data[id]
		if !macroChunkHasText(dataEntry) {
			continue
		}
		var modsEntry *dto.FileEntry
		if entry, ok := mods[id]; ok {
			copy := entry
			modsEntry = &copy
		}
		copy := dataEntry
		index.addFile(version, KindMacro, id, dataEntry.Metadata.Key, &copy, modsEntry)
	}
}

func (index *textSearchIndex) addFile(
	version common.GameVersion,
	kind, id, key string,
	data, mods *dto.FileEntry,
) {
	index.files = append(index.files, indexedFileFromEntries(kind, id, data, mods))
	rel := textSearchPathForKey(version, key)
	if rel == "" {
		return
	}
	dataStamp, hasData := common.StampFileFrom(rel, common.SourceData)
	modsStamp, hasMods := common.StampFileFrom(rel, common.SourceMods)
	index.stamps[rel] = indexedSourceStamps{
		data: dataStamp, hasData: hasData,
		mods: modsStamp, hasMods: hasMods,
	}
}

func textSearchPathForKey(version common.GameVersion, key string) string {
	parsed, ok := dto.ParseKey(key)
	if !ok {
		return ""
	}
	return filepath.Join(
		common.GetLocalizationRootForVersion(version, common.DefaultLocalization),
		filepath.FromSlash(parsed.LocalizationPattern),
	)
}

func (index textSearchIndex) isFresh() bool {
	for rel, stored := range index.stamps {
		data, hasData := common.StampFileFrom(rel, common.SourceData)
		mods, hasMods := common.StampFileFrom(rel, common.SourceMods)
		if stored.hasData != hasData || (hasData && stored.data != data) ||
			stored.hasMods != hasMods || (hasMods && stored.mods != mods) {
			return false
		}
	}
	return true
}
