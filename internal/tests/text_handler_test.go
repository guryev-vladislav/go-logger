package tests

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/guryev-vladislav/tracelog/pkg/handlers"
)

func TestTextHandlerPreservesAttrsAndGroupsWithoutColor(t *testing.T) {
	var output bytes.Buffer
	handler := handlers.NewTextHandler(&output, slog.LevelInfo, false, "service")
	logger := slog.New(handler).With(slog.String("request_id", "abc"))

	logger.WithGroup("http").InfoContext(context.Background(), "request", slog.Int("status", 200))

	result := output.String()
	if strings.Contains(result, "\033[") {
		t.Fatalf("expected plain output, got ANSI escape sequence: %q", result)
	}

	for _, expected := range []string{"request_id: abc", "http.status: 200"} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected %q in output %q", expected, result)
		}
	}
}

func TestTextHandlerFormatsRecord(t *testing.T) {
	var output bytes.Buffer
	handler := handlers.NewTextHandler(&output, slog.LevelInfo, false, "service")
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
	record.AddAttrs(slog.String("key", "value"))

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !strings.Contains(output.String(), "message {key: value}") {
		t.Fatalf("expected formatted attributes, got %q", output.String())
	}
}

func TestTextHandlerUsesLevelSpecificColors(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		color string
	}{
		{name: "Info", level: slog.LevelInfo, color: "\033[34m"},
		{name: "Warning", level: slog.LevelWarn, color: "\033[33m"},
		{name: "Error", level: slog.LevelError, color: "\033[31m"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			handler := handlers.NewTextHandler(&output, slog.LevelDebug, true, "service")
			record := slog.NewRecord(time.Now(), test.level, test.name, 0)

			if err := handler.Handle(context.Background(), record); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			result := output.String()
			if !strings.Contains(result, test.color+strings.ToUpper(test.level.String())) {
				t.Fatalf("expected %s color in output %q", test.name, result)
			}
		})
	}
}
