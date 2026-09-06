package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/guryev-vladislav/go-logger/pkg/handlers"
)

var (
	onceTraceLogger    sync.Once
	traceLoggerFactory *TraceLoggerFactory
	errTraceLogger     error
)

type TraceLoggerFactory struct {
	tracer trace.Tracer
	logger *slog.Logger
}

type traceLogger struct {
	span         trace.Span
	functionName string
	logger       *slog.Logger
	fields       []slog.Attr
}

func NewTraceLoggerFactory(
	serviceName, version, logFilePath, host, jaegerEndpoint string, minLogLevel slog.Level,
) (LoggerFactory, error) {
	onceTraceLogger.Do(func() {
		var (
			exporter *otlptrace.Exporter
			err      error
		)

		exporter, err = otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(jaegerEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			errTraceLogger = fmt.Errorf(errOTLP, err)
			return
		}

		res, err := resource.New(context.Background(),
			resource.WithAttributes(
				semconv.ServiceName(serviceName),
				semconv.ServiceVersion(version),
				semconv.HostName(host),
			),
			resource.WithFromEnv(),
			resource.WithTelemetrySDK(),
		)
		if err != nil {
			errTraceLogger = fmt.Errorf(errRes, err)
			return
		}

		tp := tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(exporter),
			tracesdk.WithResource(res),
		)

		otel.SetTracerProvider(tp)

		tracer := tp.Tracer(serviceName)

		slogFactory, err := NewSlogLoggerFactory(serviceName, version, logFilePath, host, minLogLevel)
		if err != nil {
			errTraceLogger = err
			return
		}

		slogFactoryImpl, ok := slogFactory.(*slogLoggerFactory)
		if !ok {
			errTraceLogger = ErrInvalidLoggerFactory
			return
		}

		traceLoggerFactory = &TraceLoggerFactory{
			tracer: tracer,
			logger: slogFactoryImpl.logger,
		}
	})

	if errTraceLogger != nil {
		return nil, errTraceLogger
	}

	return traceLoggerFactory, nil
}

func (f *TraceLoggerFactory) GetLoggerFromContext(
	ctx context.Context,
	fields ...slog.Attr,
) (context.Context, Logger) {
	functionName := getFunctionName()
	return f.startSpan(ctx, functionName, fields...)
}

func (f *TraceLoggerFactory) GetLoggerFromContextWithSpanName(
	ctx context.Context,
	spanName string,
	fields ...slog.Attr,
) (context.Context, Logger) {
	if spanName == emptyString {
		spanName = getFunctionName()
	}

	return f.startSpan(ctx, spanName, fields...)
}

func (f *TraceLoggerFactory) ForceFlush(ctx context.Context) {
	if tp := otel.GetTracerProvider(); tp != nil {
		if sdkProvider, ok := tp.(*tracesdk.TracerProvider); ok {
			if err := sdkProvider.ForceFlush(ctx); err != nil && f.logger != nil {
				f.logger.ErrorContext(ctx, err.Error())
			}
		}
	}
}

func (f *TraceLoggerFactory) Shutdown(ctx context.Context) error {
	if tp := otel.GetTracerProvider(); tp != nil {
		if sdkProvider, ok := tp.(*tracesdk.TracerProvider); ok {
			return sdkProvider.Shutdown(ctx)
		}
	}

	return nil
}

func (f *TraceLoggerFactory) logStartMessage(ctx context.Context, fields []slog.Attr) {
	pc, file, line, callerFuncName := getCallerPCForTraceLogger(5)
	functionName := getFunctionName()

	var msg string
	if len(fields) > 0 {
		msg = fmt.Sprintf(prefixFormat, functionName, msgStartWith)
	} else {
		msg = fmt.Sprintf(prefixFormat, functionName, msgStart)
	}

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pc)

	if callerFuncName != emptyString {
		if trimmed := handlers.TrimFuncName(callerFuncName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			source := fmt.Sprintf(sourceFormat, trimmed, line)
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

func (f *TraceLoggerFactory) startSpan(
	ctx context.Context,
	spanName string,
	fields ...slog.Attr,
) (context.Context, Logger) {
	ctx, span := f.tracer.Start(ctx, spanName)

	if len(fields) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(fields))
		for _, attr := range fields {
			attrs = append(attrs, convertSlogAttrToKeyValue(attr))
		}

		span.SetAttributes(attrs...)
	}

	msg := msgStart
	if len(fields) > 0 {
		msg = msgStartWith
	}

	span.AddEvent(msg, trace.WithAttributes(convertFieldsToAttributes(fields)...))

	f.logStartMessage(ctx, fields)

	logger := &traceLogger{
		span:         span,
		functionName: spanName,
		logger:       f.logger,
		fields:       append([]slog.Attr(nil), fields...),
	}

	return ctx, logger
}

func (t *traceLogger) CreateBackgroundCtx(context.Context) context.Context {
	newCtx := context.Background()

	if t.span != nil && t.span.SpanContext().IsValid() {
		tracer := otel.Tracer(emptyString)

		_, newSpan := tracer.Start(newCtx, t.functionName,
			trace.WithLinks(trace.Link{SpanContext: t.span.SpanContext()}))
		if newSpan != nil {
			newSpan.End()
		}
	}

	return newCtx
}

func (t *traceLogger) Debug(msg string, fields ...slog.Attr) {
	t.logToSpan(debugLevel, msg, fields)
	t.logToConsole(slog.LevelDebug, msg, fields)
}

func (t *traceLogger) Info(msg string, fields ...slog.Attr) {
	t.logToSpan(infoLevel, msg, fields)
	t.logToConsole(slog.LevelInfo, msg, fields)
}

func (t *traceLogger) Warning(msg string, fields ...slog.Attr) {
	t.logToSpan(warningLevel, msg, fields)
	t.logToConsole(slog.LevelWarn, msg, fields)
}

func (t *traceLogger) Error(msg string, fields ...slog.Attr) {
	t.logToSpan(errLevel, msg, fields)
	t.logToConsole(slog.LevelError, msg, fields)

	if t.span != nil {
		t.span.SetStatus(codes.Error, msg)
	}
}

func (t *traceLogger) ErrorIn(funcName string, err error, fields ...slog.Attr) {
	if err == nil {
		return
	}

	msg := fmt.Sprintf(msgCompletesError, funcName)
	allFields := make([]slog.Attr, 0, len(fields)+1)
	allFields = append(allFields, slog.String(fieldErr, err.Error()))
	allFields = append(allFields, fields...)

	t.Error(msg, allFields...)
}

func (t *traceLogger) Fatal(msg string, fields ...slog.Attr) {
	t.logToSpan(fatalLevel, msg, fields)
	t.logToConsole(slog.LevelError, fmt.Sprintf(msgFatal, msg), fields)

	if t.span != nil {
		t.span.SetStatus(codes.Error, msg)
		t.span.End()
	}

	os.Exit(1)
}

func (t *traceLogger) Panic(msg string, fields ...slog.Attr) {
	t.logToSpan(panicLevel, msg, fields)
	t.logToConsole(slog.LevelError, fmt.Sprintf(msgPanic, msg), fields)

	if t.span != nil {
		t.span.SetStatus(codes.Error, msg)
		t.span.End()
	}

	panic(msg)
}

func (t *traceLogger) End() {
	defer t.handlePanic()

	if t.span != nil {
		t.span.AddEvent(fmt.Sprintf(suffixEnd, t.functionName))
		t.span.End()
	}

	if t.logger != nil {
		t.logEndMessage(4)
	}
}

func (t *traceLogger) ReturningParams(fields ...slog.Attr) {
	t.logToSpan(infoLevel, msgReturning, fields)
	t.logToConsole(slog.LevelInfo, msgReturning, fields)
}

func (t *traceLogger) SIP(direction string, msg string, fields ...slog.Attr) {
	if strings.TrimSpace(msg) == emptyString {
		t.logToConsole(slog.LevelError, sipLogFailedMsg, []slog.Attr{
			slog.String(errorAttr, ErrEmptySIPMessage.Error()),
		})

		return
	}

	direction = strings.ToUpper(direction)
	if direction != SIPDirectionSent && direction != SIPDirectionReceived {
		errMsg := fmt.Sprintf("%s: %s", ErrInvalidSIPDirection.Error(), direction)
		t.logToConsole(slog.LevelError, sipLogFailedMsg, []slog.Attr{
			slog.String(errorAttr, errMsg),
			slog.String(sipFieldProvidedDir, direction),
		})

		return
	}

	timestamp := time.Now()
	coloredBlock := handlers.FormatSIPBlock(direction, msg, timestamp)

	callID := handlers.ExtractCallIDFromSIP(msg)
	method := handlers.ExtractMethodFromSIP(msg)
	statusCode, statusText := handlers.ExtractStatusFromSIP(msg)

	sipFields := t.buildSIPFields(direction, timestamp, callID, method,
		statusCode, statusText, fields)

	t.logToConsole(slog.LevelInfo, coloredBlock, sipFields)

	if t.span == nil {
		t.logToConsole(slog.LevelWarn, sipLogNotSentToSpan, nil)
		return
	}

	plainBlock := handlers.FormatSIPBlockPlain(direction, msg, timestamp)
	t.logToSpan(sipEventName, plainBlock, sipFields)
}

func (t *traceLogger) With(fields ...slog.Attr) {
	t.fields = append(t.fields, fields...)
}

func (t *traceLogger) handlePanic() {
	if err := recover(); err != nil {
		if t.span != nil {
			t.span.AddEvent(msgPanicCatched,
				trace.WithAttributes(
					attribute.String(fieldErr, fmt.Sprintf("%v", err)),
					attribute.String(fieldStack, string(debug.Stack())),
				))
			t.span.SetStatus(codes.Error, fmt.Sprintf("panic: %v", err))
		}

		if t.logger != nil {
			t.logPanicError(err)
		}

		panic(err)
	}
}

func (t *traceLogger) logPanicError(err any) {
	pc, file, line, funcName := getCallerPCForTraceLogger(4)
	rec := slog.NewRecord(time.Now(), slog.LevelError, msgPanicCatched, pc)
	rec.AddAttrs(
		slog.Any(fieldErr, err),
		slog.String(fieldStack, string(debug.Stack())),
		slog.String(fieldFunc, t.functionName),
	)

	if funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			source := fmt.Sprintf(sourceFormat, trimmed, line)
			rec.AddAttrs(slog.String(sourceAttr, source))
		}
	}

	handler := t.logger.Handler()
	if handler.Enabled(context.Background(), slog.LevelError) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (t *traceLogger) logEndMessage(skip int) {
	pc, file, line, funcName := getCallerPCForTraceLogger(skip)
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(suffixEnd, t.functionName), pc)

	if funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			source := fmt.Sprintf(sourceFormat, trimmed, line)
			rec.AddAttrs(slog.String(sourceAttr, source))
		}
	}

	handler := t.logger.Handler()
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (t *traceLogger) buildSIPFields(
	direction string, timestamp time.Time, callID, method string,
	statusCode int, statusText string, fields []slog.Attr,
) []slog.Attr {
	sipFields := make([]slog.Attr, 0, 6+len(fields))
	sipFields = append(sipFields,
		slog.String(sipFieldDirection, strings.ToLower(direction)),
		slog.String(sipFieldTimestamp, timestamp.Format(time.RFC3339)),
	)

	if callID != emptyString {
		sipFields = append(sipFields, slog.String(sipFieldCallID, callID))
	}

	if method != emptyString {
		sipFields = append(sipFields, slog.String(sipFieldMethod, method))
	}

	if statusCode > 0 {
		sipFields = append(sipFields,
			slog.Int(sipFieldStatusCode, statusCode),
			slog.String(sipFieldStatusText, statusText))
	}

	return append(sipFields, fields...)
}

func (t *traceLogger) logToSpan(level string, msg string, fields []slog.Attr) {
	if t.span == nil {
		return
	}

	attrs := convertFieldsToAttributes(t.allFields(fields))
	t.span.AddEvent(fmt.Sprintf(prefixFormat, level, msg), trace.WithAttributes(attrs...))
}

func (t *traceLogger) logToConsole(level slog.Level, msg string, fields []slog.Attr) {
	if t.logger == nil {
		return
	}

	fullMsg := fmt.Sprintf(prefixFormat, t.functionName, msg)
	pc, file, line, funcName := getCallerPCForTraceLogger(5)
	rec := slog.NewRecord(time.Now(), level, fullMsg, pc)

	fields = t.allFields(fields)
	hasFuncName, hasSource := t.checkConsoleFields(fields)
	t.addCallerInfoToRecord(rec, hasFuncName, hasSource, funcName, file, line)

	for _, attr := range fields {
		rec.AddAttrs(attr)
	}

	t.handleLogRecord(rec, level)
}

func (t *traceLogger) allFields(fields []slog.Attr) []slog.Attr {
	allFields := make([]slog.Attr, 0, len(t.fields)+len(fields))
	allFields = append(allFields, t.fields...)
	return append(allFields, fields...)
}

func (t *traceLogger) addCallerInfoToRecord(
	rec slog.Record,
	hasFuncName, hasSource bool,
	funcName, file string,
	line int,
) slog.Record {
	if !hasFuncName && funcName != emptyString {
		if trimmed := handlers.TrimFuncName(funcName); trimmed != emptyString {
			rec.AddAttrs(slog.String(funcNameAttr, trimmed))
		}
	}

	if !hasSource && file != emptyString {
		if trimmed := handlers.TrimFilePath(file); trimmed != emptyString {
			source := fmt.Sprintf(sourceFormat, trimmed, line)
			rec.AddAttrs(slog.String(sourceAttr, source))
		}
	}

	return rec
}

func (t *traceLogger) handleLogRecord(rec slog.Record, level slog.Level) {
	handler := t.logger.Handler()
	if handler.Enabled(context.Background(), level) {
		if err := handler.Handle(context.Background(), rec); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, errorLogHandler+": %v\n", err) //nolint:errcheck
		}
	}
}

func (t *traceLogger) checkConsoleFields(fields []slog.Attr) (bool, bool) {
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

func convertSlogAttrToKeyValue(attr slog.Attr) attribute.KeyValue {
	key := attribute.Key(attr.Key)
	value := attr.Value.Any()

	switch val := value.(type) {
	case string:
		return key.String(val)
	case int:
		return key.Int(val)
	case int64:
		return key.Int64(val)
	case float64:
		return key.Float64(val)
	case bool:
		return key.Bool(val)
	default:
		return key.String(fmt.Sprintf("%v", val))
	}
}

func convertFieldsToAttributes(fields []slog.Attr) []attribute.KeyValue {
	if len(fields) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(fields))
	for _, attr := range fields {
		attrs = append(attrs, convertSlogAttrToKeyValue(attr))
	}

	return attrs
}

func getCallerPCForTraceLogger(skip int) (uintptr, string, int, string) {
	var pcs [maxStackDepth]uintptr

	num := runtime.Callers(skip, pcs[:])
	if num == 0 {
		return 0, emptyString, 0, emptyString
	}

	callsite := pcs[0]
	if info, ok := globalCache.get(callsite); ok {
		return info.pc, info.file, info.line, info.funcName
	}

	frames := runtime.CallersFrames(pcs[:num])

	for {
		frame, _ := frames.Next()

		globalCache.set(callsite, callerInfo{
			pc:       frame.PC,
			file:     frame.File,
			line:     frame.Line,
			funcName: frame.Function,
		})

		return frame.PC, frame.File, frame.Line, frame.Function
	}
}
