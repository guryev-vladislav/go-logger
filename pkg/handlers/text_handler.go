//nolint:unused
package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
)

type TextHandler struct {
	out         io.Writer
	level       slog.Level
	mu          sync.Mutex
	useColor    bool
	serviceName string
}

func NewTextHandler(out io.Writer, level slog.Level, useColor bool, serviceName string) *TextHandler {
	return &TextHandler{
		out:         out,
		level:       level,
		useColor:    useColor,
		serviceName: serviceName,
	}
}

func (h *TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *TextHandler) Handle(ctx context.Context, rec slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.useColor = true
	timeStr := rec.Time.Format(timeFormat)
	levelStr := strings.ToUpper(rec.Level.String())
	levelStr = h.colorizeLevel(rec.Level, levelStr)

	var (
		file     string
		line     int
		funcName string
	)

	if rec.PC != 0 {
		fn := runtime.FuncForPC(rec.PC)
		if fn != nil {
			funcName = TrimFuncName(fn.Name())
			file, line = fn.FileLine(rec.PC)
			file = TrimFilePath(file)
		}
	}

	msg := rec.Message
	if rec.NumAttrs() > 0 {
		var attrs []string

		rec.Attrs(func(a slog.Attr) bool {
			if a.Key != serviceAttr && a.Key != funcNameAttr && a.Key != sourceAttr {
				attrs = append(attrs, fmt.Sprintf(attrFormat, a.Key, a.Value.Any()))
			}

			return true
		})

		if len(attrs) > 0 {
			msg = fmt.Sprintf(msgFormat, msg, strings.Join(attrs, ", "))
		}
	}

	if h.useColor {
		timeStr = colorGray + timeStr + colorReset
		file = colorBlue + file + colorReset
		funcName = colorBold + funcName + colorReset
	}

	_, err := fmt.Fprintf(
		h.out,
		outputFormat,
		timeStr,
		levelStr,
		h.serviceName,
		file,
		line,
		funcName,
		msg,
	)

	return err
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TextHandler{
		out:         h.out,
		level:       h.level,
		useColor:    h.useColor,
		serviceName: h.serviceName,
	}
}

func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *TextHandler) colorizeLevel(level slog.Level, str string) string {
	if !h.useColor {
		return str
	}

	switch level {
	case slog.LevelDebug:
		return colorGray + str + colorReset
	case slog.LevelInfo:
		return colorBlue + str + colorReset
	case slog.LevelWarn:
		return colorYellow + str + colorReset
	case slog.LevelError:
		return colorRed + str + colorReset
	default:
		return str
	}
}
