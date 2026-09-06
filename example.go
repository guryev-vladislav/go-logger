package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/guryev-vladislav/go-logger/pkg/logger"
)

const (
	serviceName    = "example-service"
	serviceVersion = "1.0.0"
	jaegerEndpoint = "localhost:4317"
	flushTimeout   = 5 * time.Second
)

var (
	envs = map[string]string{
		logger.EnvKeyLoggerDst:      "local",
		logger.EnvKeyStandDomain:    "",
		logger.EnvKeyLoggerMinLevel: "INFO",
		logger.EnvKeyLogFile:        "log.json",
	}
)

func main() {
	ctx := context.Background()

	if err := setupEnv(); err != nil {
		log.Fatalf("failed to setup environment: %v", err)
	}

	cfg := logger.LoggerConfig{
		ServiceName:    serviceName,
		Version:        serviceVersion,
		JaegerEndpoint: jaegerEndpoint,
		LogFilePath:    envs[logger.EnvKeyLogFile],
	}

	logFactory, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer shutdownLogger(logFactory)

	runApplication(ctx, logFactory)
}

func setupEnv() error {
	for key, val := range envs {
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

func shutdownLogger(logFactory logger.LoggerFactory) {
	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()

	logFactory.ForceFlush(ctx)

	if closer, ok := logFactory.(interface{ Shutdown(context.Context) error }); ok {
		if err := closer.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown logger: %v", err)
		}
	}
}

func runApplication(ctx context.Context, logFactory logger.LoggerFactory) {
	ctx, log := logFactory.GetLoggerFromContext(ctx)
	defer log.End()

	log.Info("Service started")

	logSIPExamples(log)
	processRequest(ctx, logFactory)
}

func logSIPExamples(log logger.Logger) {
	registerMsg := `REGISTER sip:example.com SIP/2.0
Via: SIP/2.0/UDP client.example.com:5060;branch=z9hG4bK74bf9
Max-Forwards: 70
From: <sip:user@example.com>;tag=9fxced76sl
To: <sip:user@example.com>
Call-ID: asd88asd77a
CSeq: 1 REGISTER
Contact: <sip:user@client.example.com:5060>
Expires: 3600
Content-Length: 0`

	log.SIP(logger.SIPDirectionSent, registerMsg)

	responseMsg := `SIP/2.0 401 Unauthorized
Via: SIP/2.0/UDP client.example.com:5060;branch=z9hG4bK74bf9;received=1.2.3.4
From: <sip:user@example.com>;tag=9fxced76sl
To: <sip:user@example.com>;tag=server123
Call-ID: asd88asd77a
CSeq: 1 REGISTER
WWW-Authenticate: Digest realm="example.com", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093"
Server:
Content-Length: 0`

	log.SIP(logger.SIPDirectionReceived, responseMsg)
}

func processRequest(ctx context.Context, logFactory logger.LoggerFactory) {
	_, log := logFactory.GetLoggerFromContext(ctx)
	defer log.End()

	log.Info("Starting request processing")

	log.Debug("Request parameters",
		slog.String("user_id", "123"),
		slog.Int("attempt", 1),
	)

	if err := someOperation(); err != nil {
		log.Error("Error during processing",
			slog.String("error", err.Error()),
		)
		log.ErrorIn("processRequest", err)

		log.Warning("processRequest")

		return
	}

	log.ReturningParams(slog.String("result", "success"))
	log.Info("Request processed successfully")
}

var errTestOperation = errors.New("test error")

func someOperation() error {
	return errTestOperation
}
