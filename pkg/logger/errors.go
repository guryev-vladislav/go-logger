package logger

import "errors"

var ErrInvalidLoggerFactory = errors.New("failed to cast slog logger factory")
var ErrCantCreateCtxWithSpan = errors.New("can't create ctx with a span")
var ErrInvalidSIPDirection = errors.New("invalid SIP direction: must be SENT or RECEIVED")
var ErrEmptySIPMessage = errors.New("empty SIP message")
