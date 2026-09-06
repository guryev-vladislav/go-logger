package handlers

import "time"

const (
	timeFormat   = time.RFC3339
	attrFormat   = "%s: %v"
	msgFormat    = "%s {%s}"
	outputFormat = "%s - %s - %s - %s:%d - %s: %s\n"

	pathSeparator     = "/"
	funcSeparator     = '.'
	pathSlash         = '/'
	defaultMaxBackups = 2
	emptyString       = ""

	funcNameAttr = "func_name"
	sourceAttr   = "source"
	serviceAttr  = "service"
)

const (
	colorGreen   = "\033[32m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
	colorGray    = "\033[90m"
	colorBlue    = "\033[34m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBold    = "\033[1m"
)

const (
	separatorLine = "------------------------"
	newline       = "\n"
	space         = " "
)

const (
	headerPrefixColored   = "\n\n%s--- [%s %s] %s ---%s\n"
	headerPrefixPlain     = "\n\n--- [%s] %s ---\n\n%s\n%s\n"
	sipVersion            = "SIP/2.0"
	callIDPrefix          = "Call-ID:"
	branchParam           = "branch="
	authorizationHeader   = "Authorization:"
	wwwAuthenticateHeader = "WWW-Authenticate:"
)

const (
	symbolSent     = "➤"
	symbolReceived = "◀"
)

const (
	headerTextSent     = "SIP MESSAGE SENT"
	headerTextReceived = "SIP MESSAGE RECEIVED"
)

var (
	sipMethods = []string{
		"INVITE", "REGISTER", "BYE", "ACK", "CANCEL",
		"OPTIONS", "SUBSCRIBE", "NOTIFY", "MESSAGE", "PUBLISH",
	}

	sipHeaders = []string{
		"Via:", "From:", "To:", "Call-ID:", "CSeq:", "Contact:",
		"Content-Type:", "Content-Length:", "Max-Forwards:", "Expires:",
		"Allow:", "Supported:", "User-Agent:", "Server:", "Record-Route:",
		"WWW-Authenticate:", "Authorization:", "Proxy-Authorization:",
	}

	directionSent = "SENT"
)
