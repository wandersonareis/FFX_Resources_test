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

// StripVersionSuffix remove o sufixo de versão do nome, se presente.
func StripVersionSuffix(fileName string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	for _, s := range []string{"_lastmiss", "_ffx2", "_ffx"} {
		if strings.HasSuffix(base, s) {
			base = strings.TrimSuffix(base, s)
			break
		}
	}
	return base + ext
}

// WithVersionSuffix insere o sufixo da versão ativa no nome do arquivo,
// antes da extensão (ex: events_all_localizations.json ->
// events_all_localizations_ffx2.json). Idempotente: se o nome já possui
// um sufixo de versão, ele é normalizado para a versão ativa.
func WithVersionSuffix(fileName string) string {
	base := StripVersionSuffix(fileName)
	suffix := FileVersionSuffix()
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + suffix + ext
}

// WithVersionSuffixFor é a variante explícita, sem ler o estado global.
// Idempotente: normaliza qualquer sufixo existente para gv.
func WithVersionSuffixFor(fileName string, gv GameVersion) string {
	base := StripVersionSuffix(fileName)
	suffix := gv.Suffix()
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + suffix + ext
}
