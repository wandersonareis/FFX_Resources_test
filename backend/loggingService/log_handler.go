package loggingService

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog"
)

// Métodos de log do projeto: console colorido pelo tipo + arquivo em JSON.
// O arquivo é persistência — cabe mais detalhe; console e arquivo não
// precisam ser idênticos.

// Debug registra em nível debug (só com verbose/LOG_LEVEL no nível debug).
func Debug(format string, args ...any) {
	if !debugEnabled() {
		return
	}
	Init()
	emit(zerolog.DebugLevel, format, args...)
}

// Info registra em nível info (console verde, arquivo JSON).
func Info(format string, args ...any) {
	Init()
	emit(zerolog.InfoLevel, format, args...)
}

// Warn registra um aviso recuperável (entrada ignorada, divergência de
// árvore, degradação controlada) — visível sempre.
func Warn(format string, args ...any) {
	Init()
	emit(zerolog.WarnLevel, format, args...)
}

// Error registra um erro recuperável.
func Error(format string, args ...any) {
	Init()
	emit(zerolog.ErrorLevel, format, args...)
}

// Verbose respeita o VERBOSE_MODE do projeto: com verbose desligado a
// mensagem não é emitida em lugar nenhum.
func Verbose(format string, args ...any) {
	if !verboseEnabled() {
		return
	}
	Debug(format, args...)
}

func emit(level zerolog.Level, format string, args ...any) {
	ev := log.WithLevel(level)
	if len(args) > 0 {
		ev.Msgf(format, args...)
		return
	}
	ev.Msg(format)
}

// DiagInfo registra um evento de inventário: console recebe a linha curta
// colorida; o arquivo de diagnóstico recebe o JSON completo com os detalhes
// (listas de arquivos/rows, caminhos).
func DiagInfo(key, msg string, details map[string]any) {
	diag(zerolog.InfoLevel, key, msg, details)
}

// DiagWarn é o DiagInfo em nível warning — divergências, arquivos fora de
// data/ e degradações, sempre com contexto de investigação no arquivo.
func DiagWarn(key, msg string, details map[string]any) {
	diag(zerolog.WarnLevel, key, msg, details)
}

func diag(level zerolog.Level, key, msg string, details map[string]any) {
	Init()
	ev := diagLog.WithLevel(level).Str("key", key)
	for _, k := range sortedKeys(details) {
		ev = ev.Interface(k, details[k])
	}
	ev.Msg(msg)
	// Console: linha curta colorida — o detalhe fica no arquivo.
	emit(level, "%s: %s", key, msg)
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FromFrontend persiste um log vindo do frontend no MESMO sistema
// (console colorido + arquivo JSON, com source=frontend).
//
// Entrada do frontend não é confiável: nível é validado contra a lista,
// message e valores de campo são truncados, e as chaves dos campos são
// sanitizadas antes de virarem chaves JSON.
func FromFrontend(level, message string, fields map[string]any) error {
	Init()

	var lvl zerolog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn", "warning":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	default:
		return fmt.Errorf("nível de log inválido: %q", level)
	}

	msg := truncate(message, maxFrontendMessage)
	if msg == "" {
		msg = "(vazio)"
	}

	ev := log.WithLevel(lvl).Str("source", "frontend")
	for _, k := range sanitizedFieldKeys(fields) {
		ev = ev.Interface(k, truncateAny(fields[k]))
	}
	ev.Msg(msg)
	return nil
}

const (
	maxFrontendMessage = 4000
	maxFrontendFields  = 16
	maxFieldName       = 32
	maxFieldValue      = 1000
)

// sanitizedFieldKeys filtra e normaliza as chaves dos campos: apenas
// [A-Za-z0-9_.-], no máximo maxFieldName runes e maxFrontendFields campos.
func sanitizedFieldKeys(fields map[string]any) []string {
	if len(fields) == 0 {
		return nil
	}
	out := make([]string, 0, len(fields))
	for k := range fields {
		if len(out) == maxFrontendFields {
			break
		}
		clean := make([]rune, 0, maxFieldName)
		for _, r := range k {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
				r == '_' || r == '.' || r == '-' {
				clean = append(clean, r)
			}
			if len(clean) == maxFieldName {
				break
			}
		}
		if len(clean) == 0 {
			continue
		}
		out = append(out, string(clean))
	}
	sort.Strings(out)
	return out
}

func truncateAny(v any) any {
	if s, ok := v.(string); ok {
		return truncate(s, maxFieldValue)
	}
	return v
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
