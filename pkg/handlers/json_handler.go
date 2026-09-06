package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
)

type JSONHandlerWrapper struct {
	handler slog.Handler
}

func NewJSONHandlerWrapper(handler slog.Handler, serviceName string) *JSONHandlerWrapper {
	return &JSONHandlerWrapper{
		handler: handler,
	}
}

func (h *JSONHandlerWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *JSONHandlerWrapper) Handle(ctx context.Context, rec slog.Record) error {
	if rec.PC != 0 {
		var (
			hasFuncName bool
			hasSource   bool
		)

		rec.Attrs(func(a slog.Attr) bool {
			if a.Key == funcNameAttr {
				hasFuncName = true
			}

			if a.Key == sourceAttr {
				hasSource = true
			}

			return true
		})

		fn := runtime.FuncForPC(rec.PC)
		if fn != nil {
			if !hasFuncName {
				funcName := TrimFuncName(fn.Name())
				if funcName != "" {
					rec.AddAttrs(slog.String(funcNameAttr, funcName))
				}
			}

			if !hasSource {
				file, line := fn.FileLine(rec.PC)

				file = TrimFilePath(file)
				if file != "" {
					source := fmt.Sprintf("%s:%d", file, line)
					rec.AddAttrs(slog.String(sourceAttr, source))
				}
			}
		}
	}

	return h.handler.Handle(ctx, rec)
}

func (h *JSONHandlerWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &JSONHandlerWrapper{
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *JSONHandlerWrapper) WithGroup(name string) slog.Handler {
	return &JSONHandlerWrapper{
		handler: h.handler.WithGroup(name),
	}
}
