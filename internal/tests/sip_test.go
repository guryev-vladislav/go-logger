package tests

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/guryev-vladislav/tracelog/pkg/handlers"
	logger "github.com/guryev-vladislav/tracelog/pkg/logger"
)

func newLocalTestLogger(t *testing.T) logger.Logger {
	t.Helper()
	t.Setenv(logger.EnvKeyLoggerDst, logger.Local)

	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion})
	if err != nil {
		t.Fatalf("failed to create logger factory: %v", err)
	}

	_, testLogger := factory.GetLoggerFromContext(context.Background())
	return testLogger
}

func TestSIPValidation(t *testing.T) {
	testLogger := newLocalTestLogger(t)
	sipMessage := `REGISTER sip:10.202.106.245 SIP/2.0
Via: SIP/2.0/WS p6qufb5fe1od.invalid;branch=z9hG4bKk13semp6u
Call-ID: test-call-id-123
From: <sip:5114@10.202.106.245>
To: <sip:5114@10.202.106.245>
CSeq: 1 REGISTER
Content-Length: 0`

	t.Run("ValidDirectionSent", func(t *testing.T) { testLogger.SIP("SENT", sipMessage) })
	t.Run("ValidDirectionReceived", func(t *testing.T) { testLogger.SIP("RECEIVED", sipMessage) })
	t.Run("ValidDirectionLowercase", func(t *testing.T) {
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
	testLogger := newLocalTestLogger(t)

	t.Run("SIPRequestWithCallID", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER sip:example.com SIP/2.0\nCall-ID: abc123-def456-ghi789\nContent-Length: 0")
	})
	t.Run("SIPRequestWithoutCallID", func(t *testing.T) {
		testLogger.SIP("SENT", "INVITE sip:user@example.com SIP/2.0\nContent-Length: 0")
	})
	t.Run("SIPResponseWithStatusCode", func(t *testing.T) {
		testLogger.SIP("RECEIVED", "SIP/2.0 200 OK\nCall-ID: test-call-456\nContent-Length: 0")
	})
	t.Run("SIPResponseWithMultipleStatusCodes", func(t *testing.T) {
		responses := []string{"100 Trying", "180 Ringing", "200 OK", "302 Moved Temporarily", "400 Bad Request", "401 Unauthorized", "403 Forbidden", "404 Not Found", "408 Request Timeout", "480 Temporarily Unavailable", "486 Busy Here", "500 Server Internal Error", "503 Service Unavailable", "600 Busy Everywhere"}
		for _, response := range responses {
			testLogger.SIP("RECEIVED", "SIP/2.0 "+response)
		}
	})
	t.Run("SIPRequestWithDifferentMethods", func(t *testing.T) {
		methods := []string{"INVITE", "ACK", "BYE", "CANCEL", "OPTIONS", "SUBSCRIBE", "NOTIFY", "PUBLISH", "MESSAGE"}
		for _, method := range methods {
			testLogger.SIP("SENT", method+" sip:user@example.com SIP/2.0\nCall-ID: test-"+method+"\nContent-Length: 0")
		}
	})
}

func TestExtractCallIDFromSIPIsCaseInsensitive(t *testing.T) {
	message := "INVITE sip:example.com SIP/2.0\ncall-id: abc-123\n"

	if got := handlers.ExtractCallIDFromSIP(message); got != "abc-123" {
		t.Fatalf("ExtractCallIDFromSIP() = %q, want %q", got, "abc-123")
	}
}

func TestSIPWithAdditionalFields(t *testing.T) {
	testLogger := newLocalTestLogger(t)
	sipMessage := "REGISTER sip:example.com SIP/2.0\nCall-ID: test-call-789\nContent-Length: 0"

	t.Run("WithSingleField", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage, slog.String("transaction_id", "txn-123"))
	})
	t.Run("WithMultipleFields", func(t *testing.T) {
		testLogger.SIP("RECEIVED", sipMessage, slog.String("proxy", "edge-01"), slog.Int("retry_count", 3), slog.Duration("latency", 150*time.Millisecond), slog.Bool("cached", false))
	})
	t.Run("WithAllFieldTypes", func(t *testing.T) {
		testLogger.SIP("SENT", sipMessage, slog.String("string_field", "value"), slog.Int("int_field", 42), slog.Float64("float_field", 3.14), slog.Bool("bool_field", true), slog.Any("any_field", map[string]string{"key": "value"}))
	})
}

func TestSIPComplexMessages(t *testing.T) {
	testLogger := newLocalTestLogger(t)

	t.Run("FullRegisterMessage", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER SIP/2.0\nVia: SIP/2.0/WS p6qufb5fe1od.invalid;branch=z9hG4bKk13semp6u\nMax-Forwards: 69\nTo: <sip>\nFrom: <sip:>;tag=c4vgt5vjha\nCall-ID: glpn6hr5nxbh9rsh6xugi6\nCSeq: 1 REGISTER\nContent-Length: 0")
	})
	t.Run("FullUnauthorizedResponse", func(t *testing.T) {
		testLogger.SIP("RECEIVED", "SIP/2.0 401 Unauthorized\nCall-ID: glpn6hr5nxbh9rsh6xugi6\nWWW-Authenticate: Digest realm=\"SIP-REGISTRAR\"\nContent-Length: 0")
	})
	t.Run("InviteWithSDP", func(t *testing.T) {
		testLogger.SIP("SENT", "INVITE sip:user@example.com SIP/2.0\nCall-ID: asd88asd77a@1.2.3.4\nContent-Type: application/sdp\nContent-Length: 147\n\nv=0\no=caller 2890844526 2890844526 IN IP4 client.example.com")
	})
	t.Run("MessageWithAuthHeaders", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER sip:example.com SIP/2.0\nCall-ID: asd88asd77a@1.2.3.4\nAuthorization: Digest username=\"user\"\nContent-Length: 0")
	})
}

func TestSIPEdgeCases(t *testing.T) {
	testLogger := newLocalTestLogger(t)

	t.Run("VeryLongMessage", func(t *testing.T) {
		longHeader := strings.Repeat("X-Extra-Header: "+strings.Repeat("value", 100)+"\n", 10)
		testLogger.SIP("SENT", "REGISTER sip:example.com SIP/2.0\n"+longHeader+"Content-Length: 0")
	})
	t.Run("MessageWithSpecialCharacters", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER sip:example.com SIP/2.0\nCall-ID: test-id-123\nContact: <sip:user@[2001:db8::1]:5060>\nContent-Length: 0")
	})
	t.Run("MessageWithoutNewlines", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER sip:example.com SIP/2.0 Call-ID: test Content-Length: 0")
	})
	t.Run("MessageWithMultipleSpaces", func(t *testing.T) {
		testLogger.SIP("SENT", "REGISTER   sip:example.com   SIP/2.0\nCall-ID:   test-123\nContent-Length:   0")
	})
}

func TestSIPConcurrent(t *testing.T) {
	testLogger := newLocalTestLogger(t)
	sipMessage := "REGISTER sip:example.com SIP/2.0\nCall-ID: concurrent-test\nContent-Length: 0"

	const iterations = 100
	done := make(chan struct{}, iterations)
	for i := range iterations {
		go func(index int) {
			testLogger.SIP("SENT", sipMessage, slog.Int("goroutine", index))
			done <- struct{}{}
		}(i)
	}
	for range iterations {
		<-done
	}
}

func TestSIPWithSpan(t *testing.T) {
	t.Setenv(logger.EnvKeyLoggerDst, "")
	factory, err := logger.New(logger.LoggerConfig{ServiceName: serviceName, Version: serviceVersion, JaegerEndpoint: "localhost:4317"})
	if err != nil {
		t.Skipf("Skipping trace test: %v", err)
	}

	ctx, testLogger := factory.GetLoggerFromContext(context.Background())
	sipMessage := "REGISTER sip:example.com SIP/2.0\nCall-ID: span-test-123\nContent-Length: 0"

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
		_, nilSpanLogger := factory.GetLoggerFromContext(context.Background())
		nilSpanLogger.SIP("SENT", sipMessage)
	})

}
