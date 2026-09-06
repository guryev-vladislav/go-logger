package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/guryev-vladislav/tracelog/pkg/handlers"
)

type slogLogger struct {
	logger       *slog.Logger
	fields       []slog.Attr
	functionName string
}

func (l *slogLogger) Debug(msg string, fields ...slog.Attr) {
	l.log(slog.LevelDebug, msg, fields...)
}

func (l *slogLogger) Info(msg string, fields ...slog.Attr) {
	l.log(slog.LevelInfo, msg, fields...)
}

func (l *slogLogger) Warning(msg string, fields ...slog.Attr) {
	l.log(slog.LevelWarn, msg, fields...)
}

func (l *slogLogger) Error(msg string, fields ...slog.Attr) {
	l.log(slog.LevelError, msg, fields...)
}

func (l *slogLogger) ErrorIn(funcName string, err error, fields ...slog.Attr) {
	if err == nil {
		return
	}

	allFields := append(
		[]slog.Attr{
			slog.String(errorAttr, err.Error()),
			slog.String(funcNameAttr, funcName),
		},
		fields...,
	)

	l.log(slog.LevelError, fmt.Sprintf(msgCompletesError, funcName), allFields...)
}

func (l *slogLogger) Fatal(msg string, fields ...slog.Attr) {
	l.log(slog.LevelError, fmt.Sprintf(msgFatal, msg), fields...)
	os.Exit(1)
}

func (l *slogLogger) Panic(msg string, fields ...slog.Attr) {
	l.log(slog.LevelError, fmt.Sprintf(msgPanic, msg), fields...)
	panic(msg)
}

func (l *slogLogger) With(fields ...slog.Attr) {
	l.fields = append(l.fields, fields...)
}

func (l *slogLogger) End() {
	defer l.handlePanic()

	if l.logger != nil && l.functionName != emptyString {
		l.logEndMessage()
	}
}

func (l *slogLogger) ReturningParams(fields ...slog.Attr) {
	l.log(slog.LevelDebug, msgReturning, fields...)
}

func (l *slogLogger) SIP(direction string, msg string, fields ...slog.Attr) {
	if strings.TrimSpace(msg) == emptyString {
		l.log(slog.LevelError, sipLogFailedMsg,
			slog.String(errorAttr, ErrEmptySIPMessage.Error()))

		return
	}

	direction = strings.ToUpper(direction)
	if direction != SIPDirectionSent && direction != SIPDirectionReceived {
		errMsg := fmt.Sprintf("%s: %s", ErrInvalidSIPDirection.Error(), direction)
		l.log(slog.LevelError, sipLogFailedMsg,
			slog.String(errorAttr, errMsg),
			slog.String(sipFieldProvidedDir, direction))

		return
	}

	timestamp := time.Now()
	formattedBlock := handlers.FormatSIPBlock(direction, msg, timestamp)
	l.log(slog.LevelInfo, formattedBlock, fields...)
}

func (l *slogLogger) handlePanic() {
	if err := recover(); err != nil {
		if l.logger != nil {
			l.logPanicError(err)
		}

		panic(err)
	}
}

func (l *slogLogger) logPanicError(err any) {
	pc, file, line, funcName := getCallerPCFast()
	rec := slog.NewRecord(time.Now(), slog.LevelError, msgPanicCatched, pc)
	rec.AddAttrs(
		slog.Any(fieldErr, err),
		slog.String(fieldStack, string(debug.Stack())),
		slog.String(fieldFunc, l.functionName),
	)

	if funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			rec.AddAttrs(slog.String(sourceAttr, fmt.Sprintf(sourceFormat, trimmed, line)))
		}
	}

	handler := l.logger.Handler()
	if handler.Enabled(context.Background(), slog.LevelError) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (l *slogLogger) logEndMessage() {
	pc, file, line, funcName := getCallerPCFast()
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(suffixEnd, l.functionName), pc)

	if funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			rec.AddAttrs(slog.String(sourceAttr, fmt.Sprintf(sourceFormat, trimmed, line)))
		}
	}

	handler := l.logger.Handler()
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (l *slogLogger) log(level slog.Level, msg string, fields ...slog.Attr) {
	pc, file, line, funcName := getCallerPCFast()
	rec := slog.NewRecord(time.Now(), level, msg, pc)

	allFields := l.buildFields(funcName, file, line, fields)

	rec.AddAttrs(allFields...)

	handler := l.logger.Handler()
	if handler.Enabled(context.Background(), level) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (l *slogLogger) buildFields(funcName, file string, line int, fields []slog.Attr) []slog.Attr {
	allFields := make([]slog.Attr, 0, len(l.fields)+len(fields)+2)
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	hasFuncName, hasSource := l.checkExistingFields(allFields)

	if !hasFuncName && funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			allFields = append(allFields, slog.String(funcNameAttr, trimmed))
		}
	}

	if !hasSource && file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			source := fmt.Sprintf(sourceFormat, trimmed, line)
			allFields = append(allFields, slog.String(sourceAttr, source))
		}
	}

	return allFields
}

func (l *slogLogger) checkExistingFields(fields []slog.Attr) (bool, bool) {
	hasFuncName := false
	hasSource := false

	for _, attr := range fields {
		if attr.Key == funcNameAttr {
			hasFuncName = true
		}

		if attr.Key == sourceAttr {
			hasSource = true
		}
	}

	return hasFuncName, hasSource
}
