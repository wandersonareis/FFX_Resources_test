// Package loggingService é o sistema único de logs do aplicativo: console
// colorido por nível + persistência em arquivo (lumberjack). O arquivo é
// JSON — cabe mais detalhe (caller, campos estruturados) do que a linha
// curta do console; console e arquivo não precisam ser idênticos.
//
// UM arquivo por início de app, em `logs/` ao lado do EXECUTÁVEL (nunca
// relativo ao CWD — o destino não pode mudar dependendo de quem lança o
// processo: app, `go test`, atalho com outro "Início em"):
//
//	logs/ffx-<AAAA-MM-DD_HH-MM-SS>.log
//
// O diagnóstico (DiagInfo/DiagWarn) está FUNDIDO neste mesmo arquivo: a
// linha curta vai ao console e o JSON completo com os detalhes (listas de
// rows, caminhos) vai ao arquivo, identificado pelo campo `key`. Arquivos
// de execuções passadas além da retenção são podados no Init.
package loggingService

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once sync.Once

	// log é o logger geral: console colorido + arquivo (JSON) — todo o app.
	log zerolog.Logger

	// diagLog escreve os eventos de diagnóstico (DiagInfo/DiagWarn) com o
	// JSON completo e detalhes — no MESMO arquivo de `log` (campo `key`);
	// o console recebe só a linha curta (via consoleLog).
	diagLog zerolog.Logger

	// consoleLog emite a linha curta do diagnóstico SÓ no console: o
	// arquivo já recebeu o evento completo, não precisa duplicar a linha.
	consoleLog zerolog.Logger

	logDirPath string

	// consoleOutWriter pode ser trocado em teste (ResetForTest) para
	// silenciar o stdout.
	consoleOutWriter io.Writer = os.Stdout

	// logDirOverride aponta os arquivos para outro diretório (ResetForTest).
	logDirOverride string

	// arquivo ativo desta execução — fechado no ResetForTest (Windows
	// mantém o arquivo travado enquanto o handle existe).
	activeAppFile *lumberjack.Logger
)

const (
	diagVerboseEnv = "VERBOSE_MODE"
	logLevelEnv    = "LOG_LEVEL"

	// logFilePrefix prefixia todo arquivo gerado (inclusive os backups da
	// rotação) — é o critério da poda de execuções antigas.
	logFilePrefix = "ffx-"

	// logRetention é a idade máxima de um arquivo de log de execução
	// anterior; o Init apaga o que estiver além dela.
	logRetention = 30 * 24 * time.Hour
)

// LogDir é o diretório onde os logs são persistidos: absoluto (ao lado do
// executável) no fluxo de app, o diretório de teste depois de ResetForTest.
// Vazio quando não há persistência (só console).
func LogDir() string {
	Init()
	return logDirPath
}

// Get devolve o logger geral (console colorido + arquivo). É o mesmo
// zerolog.Logger usado pelo app para panics.
func Get() zerolog.Logger {
	Init()
	return log
}

// Init inicializa o sistema de logs uma única vez.
func Init() {
	once.Do(initLoggers)
}

func initLoggers() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	consoleWriter := zerolog.ConsoleWriter{
		Out:        consoleOut(),
		TimeFormat: "15:04:05",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{zerolog.TimestampFieldName},
	}

	// UM arquivo por início de app, em `logs/` ao lado do executável.
	// Sem base resolvível ou sem pasta: segue só no console — logging não
	// pode derrubar a aplicação.
	dir := logDir()
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "loggingService: não criou %s: %v\n", dir, err)
		} else {
			pruneOldLogs(dir)
			activeAppFile = rollingFile(filepath.Join(dir, startFileName()))
			logDirPath = dir
		}
	}

	// Console re-renderiza o evento JSON (colorido por nível); o arquivo
	// recebe o JSON cru — mesmos campos, formatos distintos.
	var sink io.Writer = consoleWriter
	if activeAppFile != nil {
		sink = zerolog.MultiLevelWriter(consoleWriter, activeAppFile)
	}
	log = zerolog.New(sink).
		Level(resolveLevel()).
		With().
		Timestamp().
		Logger()
	consoleLog = zerolog.New(consoleWriter).
		Level(resolveLevel()).
		With().
		Timestamp().
		Logger()

	// Diagnóstico: MESMO arquivo do app (o campo `key` distingue), sem
	// repetir a linha curta. Sem arquivo (pasta não criada) os detalhes vão
	// ao console — nada se perde na degradação.
	if activeAppFile != nil {
		diagLog = zerolog.New(activeAppFile).With().Timestamp().Logger()
	} else {
		diagLog = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
}

// closeFiles solta o handle de arquivo (Windows trava o arquivo aberto).
func closeFiles() {
	if activeAppFile != nil {
		_ = activeAppFile.Close()
	}
	activeAppFile = nil
}

// logDir devolve o diretório de persistência: o override de teste ou
// `logs/` ao lado do EXECUTÁVEL. NUNCA relativo ao CWD — o destino não
// pode mudar dependendo de quem lança o processo (app, `go test`, atalho).
// Vazio quando o executável não é resolvível (fica só o console).
//
// O cálculo de common.GetExecDir é local: não se importa backend/common
// aqui, que é o pacote que importa este (fachada de log) — seria ciclo.
func logDir() string {
	if logDirOverride != "" {
		return logDirOverride
	}
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "logs")
}

func consoleOut() io.Writer {
	if consoleOutWriter != nil {
		return consoleOutWriter
	}
	return os.Stdout
}

// rollingFile é a rotação padrão do log de aplicação: por segurança em
// sessões muito longas — a retenção dos arquivos de execuções passadas é
// da poda do Init (logRetention).
func rollingFile(name string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   name,
		MaxSize:    5,
		MaxBackups: 10,
		MaxAge:     int(logRetention.Hours() / 24),
		Compress:   true,
	}
}

// resolveLevel: o logger geral roda em Debug (nada é cortado no zerolog) —
// o corte do debug/verbose é POR CHAMADA, para que o VERBOSE_MODE ligado em
// runtime (env alterado depois do Init) continue valendo.
func resolveLevel() zerolog.Level {
	return zerolog.DebugLevel
}

func verboseEnabled() bool {
	return os.Getenv(diagVerboseEnv) == "1"
}

// debugEnabled decide se eventos debug são emitidos: verbose ligado ou
// LOG_LEVEL explícito no nível debug.
func debugEnabled() bool {
	if verboseEnabled() {
		return true
	}
	if raw, err := strconv.Atoi(strings.TrimSpace(os.Getenv(logLevelEnv))); err == nil &&
		raw <= int(zerolog.DebugLevel) {
		return true
	}
	return false
}

// startFileName é o nome do arquivo único DESTA execução: o timestamp de
// início separa dois starts no mesmo dia (ordenável cronologicamente).
func startFileName() string {
	return logFilePrefix + time.Now().Format("2006-01-02_15-04-05") + ".log"
}

// pruneOldLogs apaga, em `dir`, os arquivos de log (do prefixo logFilePrefix,
// inclusive backups da rotação) mais velhos que logRetention. Roda no Init,
// antes de abrir o arquivo desta execução — o MaxAge do lumberjack só
// cuida dos backups. Erros são ignorados: logging não pode derrubar a
// aplicação.
func pruneOldLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	limit := time.Now().Add(-logRetention)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), logFilePrefix) {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil || !info.ModTime().Before(limit) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

// ResetForTest reinicializa o sistema de logs escrevendo em `dir` e
// silenciando o console (os testes asserem os ARQUIVOS, não o stdout).
func ResetForTest(dir string) {
	closeFiles()
	once = sync.Once{}
	logDirOverride = dir
	consoleOutWriter = io.Discard
	logDirPath = ""
	Init()
}
