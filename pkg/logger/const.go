package logger

const (
	EnvKeyLoggerDst      = "LOGGER_DST"
	EnvKeyStandDomain    = "STAND_DOMAIN"
	EnvKeyLoggerMinLevel = "LOGGER_MIN_LEVEL"
	EnvKeyLogFile        = "LOG_FILE"
)

const (
	msgFatal          = "FATAL: %s"
	msgPanic          = "PANIC: %s"
	msgStart          = "start"
	msgStartWith      = "start with params"
	msgPanicCatched   = "the panic was catched"
	msgCompletesError = "%s completes with error"
	msgReturning      = "returning params"

	debugLevel   = "DEBUG"
	infoLevel    = "INFO"
	warningLevel = "WARN"
	errLevel     = "ERROR"
	fatalLevel   = "FATAL"
	panicLevel   = "PANIC"
)

const (
	prefixFormat = "%s: %s"
	suffixEnd    = "%s: end"
)

const (
	fieldErr   = "error"
	fieldFunc  = "func"
	fieldStack = "stacktrace"
)

const (
	errOTLP         = "failed to create OTLP exporter: %w"
	errRes          = "failed to create resource: %w"
	errorLogHandler = "log handler error"
)

const (
	callerSkipDepth = 2
	emptyString     = ""

	logFileMB       = 100
	logFileBackups  = 3
	logFileAge      = 7
	logFileCompress = true

	errorAttr    = "error"
	funcNameAttr = "func_name"
	serviceAttr  = "service"
	versionAttr  = "version"
	standAttr    = "stand"
	sourceAttr   = "source"
	Local        = "local"

	sourceFormat = "%s:%d"
)

const (
	SIPDirectionSent     = "SENT"
	SIPDirectionReceived = "RECEIVED"
	sipLogFailedMsg      = "SIP logging failed"
	sipLogNotSentToSpan  = "SIP log not sent to span: span is nil"
	sipEventName         = "SIP"
	sipFieldDirection    = "sip.direction"
	sipFieldTimestamp    = "sip.timestamp"
	sipFieldCallID       = "sip.call_id"
	sipFieldMethod       = "sip.method"
	sipFieldStatusCode   = "sip.status_code"
	sipFieldStatusText   = "sip.status_text"
	sipFieldProvidedDir  = "provided_direction"
)
