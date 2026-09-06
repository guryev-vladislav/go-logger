//nolint:unused
package tests

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	logger "gitlab.satel.org/rtuc-forks/aqa/logger.git/pkg/logger"
)

var ErrTestError = errors.New("test error")

const (
	serviceVersion = "1.0.0"
	serviceName    = "test-service"
)

func TestSlogLogger(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

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
		t.Log("done")
	})

	t.Run("Info", func(t *testing.T) {
		testLogger.Info("info message")
		testLogger.Info("info message with fields", slog.String("key", "value"))
		t.Log("done")
	})

	t.Run("Warning", func(t *testing.T) {
		testLogger.Warning("warning message")
		testLogger.Warning("warning message with fields", slog.String("key", "value"))
		t.Log("done")
	})

	t.Run("Error", func(t *testing.T) {
		testLogger.Error("error message")
		testLogger.Error("error message with fields", slog.String("key", "value"))
		t.Log("done")
	})

	t.Run("ErrorIn", func(t *testing.T) {
		testLogger.ErrorIn("testFunc", ErrTestError)
		testLogger.ErrorIn("testFunc", ErrTestError, slog.String("key", "value"))
		t.Log("done")
	})

	t.Run("ReturningParams", func(t *testing.T) {
		testLogger.ReturningParams()
		testLogger.ReturningParams(slog.String("key", "value"))
		t.Log("done")
	})

	t.Run("End", func(t *testing.T) {
		testLogger.End()
		t.Log("done")
	})
}

func TestSlogLoggerFactory(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	t.Run("GetLoggerFromContext", func(t *testing.T) {
		ctx := context.Background()

		newCtx, testLogger := factory.GetLoggerFromContext(ctx)
		if newCtx == nil {
			t.Error("expected non-nil context")
		}

		if testLogger == nil {
			t.Error("expected non-nil logger")
		}
	})

	t.Run("GetLoggerFromContextWithFields", func(t *testing.T) {
		ctx := context.Background()

		newCtx, testLogger := factory.GetLoggerFromContext(ctx, slog.String("param", "value"))
		if newCtx == nil {
			t.Error("expected non-nil context")
		}

		if testLogger == nil {
			t.Error("expected non-nil logger")
		}
	})

	t.Run("GetLoggerFromContextWithSpanName", func(t *testing.T) {
		ctx := context.Background()

		newCtx, testLogger := factory.GetLoggerFromContextWithSpanName(ctx, "customSpan")
		if newCtx == nil {
			t.Error("expected non-nil context")
		}

		if testLogger == nil {
			t.Error("expected non-nil logger")
		}
	})

	t.Run("GetLoggerFromContextWithEmptySpanName", func(t *testing.T) {
		ctx := context.Background()

		newCtx, testLogger := factory.GetLoggerFromContextWithSpanName(ctx, "")
		if newCtx == nil {
			t.Error("expected non-nil context")
		}

		if testLogger == nil {
			t.Error("expected non-nil logger")
		}
	})

	t.Run("ForceFlush", func(t *testing.T) {
		factory.ForceFlush(context.Background())
		t.Log()
	})
}

func TestSlogLoggerPanic(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but did not occur")
		}
	}()

	testLogger.Panic("test panic")
}

func TestSlogLoggerPanicWithFields(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic but did not occur")
		}
	}()

	testLogger.Panic("test panic with fields", slog.String("key", "value"))
}

func TestSlogLoggerEndWithPanic(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

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

func TestSIPValidation(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	sipMessage := `REGISTER sip:10.202.106.245 SIP/2.0
Via: SIP/2.0/WS p6qufb5fe1od.invalid;branch=z9hG4bKk13semp6u
Call-ID: test-call-id-123
From: <sip:5114@10.202.106.245>
To: <sip:5114@10.202.106.245>
CSeq: 1 REGISTER
Content-Length: 0`

	t.Run("ValidDirectionSent", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage)
	})

	t.Run("ValidDirectionReceived", func(_ *testing.T) {
		testLogger.SIP("RECEIVED", sipMessage)
	})

	t.Run("ValidDirectionLowercase", func(_ *testing.T) {
		testLogger.SIP("sent", sipMessage)
		testLogger.SIP("received", sipMessage)
	})

	t.Run("EmptyMessage", func(t *testing.T) {
		testLogger.SIP("SENT", "")
		testLogger.SIP("SENT", "   ")
		testLogger.SIP("SENT", "\n\n\t")
	})

	t.Run("EmptyMessageWithFields", func(t *testing.T) {
		testLogger.SIP("SENT", "", slog.String("extra", "value"))
	})

	t.Run("InvalidDirection", func(t *testing.T) {
		testLogger.SIP("INVALID", sipMessage)
		testLogger.SIP("", sipMessage)
		testLogger.SIP("SEND", sipMessage)
	})

	t.Run("InvalidDirectionWithFields", func(t *testing.T) {
		testLogger.SIP("WRONG", sipMessage, slog.String("key", "value"))
	})
}

func TestSIPMetadataExtraction(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	t.Run("SIPRequestWithCallID", func(t *testing.T) {
		request := `REGISTER sip:example.com SIP/2.0
Call-ID: abc123-def456-ghi789
From: <sip:user@example.com>
To: <sip:user@example.com>
CSeq: 1 REGISTER
Content-Length: 0`

		testLogger.SIP("SENT", request)
	})

	t.Run("SIPRequestWithoutCallID", func(t *testing.T) {
		request := `INVITE sip:user@example.com SIP/2.0
From: <sip:user@example.com>
To: <sip:user@example.com>
CSeq: 1 INVITE
Content-Length: 0`

		testLogger.SIP("SENT", request)
	})

	t.Run("SIPResponseWithStatusCode", func(t *testing.T) {
		response := `SIP/2.0 200 OK
Via: SIP/2.0/UDP client.example.com
From: <sip:user@example.com>
To: <sip:user@example.com>
Call-ID: test-call-456
CSeq: 1 INVITE
Content-Length: 0`

		testLogger.SIP("RECEIVED", response)
	})

	t.Run("SIPResponseWithMultipleStatusCodes", func(t *testing.T) {
		responses := []string{
			"SIP/2.0 100 Trying",
			"SIP/2.0 180 Ringing",
			"SIP/2.0 200 OK",
			"SIP/2.0 302 Moved Temporarily",
			"SIP/2.0 400 Bad Request",
			"SIP/2.0 401 Unauthorized",
			"SIP/2.0 403 Forbidden",
			"SIP/2.0 404 Not Found",
			"SIP/2.0 408 Request Timeout",
			"SIP/2.0 480 Temporarily Unavailable",
			"SIP/2.0 486 Busy Here",
			"SIP/2.0 500 Server Internal Error",
			"SIP/2.0 503 Service Unavailable",
			"SIP/2.0 600 Busy Everywhere",
		}

		for _, response := range responses {
			testLogger.SIP("RECEIVED", response)
		}
	})

	t.Run("SIPRequestWithDifferentMethods", func(t *testing.T) {
		methods := []string{"INVITE", "ACK", "BYE", "CANCEL", "OPTIONS", "SUBSCRIBE", "NOTIFY", "PUBLISH", "MESSAGE"}

		for _, method := range methods {
			request := method + " sip:user@example.com SIP/2.0\nCall-ID: test-" + method + "\nContent-Length: 0"
			testLogger.SIP("SENT", request)
		}
	})
}

func TestSIPWithAdditionalFields(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	sipMessage := `REGISTER sip:example.com SIP/2.0
Call-ID: test-call-789
Content-Length: 0`

	t.Run("WithSingleField", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage, slog.String("transaction_id", "txn-123"))
	})

	t.Run("WithMultipleFields", func(t *testing.T) {
		testLogger.SIP("RECEIVED", sipMessage,
			slog.String("proxy", "edge-01"),
			slog.Int("retry_count", 3),
			slog.Duration("latency", 150*time.Millisecond),
			slog.Bool("cached", false))
	})

	t.Run("WithAllFieldTypes", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage,
			slog.String("string_field", "value"),
			slog.Int("int_field", 42),
			slog.Float64("float_field", 3.14),
			slog.Bool("bool_field", true),
			slog.Any("any_field", map[string]string{"key": "value"}))
	})
}

func TestSIPComplexMessages(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	t.Run("FullRegisterMessage", func(t *testing.T) {
		registerMsg := `REGISTER SIP/2.0
Via: SIP/2.0/WS p6qufb5fe1od.invalid;branch=z9hG4bKk13semp6u
Max-Forwards: 69
To: <sip>
From: <sip:>;tag=c4vgt5vjha
Call-ID: glpn6hr5nxbh9rsh6xugi6
CSeq: 1 REGISTER
Contact: <sip:va90cw3y@p6qufb5fe1od.invalid;
Allow: INVITE,ACK,CANCEL,BYE,UPDATE,MESSAGE,OPTIONS,REFER,INFO,NOTIFY,SUBSCRIBE
Supported: path,gruu,outbound
Content-Length: 0`

		testLogger.SIP("SENT", registerMsg)
	})

	t.Run("FullUnauthorizedResponse", func(t *testing.T) {
		responseMsg := `SIP/2.0 401 Unauthorized
Via: SIP/2.0/WS p6qufb5fe1od.invalid:5060;branch=z9hG4bKk13semp6u;received=127.0.0.1;xrport=50856
From: <sip:;tag=c4vgt5vjha
To: <sip:;tag=b89cbf843f2011f1a49e005056011fa5
Call-ID: glpn6hr5nxbh9rsh6xugi6
CSeq: 1 REGISTER
WWW-Authenticate: Digest realm="SIP-REGISTRAR", nonce="b89cbc823f2011f1a49e005056011fa5"
Server: TS-v4.7.8-06
Content-Length: 0`

		testLogger.SIP("RECEIVED", responseMsg)
	})

	t.Run("InviteWithSDP", func(t *testing.T) {
		inviteWithSDP := `INVITE sip:user@example.com SIP/2.0
Via: SIP/2.0/UDP client.example.com:5060;branch=z9hG4bK74bf9
Max-Forwards: 70
From: <sip:caller@example.com>;tag=9fxced76sl
To: <sip:user@example.com>
Call-ID: asd88asd77a@1.2.3.4
CSeq: 1 INVITE
Contact: <sip:caller@client.example.com:5060>
Content-Type: application/sdp
Content-Length: 147

v=0
o=caller 2890844526 2890844526 IN IP4 client.example.com
s=Session SDP
c=IN IP4 client.example.com
t=0 0
m=audio 49170 RTP/AVP 0
a=rtpmap:0 PCMU/8000`

		testLogger.SIP("SENT", inviteWithSDP)
	})

	t.Run("MessageWithAuthHeaders", func(t *testing.T) {
		authMsg := `REGISTER sip:example.com SIP/2.0
Via: SIP/2.0/UDP client.example.com:5060;branch=z9hG4bK74bf9
From: <sip:user@example.com>;tag=9fxced76sl
To: <sip:user@example.com>
Call-ID: asd88asd77a@1.2.3.4
CSeq: 2 REGISTER
Authorization: Digest username="user"
Content-Length: 0`

		testLogger.SIP("SENT", authMsg)
	})
}

func TestSIPEdgeCases(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	t.Run("VeryLongMessage", func(t *testing.T) {
		longHeader := strings.Repeat("X-Extra-Header: "+strings.Repeat("value", 100)+"\n", 10)
		longMsg := "REGISTER sip:example.com SIP/2.0\n" + longHeader + "Content-Length: 0"

		testLogger.SIP("SENT", longMsg)
	})

	t.Run("MessageWithSpecialCharacters", func(t *testing.T) {
		specialMsg := `REGISTER sip:example.com SIP/2.0
From: <sip:user@example.com>;tag=abc123!@#$%^&*()
To: <sip:user@example.com>
Call-ID: test-id-123
Contact: <sip:user@[2001:db8::1]:5060>
Content-Length: 0`

		testLogger.SIP("SENT", specialMsg)
	})

	t.Run("MessageWithoutNewlines", func(t *testing.T) {
		oneLineMsg := "REGISTER sip:example.com SIP/2.0 Call-ID: test Content-Length: 0"
		testLogger.SIP("SENT", oneLineMsg)
	})

	t.Run("MessageWithMultipleSpaces", func(t *testing.T) {
		multipleSpacesMsg := "REGISTER   sip:example.com   SIP/2.0\nCall-ID:   test-123\nContent-Length:   0"
		testLogger.SIP("SENT", multipleSpacesMsg)
	})
}

func TestSIPConcurrent(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	cfg := logger.LoggerConfig{
		ServiceName: serviceName,
		Version:     serviceVersion,
		LogFilePath: "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	ctx := context.Background()
	_, testLogger := factory.GetLoggerFromContext(ctx)

	sipMessage := `REGISTER sip:example.com SIP/2.0
Call-ID: concurrent-test
Content-Length: 0`

	t.Run("ConcurrentSIPCalls", func(t *testing.T) {
		done := make(chan bool)
		iterations := 100

		for i := range iterations {
			go func(idx int) {
				testLogger.SIP("SENT", sipMessage, slog.Int("goroutine", idx))

				done <- true
			}(i)
		}

		for range iterations {
			<-done
		}
	})
}

func TestSIPWithSpan(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, "")

	cfg := logger.LoggerConfig{
		ServiceName:    "test-service",
		Version:        serviceVersion,
		JaegerEndpoint: "localhost:4317",
		LogFilePath:    "",
	}

	factory, err := logger.New(cfg)
	if err != nil {
		t.Skipf("Skipping trace test: %v", err)
	}

	ctx := context.Background()
	ctx, testLogger := factory.GetLoggerFromContext(ctx)

	sipMessage := `REGISTER sip:example.com SIP/2.0
Call-ID: span-test-123
Content-Length: 0`

	t.Run("SIPWithActiveSpan", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage, slog.String("test_id", t.Name()))
		testLogger.End()
	})

	t.Run("MultipleSIPInSameSpan", func(t *testing.T) {
		_, testLogger2 := factory.GetLoggerFromContext(ctx)

		testLogger2.SIP("SENT", "INVITE sip:user@example.com SIP/2.0\nCall-ID: inv-001\nContent-Length: 0")
		testLogger2.SIP("RECEIVED", "SIP/2.0 100 Trying\nCall-ID: inv-001\nContent-Length: 0")
		testLogger2.SIP("RECEIVED", "SIP/2.0 180 Ringing\nCall-ID: inv-001\nContent-Length: 0")
		testLogger2.SIP("RECEIVED", "SIP/2.0 200 OK\nCall-ID: inv-001\nContent-Length: 0")
		testLogger2.SIP("SENT", "ACK sip:user@example.com SIP/2.0\nCall-ID: inv-001\nContent-Length: 0")

		testLogger2.End()
	})

	t.Run("SIPWithNilSpan", func(t *testing.T) {
		backgroundCtx := context.Background()
		_, nilSpanLogger := factory.GetLoggerFromContext(backgroundCtx)

		nilSpanLogger.SIP("SENT", sipMessage)
	})
}
