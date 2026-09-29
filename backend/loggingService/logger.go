// Package loggingService é o sistema único de logs do aplicativo: console
// colorido por nível + persistência em arquivo (lumberjack). O arquivo é
// JSON — cabe mais detalhe (caller, campos estruturados) do que a linha
// curta do console; console e arquivo não precisam ser idênticos.
//
// Dois destinos de arquivo em logs/:
//   - ffx-<data>.log         log geral (tudo);
//   - diagnostico-<data>.log eventos de divergência/inventário com os
//     detalhes completos (listas de rows, caminhos).
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

	// diagLog persiste os eventos de diagnóstico em JSON detalhado; console
	// recebe apenas a linha curta (via log).
	diagLog zerolog.Logger

	logDirPath string

	// consoleOutWriter pode ser trocado em teste (ResetForTest) para
	// silenciar o stdout.
	consoleOutWriter io.Writer = os.Stdout

	// logDirOverride aponta os arquivos para outro diretório (ResetForTest).
	logDirOverride string

	// escritores ativos de arquivo — fechados no ResetForTest (Windows
	// mantém o arquivo travado enquanto o handle existe).
	activeAppFile  *lumberjack.Logger
	activeDiagFile *lumberjack.Logger
)

const (
	diagVerboseEnv = "VERBOSE_MODE"
	logLevelEnv    = "LOG_LEVEL"
)

// LogDir é o diretório onde os logs são persistidos.
func LogDir() string {
	if logDirPath != "" {
		return logDirPath
	}
	return "logs"
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

	dir := logDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		// Sem pasta de log: segue só no console — logging não pode derrubar
		// a aplicação.
		fmt.Fprintf(os.Stderr, "loggingService: não criou %s: %v\n", dir, err)
	}

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

	appFile := rollingFile(filepath.Join(dir, "ffx-"+today()+".log"))
	diagFile := rollingFile(filepath.Join(dir, "diagnostico-"+today()+".log"))

	// Console re-renderiza o evento JSON (colorido por nível); o arquivo
	// recebe o JSON cru — mesmos campos, formatos distintos.
	log = zerolog.New(zerolog.MultiLevelWriter(consoleWriter, appFile)).
		Level(resolveLevel()).
		With().
		Timestamp().
		Logger()

	// Diagnóstico: arquivo dedicado, sempre JSON, sem ruído no console.
	diagLog = zerolog.New(diagFile).With().Timestamp().Logger()
	activeAppFile = appFile
	activeDiagFile = diagFile
	logDirPath = dir
}

// closeFiles solta os handles de arquivo (Windows trava o arquivo aberto).
func closeFiles() {
	if activeAppFile != nil {
		_ = activeAppFile.Close()
	}
	if activeDiagFile != nil {
		_ = activeDiagFile.Close()
	}
	activeAppFile, activeDiagFile = nil, nil
}

func logDir() string {
	if logDirOverride != "" {
		return logDirOverride
	}
	return "logs"
}

func consoleOut() io.Writer {
	if consoleOutWriter != nil {
		return consoleOutWriter
	}
	return os.Stdout
}

// rollingFile é a rotação padrão do log de aplicação.
func rollingFile(name string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   name,
		MaxSize:    5,
		MaxBackups: 10,
		MaxAge:     30,
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

func today() string {
	return time.Now().Format("02-01-2006")
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
