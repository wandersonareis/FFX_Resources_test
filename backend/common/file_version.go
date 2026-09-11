package common

import (
	"path/filepath"
	"strings"
)

// FileVersionSuffix retorna o sufixo unificado da versão ativa:
// _ffx, _ffx2 ou _lastmiss.
func FileVersionSuffix() string {
	return CurrentGameVersion().Suffix()
}

// WithVersionSuffix insere o sufixo da versão ativa no nome do arquivo,
// antes da extensão (ex: events_all_localizations.json ->
// events_all_localizations_ffx2.json).
func WithVersionSuffix(fileName string) string {
	suffix := FileVersionSuffix()
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	return base + suffix + ext
}

// WithVersionSuffixFor é a variante explícita, sem ler o estado global.
func WithVersionSuffixFor(fileName string, gv GameVersion) string {
	suffix := gv.Normalize().Suffix()
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	return base + suffix + ext
}
