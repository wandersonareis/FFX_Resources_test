// Package hash centraliza o cálculo de hash xxHash64 (XXH64) das frases
// exportadas para JSON, via github.com/bouine-cache/xxhash/v3 (Sum64).
//
// O JSON carrega o hash como hex de 16 chars ("%016x"), não como uint64,
// para não perder precisão no frontend JS (>2^53).
package hash

import (
	"fmt"
	"sort"
	"strings"

	xxhash "github.com/bouine-cache/xxhash/v3"
)

// RefPrefix marca hashes e referências de dedup no JSON. O DTO em memória
// carrega hex puro; o `$` nasce e morre dentro de formatters/json.
const RefPrefix = "$"

// Sum64Hex calcula o XXH64 (seed zero) da frase e devolve hex "%016x".
// A frase hashed é o texto raw, incluindo tags de controle.
func Sum64Hex(s string) string {
	return fmt.Sprintf("%016x", xxhash.Sum64String(s))
}

// Texts calcula o hash por idioma, pulando textos vazios (sem entrada no map).
// Retorna nil quando nenhum idioma tem texto.
func Texts(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for lang, text := range in {
		if text == "" {
			continue
		}
		out[lang] = Sum64Hex(text)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SortedLangs devolve as chaves de idioma ordenadas, para saída determinística.
func SortedLangs(m map[string]string) []string {
	langs := make([]string, 0, len(m))
	for lang := range m {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// Prefix garante o prefixo de referência no hash (idempotente).
func Prefix(h string) string {
	if h == "" || strings.HasPrefix(h, RefPrefix) {
		return h
	}
	return RefPrefix + h
}

// Strip remove o prefixo de referência, informando se ele existia.
func Strip(s string) (string, bool) {
	if strings.HasPrefix(s, RefPrefix) {
		return strings.TrimPrefix(s, RefPrefix), true
	}
	return s, false
}
