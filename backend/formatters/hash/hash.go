// Package hash centraliza o cálculo de hash xxHash64 (XXH64) das frases
// exportadas para JSON/.strings/CSV/XLSX, via github.com/bouine-cache/xxhash/v3
// (Sum64).
//
// O hash carrega como hex de 16 chars ("%016x"), não como uint64,
// para não perder precisão no frontend JS (>2^53).
//
// O pacote é o DONO do formato do hash em artefato: o literal do prefixo
// existe só aqui (RefPrefix) e os formatters só fazem/refazem/medem o
// prefixo por Prefix/Strip/IsRef — nada de manipular "$" solto no código.
// Se o formato um dia mudar (prefixo, tamanho, codificação), muda aqui.
// A régua de dedup (elegibilidade de colapso pelo miolo visível) vive em
// dedup.go, no mesmo pacote.
package hash

import (
	"fmt"
	"strings"

	xxhash "github.com/bouine-cache/xxhash/v3"
)

// RefPrefix marca hashes e referências de dedup. O DTO em memória carrega hex
// puro com uma única exceção: a entrega de help ao frontend — o builders
// DedupHelpDTO escreve refs "$hash" em text[us] (ocultas na tabela) e o apply
// as resolve de volta. No formato artefato (JSON/.strings/CSV/XLSX) o `$`
// nasce nos formatters, sempre via Prefix/Strip/IsRef.
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

// Prefix garante o prefixo de referência no hash (idempotente): é o hash no
// formato de artefato ("$"+hex) e também a ref de dedup que ocupa o lugar de
// um texto repetido.
func Prefix(h string) string {
	if h == "" || strings.HasPrefix(h, RefPrefix) {
		return h
	}
	return RefPrefix + h
}

// IsRef informa se o texto carrega o prefixo de referência de hash — a
// pergunta "isto é uma ref/hash de artefato?" sem manipular o literal.
func IsRef(s string) bool {
	return strings.HasPrefix(s, RefPrefix)
}

// Strip remove o prefixo de referência, informando se ele existia.
func Strip(s string) (string, bool) {
	if strings.HasPrefix(s, RefPrefix) {
		return strings.TrimPrefix(s, RefPrefix), true
	}
	return s, false
}
