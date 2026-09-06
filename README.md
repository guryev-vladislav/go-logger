# Go Logger

`go-logger` is a structured logging package for Go applications. It combines local `log/slog` logging with OpenTelemetry tracing and supports SIP message formatting for readable console output and trace events.

## Features

- structured logs with `slog.Attr` fields;
- local console output with level-based colors;
- OpenTelemetry span events;
- JSON file logging with rotation;
- SIP request and response formatting;
- caller, service, version, and stand metadata;
- configurable log levels and destinations.

## Requirements

- Go 1.27 or later;
- an OpenTelemetry-compatible OTLP endpoint for trace mode.

## Installation

Install the package from GitHub:

```text
go get github.com/guryev-vladislav/go-logger/pkg/logger
```

## Configuration

| Variable | Description | Default |
| --- | --- | --- |
| `LOGGER_DST` | Set to `local` to use console logging. Any other value enables OpenTelemetry. | OpenTelemetry |
| `LOGGER_MIN_LEVEL` | Minimum level: `DEBUG`, `INFO`, `WARN`, or `ERROR`. | `INFO` |
| `LOG_FILE` | Path to the JSON log file. | Empty |
| `STAND_DOMAIN` | Host or environment identifier. | Empty |

`LoggerConfig.LogFilePath` takes priority over `LOG_FILE`.

## Example

See [example.go](example.go) for a complete usage example with local logging, SIP messages, structured fields, and graceful shutdown.

## Development

Run the test suite and static checks before submitting changes:

```text
go test -race ./...
go vet ./...
```
