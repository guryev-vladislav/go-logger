//nolint:unused
package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	// Debug logs a debug msg with fields.
	Debug(msg string, fields ...slog.Attr)
	// Info logs an info msg with fields.
	Info(msg string, fields ...slog.Attr)
	// Warning logs a warning msg with fields.
	Warning(msg string, fields ...slog.Attr)
	// Error logs an error msg with fields.
	Error(msg string, fields ...slog.Attr)
	// ErrorIn write message like "<funcName> completes with error" with specified err and fields.
	ErrorIn(funcName string, err error, fields ...slog.Attr)
	// Fatal logs a fatal error msg with fields.
	Fatal(msg string, fields ...slog.Attr)
	// Panic logs a panic error msg with fields.
	Panic(msg string, fields ...slog.Attr)
	// With adds fields to the existing logger instance.
	With(fields ...slog.Attr)
	// End the span; function end logging.
	End()
	// ReturningParams write message like "returning params" with specified fields.
	ReturningParams(fields ...slog.Attr)
	// SIP logs a SIP message with direction and optional structured fields.
	// It formats the message for both console output (with color highlighting)
	// and OpenTelemetry spans (without colors for trace storage).
	//
	// Parameters:
	//   - direction: message direction, must be one of: SENT, RECEIVED
	//    (const: logger.SIPDirectionSent, logger.SIPDirectionReceived)
	//   - msg: raw SIP message (request or response)
	//   - fields: optional structured logging attributes (e.g., slog.String("key", "value"))
	// Behavior:
	//   - Error if direction is invalid
	//   - Error if msg is empty
	//   - Logs to console with ANSI colors
	//   - Adds event to current OpenTelemetry span without colors
	//
	// Example:
	//
	//	log.SIP(logger.SIPDirectionSent, "REGISTER sip:example.com SIP/2.0\r\n...")
	//	log.SIP(logger.SIPDirectionReceived, "SIP/2.0 200 OK", slog.String("str", "str"))
	SIP(direction string, msg string, fields ...slog.Attr)
}

type LoggerFactory interface {
	// GetLoggerFromContext start a span; function start logging.
	// Must be used in conjunction with the End() function.
	// Example:
	// ctx, logger := loggerFactory.GetLoggerFromContext(cxt, fields...)
	// defer logger.End()
	GetLoggerFromContext(ctx context.Context, fields ...slog.Attr) (context.Context, Logger)
	// GetLoggerFromContextWithSpanName start a span with a given name; function start logging.
	// Must be used in conjunction with the End() function.
	// If spanName is empty, spanName will be set to caller function name.
	// Example:
	// ctx, logger := loggerFactory.GetLoggerFromContext(cxt, "span_name", fields...)
	// defer logger.End()
	GetLoggerFromContextWithSpanName(ctx context.Context, spanName string, fields ...slog.Attr) (context.Context, Logger)
	// ForceFlush immediately exports all spans that have not yet been exported.
	ForceFlush(ctx context.Context)
}
