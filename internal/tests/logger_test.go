//nolint:unused
package tests

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	logger "github.com/guryev-vladislav/go-logger/pkg/logger"
)

var ErrTestError = errors.New("test error")

const (
	serviceVersion = "1.0.0"
	serviceName    = "test-service"
)

func assertLoggerContext(t *testing.T, call func(context.Context) (context.Context, logger.Logger)) {
	t.Helper()

	newCtx, testLogger := call(context.Background())
	if newCtx == nil || testLogger == nil {
		t.Error("expected non-nil context and logger")
	}
}

func TestSlogLogger(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	if factory == nil {
		t.Fatal("factory is nil")
	}

	ctx := context.Background()

	_, testLogger := factory.GetLoggerFromContext(ctx)
	if testLogger == nil {
		t.Fatal("logger is nil")
	}

	defer func() {
		factory.ForceFlush(ctx)
		time.Sleep(100 * time.Millisecond)
	}()

	t.Run("Debug", func(t *testing.T) {
		testLogger.Debug("debug message")
		testLogger.Debug("debug message with fields", slog.String("key", "value"))
	})
	t.Run("Info", func(t *testing.T) {
		testLogger.Info("info message")
		testLogger.Info("info message with fields", slog.String("key", "value"))
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger.Warning("warning message")
		testLogger.Warning("warning message with fields", slog.String("key", "value"))
	})
	t.Run("Error", func(t *testing.T) {
		testLogger.Error("error message")
		testLogger.Error("error message with fields", slog.String("key", "value"))
	})
	t.Run("ErrorIn", func(t *testing.T) {
		testLogger.ErrorIn("testFunc", ErrTestError)
		testLogger.ErrorIn("testFunc", ErrTestError, slog.String("key", "value"))
	})
	t.Run("ReturningParams", func(t *testing.T) {
		testLogger.ReturningParams()
		testLogger.ReturningParams(slog.String("key", "value"))
	})
	testLogger.End()
}

func TestSlogLoggerFactory(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion})
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	tests := []struct {
		name string
		call func(context.Context) (context.Context, logger.Logger)
	}{
		{name: "GetLoggerFromContext", call: func(ctx context.Context) (context.Context, logger.Logger) {
			return factory.GetLoggerFromContext(ctx)
		}},
		{name: "GetLoggerFromContextWithFields", call: func(ctx context.Context) (context.Context, logger.Logger) {
			return factory.GetLoggerFromContext(ctx, slog.String("param", "value"))
		}},
		{name: "GetLoggerFromContextWithSpanName", call: func(ctx context.Context) (context.Context, logger.Logger) {
			return factory.GetLoggerFromContextWithSpanName(ctx, "customSpan")
		}},
		{name: "GetLoggerFromContextWithEmptySpanName", call: func(ctx context.Context) (context.Context, logger.Logger) {
			return factory.GetLoggerFromContextWithSpanName(ctx, "")
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertLoggerContext(t, test.call)
		})
	}

	t.Run("ForceFlush", func(t *testing.T) {
		factory.ForceFlush(context.Background())
	})
}

func TestSlogLoggerPanic(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion})
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	_, testLogger := factory.GetLoggerFromContext(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but did not occur")
		}
	}()

	testLogger.Panic("test panic")
}

func TestSlogLoggerPanicWithFields(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion})
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	_, testLogger := factory.GetLoggerFromContext(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but did not occur")
		}
	}()

	testLogger.Panic("test panic with fields", slog.String("key", "value"))
}

func TestSlogLoggerEndWithPanic(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion})
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	_, testLogger := factory.GetLoggerFromContext(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but did not occur")
		}
	}()

	func() {
		defer testLogger.End()

		panic("test panic")
	}()
}
