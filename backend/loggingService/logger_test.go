package loggingService_test

import (
	"os"
	"strings"
	"testing"

	"ffxresources/backend/loggingService"
)

// reset reinicializa o sistema de logs em um diretório temporário com o
// console silenciado: os testes asserem os ARQUIVOS.
func resetForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	loggingService.ResetForTest(dir)
	t.Cleanup(func() { loggingService.ResetForTest(dir) })
	return dir
}

// readFile devolve o conteúdo de um arquivo de log.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ler %s: %v", path, err)
	}
	return string(b)
}

// diagLogContent devolve o conteúdo do log de diagnóstico ("" quando ausente).
func diagLogContent(t *testing.T, dir string) string {
	t.Helper()
	_, diag := logFilesIn(dir)
	if diag == "" {
		return ""
	}
	return readFile(t, dir+"\\"+diag)
}

// appLogContent devolve o conteúdo do log geral ("" quando ausente).
func appLogContent(t *testing.T, dir string) string {
	t.Helper()
	app, _ := logFilesIn(dir)
	if app == "" {
		return ""
	}
	return readFile(t, dir+"\\"+app)
}

func logFilesIn(dir string) (app, diag string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "ffx-") {
			app = e.Name()
		}
		if strings.HasPrefix(e.Name(), "diagnostico-") {
			diag = e.Name()
		}
	}
	return app, diag
}

// Info/Warn vão para console+arquivo; o arquivo é JSON com os campos.
func TestInfoAndWarnPersistToJSONFile(t *testing.T) {
	dir := resetForTest(t)

	loggingService.Info("carregando %s com %d arquivos", "events", 7)
	loggingService.Warn("divergência em %s", "command.bin")

	content := appLogContent(t, dir)
	if content == "" {
		t.Fatal("arquivo de log geral ausente")
	}
	if !strings.Contains(content, `"level":"info"`) ||
		!strings.Contains(content, `"message":"carregando events com 7 arquivos"`) {
		t.Fatalf("info ausente no arquivo: %s", content)
	}
	if !strings.Contains(content, `"level":"warn"`) ||
		!strings.Contains(content, `"message":"divergência em command.bin"`) {
		t.Fatalf("warn ausente no arquivo: %s", content)
	}
}

// DiagWarn: console recebe a linha curta; o ARQUIVO de diagnóstico recebe
// o JSON completo com os detalhes (listas, contagens).
func TestDiagWarnWritesFullDetailsToDiagnosticsFile(t *testing.T) {
	dir := resetForTest(t)

	loggingService.DiagWarn("objects/command.bin (ffx2)",
		"original e tradução divergem — 424 de 429 rows casaram (5 rows só em mods, 0 rows só em data)",
		map[string]any{
			"kind":      "objects",
			"id":        "command.bin",
			"data_rows": 424,
			"mods_rows": 429,
			"rows_somente_mods": []map[string]any{
				{"index": 18, "name": "name", "snippet": "Berserker"},
			},
		})

	content := diagLogContent(t, dir)
	if content == "" {
		t.Fatal("arquivo de diagnóstico ausente")
	}
	for _, want := range []string{
		`"key":"objects/command.bin (ffx2)"`,
		`"level":"warn"`,
		`"data_rows":424`,
		`"rows_somente_mods"`,
		`"snippet":"Berserker"`,
		"original e tradução divergem",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("diagnóstico sem %q:\n%s", want, content)
		}
	}
}

// Verbose respeita VERBOSE_MODE: desligado não emite nada.
func TestVerboseRespectsEnv(t *testing.T) {
	dir := resetForTest(t)
	t.Setenv("VERBOSE_MODE", "0")

	loggingService.Verbose("não deveria emitir: %d", 42)

	if c := appLogContent(t, dir); strings.Contains(c, "não deveria emitir") {
		t.Fatalf("verbose deveria ser silenciado sem VERBOSE_MODE: %s", c)
	}

	t.Setenv("VERBOSE_MODE", "1")
	loggingService.Verbose("agora emitir: %d", 7)
	if c := appLogContent(t, dir); !strings.Contains(c, `"message":"agora emitir: 7"`) {
		t.Fatalf("verbose deveria emitir com VERBOSE_MODE=1: %q", c)
	}
}

// FromFrontend: nível validado, campos sanitizados/truncados e gravação no
// MESMO arquivo geral com source=frontend.
func TestFromFrontendValidationAndPersist(t *testing.T) {
	dir := resetForTest(t)

	if err := loggingService.FromFrontend("banana", "oi", nil); err == nil {
		t.Fatal("nível inválido deveria ser rejeitado")
	}
	if err := loggingService.FromFrontend("info", "", nil); err != nil {
		t.Fatalf("mensagem vazia deveria virar (vazio): %v", err)
	}

	if err := loggingService.FromFrontend("warn", "célula editada sem tag protegida",
		map[string]any{"arquivo": "command.bin", "chave@ruim": "valor"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	app, _ := logFilesIn(dir)
	if app == "" {
		t.Fatal("arquivo de log geral ausente")
	}
	content := readFile(t, dir+"\\"+app)
	if !strings.Contains(content, `"source":"frontend"`) ||
		!strings.Contains(content, `"level":"warn"`) ||
		!strings.Contains(content, `"arquivo":"command.bin"`) {
		t.Fatalf("log do frontend ausente:\n%s", content)
	}
	// Campo com chave sanitizada gravado; chave inválida descartada.
	if strings.Contains(content, `"chave@ruim"`) {
		t.Fatalf("chave não sanitizada no arquivo:\n%s", content)
	}
}
