package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/guryev-vladislav/go-logger/pkg/handlers"
)

type slogLoggerFactory struct {
	logger *slog.Logger
}

func NewSlogLoggerFactory(
	serviceName, version, logFilePath, standDomain string,
	level slog.Level,
) (LoggerFactory, error) {
	var (
		consoleHandler slog.Handler
		fileHandler    slog.Handler
		fileWriter     io.Writer
	)

	consoleHandler = handlers.NewTextHandler(os.Stdout, level, handlers.IsTerminal(os.Stdout), serviceName)

	if logFilePath != emptyString {
		fileWriter = &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    logFileMB,
			MaxBackups: logFileBackups,
			MaxAge:     logFileAge,
			Compress:   logFileCompress,
		}

		baseJSONHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
			Level: level,
		})

		fileHandler = handlers.NewJSONHandlerWrapper(baseJSONHandler, serviceName)
	}

	var handler slog.Handler
	if fileHandler != nil {
		handler = handlers.NewMultiHandler(consoleHandler, fileHandler)
	} else {
		handler = consoleHandler
	}

	logger := slog.New(handler).With(
		slog.String(serviceAttr, serviceName),
		slog.String(versionAttr, version),
	)

	if standDomain != emptyString {
		logger = logger.With(slog.String(standAttr, standDomain))
	}

	return &slogLoggerFactory{
		logger: logger,
	}, nil
}

func (f *slogLoggerFactory) GetLoggerFromContext(
	ctx context.Context,
	fields ...slog.Attr,
) (context.Context, Logger) {
	functionName := getFunctionName()

	f.logStartMessage(ctx, fields)

	return ctx, &slogLogger{
		logger:       f.logger,
		functionName: functionName,
	}
}

func (f *slogLoggerFactory) GetLoggerFromContextWithSpanName(
	ctx context.Context,
	spanName string,
	fields ...slog.Attr,
) (context.Context, Logger) {
	if spanName == emptyString {
		spanName = getFunctionName()
	}

	functionName := spanName
	if functionName == emptyString {
		functionName = getFunctionName()
	}

	if len(fields) > 0 {
		args := make([]any, 0, len(fields)*2)
		for _, attr := range fields {
			args = append(args, attr.Key, attr.Value.Any())
		}

		f.logger.InfoContext(ctx, fmt.Sprintf(prefixFormat, functionName, msgStartWith), args...)
	} else {
		f.logger.InfoContext(ctx, fmt.Sprintf(prefixFormat, functionName, msgStart))
	}

	return ctx, &slogLogger{
		logger:       f.logger,
		functionName: functionName,
	}
}

func (f *slogLoggerFactory) ForceFlush(context.Context) {
}

func (f *slogLoggerFactory) logStartMessage(ctx context.Context, fields []slog.Attr) {
	pc, file, line, callerFuncName := getCallerPCFast()
	functionName := getFunctionName()

	var msg string
	if len(fields) > 0 {
		msg = fmt.Sprintf(prefixFormat, functionName, msgStartWith)
	} else {
		msg = fmt.Sprintf(prefixFormat, functionName, msgStart)
	}

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pc)

	if callerFuncName != emptyString {
		callerFuncName = handlers.TrimFuncName(callerFuncName)
		if callerFuncName != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, callerFuncName))
		}
	}

	if file != emptyString {
		file = handlers.TrimFilePath(file)
		if file != emptyString {
			source := fmt.Sprintf(sourceFormat, file, line)
			rec.AddAttrs(slog.String(sourceAttr, source))
		}
	}

	if len(fields) > 0 {
		rec.AddAttrs(fields...)
	}

	handler := f.logger.Handler()
	if handler.Enabled(ctx, slog.LevelInfo) {
		if err := handler.Handle(ctx, rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}
