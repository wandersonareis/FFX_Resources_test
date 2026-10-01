package helpfile

import (
	"path/filepath"
	"sync"

	"ffxresources/backend/common"
)

// Vigia de frescura dos painéis de ajuda: carimbo físico (path/size/mtime)
// do .sps2 de cada painel na localização padrão. Binário modificado no disco
// (tradução copiada/manual para mods/) recarrega o store na próxima leitura,
// sem reiniciar o app — mesmo vigia do lockit/objectsfile/events.

var (
	helpStampsMu sync.Mutex
	helpStamps   = map[common.GameVersion]map[string]common.FileStamp{}
)

// helpStampRelPath devolve o caminho relativo (à raiz de gamefiles) do .sps2
// do painel na localização padrão.
func helpStampRelPath(version common.GameVersion, name string) string {
	return filepath.Join(
		common.GetLocalizationRootForVersion(version, common.DefaultLocalization),
		HelpRelPath(name),
	)
}

// stampPanel registra o carimbo físico do painel (ou o remove quando o
// binário não existe mais).
func stampPanel(version common.GameVersion, name string) {
	helpStampsMu.Lock()
	defer helpStampsMu.Unlock()
	m := helpStamps[version]
	if m == nil {
		m = map[string]common.FileStamp{}
		helpStamps[version] = m
	}
	if stamp, ok := common.StampFile(helpStampRelPath(version, name)); ok {
		m[name] = stamp
	} else {
		delete(m, name)
	}
}

// StampAllPanels carimba todos os painéis da versão (após a carga bulk).
func StampAllPanels(version common.GameVersion, names []string) {
	for _, name := range names {
		stampPanel(version, name)
	}
}

// EnsureHelpFresh garante o store fresco: carga quando vazio (a carga
// inicial não é mudança externa) e recarga completa quando algum .sps2
// mudou no disco. Devolve true quando o store foi RECARGADO (o chamador
// invalida caches derivados, como o view dedupado).
func EnsureHelpFresh(version common.GameVersion) bool {
	helpStoreMu.Lock()
	loaded := len(helpStore[version]) > 0
	helpStoreMu.Unlock()

	if !loaded {
		_ = EnsureHelpLoaded(version)
		return false
	}

	changed := false
	for _, entry := range HelpEntries {
		if GetHelp(version, entry.Name) == nil {
			continue // painel sem árvore nesta versão
		}
		helpStampsMu.Lock()
		stored, watched := helpStamps[version][entry.Name]
		helpStampsMu.Unlock()
		if !watched {
			continue // painel sem carimbo (Register direto): confiada
		}
		current, exists := common.StampFile(helpStampRelPath(version, entry.Name))
		if !exists && stored == (common.FileStamp{}) {
			continue
		}
		if exists && current == stored {
			continue
		}
		changed = true
		break
	}
	if !changed {
		return false
	}

	common.LogVerbose("[helpfile] %s: binário mudou no disco — recarregando painéis", version)
	ClearHelp(version)
	if err := EnsureHelpLoaded(version); err != nil {
		common.LogWarning("[helpfile] recarga falhou (store mantém o anterior): %v", err)
		return false
	}
	return true
}
