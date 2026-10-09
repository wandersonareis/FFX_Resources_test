// Pré-carga de versões em background: após a aba visualizada concluir a
// carga, as demais versões são aquecidas UMA A UMA (I/O de disco serial,
// sem disputa) para que a troca de aba caia no caminho rápido.
//
// Regras:
//   - idempotente: versão já pré-cargada (nesta árvore) é pulada;
//   - invalidação: GameFilesLocation novo zera o Set e ABORTA a fila
//     (a geração da árvore muda — o aquecimento serviu para a árvore
//     anterior);
//   - sem barra de progresso: é background, o tradutor nem vê; o registro
//     fica no arquivo de diagnóstico.
package services

import (
	"sync"
	"sync/atomic"

	"ffxresources/backend/common"
	"ffxresources/backend/loggingService"
)

var (
	preloadMu sync.Mutex
	// preloadedVersions: versões já aquecidas na árvore ATUAL.
	preloadedVersions = map[string]bool{}
	// treeGeneration sobe a cada InvalidateViewCaches: o worker de pré-carga
	// compara a geração capturada no início para abortar a fila.
	treeGeneration atomic.Int64
)

func markPreloaded(version common.GameVersion) (first bool) {
	key := version.String()
	preloadMu.Lock()
	defer preloadMu.Unlock()
	if preloadedVersions[key] {
		return false
	}
	preloadedVersions[key] = true
	return true
}

// PreloadVersions aquece as versões indicadas em uma goroutine worker,
// sequencialmente. Chamada duplicada ignora as versões já aquecidas.
func (s *MetadataService) PreloadVersions(versions []common.GameVersion) {
	go func() {
		gen := treeGeneration.Load()
		for _, version := range versions {
			if treeGeneration.Load() != gen {
				loggingService.Info("pré-carga abortada: a árvore de gamefiles mudou no meio")
				return
			}
			if err := s.preloadVersion(version, gen); err != nil {
				loggingService.Warn("pré-carga %s falhou: %v", version, err)
			}
		}
	}()
}

// preloadVersion aquece UMA versão: store de eventos + inventário de cada
// kind + a ordem global dos ponteiros dos kinds dedupados (a passagem
// pesada de normalize — lê os originais dos arquivos traduzidos). gen é a
// geração da árvore capturada no início da fila: se mudar no meio, aborta.
func (s *MetadataService) preloadVersion(version common.GameVersion, gen int64) error {
	if !markPreloaded(version) {
		return nil
	}
	if err := ensureVersionReady(version); err != nil {
		return err
	}
	if err := ensureEventsLoaded(version); err != nil {
		return err
	}
	// KindImages fica de fora: não tem store nem coleção para aquecer — é
	// um walk do disco por ListEntries, e o walk não é memoizado.
	for _, kind := range []string{KindEvents, KindObjects, KindMacro, KindLockit, KindHelp} {
		if treeGeneration.Load() != gen {
			// A árvore mudou no meio da fila: os caches aquecidos serviam
			// à árvore anterior — aborta sem tocar no estado novo.
			return nil
		}
		if _, err := s.ListEntries(kind, version); err != nil {
			// kind sem árvore na versão (help/macro p/ lastmiss): segue.
			loggingService.Verbose("pré-carga %s/%s sem entries: %v", kind, version, err)
		}
	}
	for kind := range dedupViewKinds {
		if _, err := s.rawViewOf(kind, version); err != nil {
			loggingService.Verbose("pré-carga %s/%s sem view dedupada: %v", kind, version, err)
		}
	}
	loggingService.Info("pré-carga %s concluída — trocar de aba carrega do cache", version)
	return nil
}
