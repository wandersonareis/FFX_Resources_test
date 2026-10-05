package macrodic

import (
	"path/filepath"
	"sync"

	"ffxresources/backend/common"
)

// Vigia de frescura do dicionário de macros: carimbo físico (path/size/mtime)
// do macrodic.dcp de cada localização. Binário modificado no disco (tradução
// copiada/manual para mods/) república os macros no datastore na próxima
// leitura, sem reiniciar o app — mesmo vigia do lockit/objectsfile.
//
// O DTO de display NÃO passa por aqui (BuildMacroDTO lê os containers direto
// do disco); o vigia serve ao datastore de EXPANSÃO de tags publicado por
// PublishStrings na preparação da versão.

var (
	macrostampsMu sync.Mutex
	macrostamps   = map[common.GameVersion]map[string]common.FileStamp{}
)

// macroStampPath devolve o caminho relativo (à raiz de gamefiles) do
// macrodic.dcp da localização.
func macroStampPath(version common.GameVersion, loc string) string {
	return filepath.Join(common.GetLocalizationRootForVersion(version, loc), "menu", "macrodic.dcp")
}

// StampMacros registra o carimbo físico da localização (ou o remove quando o
// binário não existe mais).
func StampMacros(version common.GameVersion, loc string) {
	macrostampsMu.Lock()
	defer macrostampsMu.Unlock()
	m := macrostamps[version]
	if m == nil {
		m = map[string]common.FileStamp{}
		macrostamps[version] = m
	}
	if stamp, ok := common.StampFile(macroStampPath(version, loc)); ok {
		m[loc] = stamp
	} else {
		delete(m, loc)
	}
}

// EnsureMacrosFresh república os macros no datastore quando algum macrodic
// mudou no disco. Devolve true quando o store foi atualizado (o chamador
// invalida caches derivados). Localização sem carimbo: sem vigia.
func EnsureMacrosFresh(version common.GameVersion) bool {
	changed := false
	for _, loc := range DefaultFirstLocalizations() {
		macrostampsMu.Lock()
		stored, watched := macrostamps[version][loc]
		macrostampsMu.Unlock()
		if !watched {
			continue
		}
		current, exists := common.StampFile(macroStampPath(version, loc))
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

	common.LogVerbose("[macrodic] %s: binário mudou no disco — república os macros", version)
	for _, loc := range DefaultFirstLocalizations() {
		c := NewMacroDictionaryBinaryFile(loc, version)
		if err := c.LoadFromBinary(); err != nil {
			common.LogVerbose("[macrodic] pular %s: %v", loc, err)
			StampMacros(version, loc)
			continue
		}
		if err := c.PublishStrings(); err != nil {
			common.LogVerbose("[macrodic] %s: república falhou: %v", loc, err)
		}
		StampMacros(version, loc)
	}
	return true
}
