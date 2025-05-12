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

var once sync.Once

var log zerolog.Logger

func NewLoggerHandler(moduleName string) ILoggerService {
	return &LoggerService{
		Logger: Get().With().Str("module", moduleName).Logger(),
	}
}

func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.DebugLevel) // default to INFO
		}

		var consoleWriter io.Writer = zerolog.ConsoleWriter{
			Out:          os.Stdout,
			TimeFormat:   time.RFC822,
			TimeLocation: time.UTC,
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
			PartsExclude: []string{
				zerolog.TimestampFieldName,
			},
		}

		var output io.Writer = consoleWriter

		if os.Getenv("APP_ENV") != "development" {
			currentTime := time.Now().Format("02-01-2006")
			fileName := fmt.Sprintf("ffx_tracker-%s.log", currentTime)

			fileLogger := &lumberjack.Logger{
				Filename:   fileName,
				MaxSize:    5, //
				MaxBackups: 10,
				MaxAge:     14,
				Compress:   true,
			}

			output = zerolog.MultiLevelWriter(consoleWriter, fileLogger)
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Logger()
	})

	return log
}
