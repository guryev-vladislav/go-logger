package logger

import (
	"log/slog"
	"os"
	"runtime"
	"strings"
)

type LoggerConfig struct {
	ServiceName    string
	Version        string
	JaegerEndpoint string
	LogFilePath    string
}

func New(cfg LoggerConfig) (LoggerFactory, error) {
	logFilePath := cfg.LogFilePath
	if logFilePath == emptyString {
		logFilePath = os.Getenv(EnvKeyLogFile)
	}

	var level = slog.LevelInfo
	if os.Getenv(EnvKeyLoggerMinLevel) != emptyString {
		level = getSlogLevelByEnv()
	}

	if os.Getenv(EnvKeyLoggerDst) == Local {
		return NewSlogLoggerFactory(
			cfg.ServiceName,
			cfg.Version,
			logFilePath,
			os.Getenv(EnvKeyStandDomain),
			level,
		)
	}

	return NewTraceLoggerFactory(
		cfg.ServiceName,
		cfg.Version,
		logFilePath,
		os.Getenv(EnvKeyStandDomain),
		cfg.JaegerEndpoint,
		level,
	)
}

func getFunctionName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown function name"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown function name"
	}

	parts := strings.Split(fn.Name(), "/")
	funcName := parts[len(parts)-1]

	parts = strings.Split(funcName, ".")
	funcName = parts[len(parts)-1]

	return funcName
}

func getSlogLevelByEnv() slog.Level {
	levelStr := strings.ToUpper(os.Getenv(EnvKeyLoggerMinLevel))

	switch levelStr {
	case debugLevel:
		return slog.LevelDebug
	case infoLevel:
		return slog.LevelInfo
	case warningLevel:
		return slog.LevelWarn
	case errLevel:
		return slog.LevelError
	case fatalLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
