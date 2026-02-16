# Commons Library

A collection of common Go utilities for building production applications.

## Features

### Logger

A structured logging wrapper around [go.uber.org/zap](https://github.com/uber-go/zap) that provides:

- JSON formatted output with ISO8601 timestamps
- Configurable log levels (debug, info, warn, error, fatal)
- Environment variable configuration via `LOG_LEVEL`
- Runtime log level changes
- Caller information in logs

## Installation

```bash
go get github.com/jbjoret/commons
```

## Usage

### Logger Package

#### Basic Usage

```go
package main

import (
    "github.com/jbjoret/commons/logger"
    "go.uber.org/zap"
)

func main() {
    // Optional: Initialize from environment (reads LOG_LEVEL env var)
    logger.Init()

    // Log at different levels
    logger.Debug("Debug information", zap.Int("count", 42))
    logger.Info("Application started", zap.String("version", "1.0"))
    logger.Warn("Warning message", zap.String("component", "api"))
    logger.Error("Error occurred", zap.Error(err))

    // Fatal logs and exits
    logger.Fatal("Critical error", zap.String("reason", "startup failed"))
}
```

#### Setting Log Level

**From Environment Variable:**

```bash
# Set before running your application
export LOG_LEVEL=debug
go run main.go
```

**Programmatically:**

```go
// Change log level at runtime
if err := logger.SetLogLevel("debug"); err != nil {
    logger.Error("Failed to set log level", zap.Error(err))
}

// Get current log level
level := logger.GetLogLevel()
logger.Info("Current log level", zap.String("level", level))
```

#### Supported Log Levels

- `debug` - Detailed information for diagnosing problems
- `info` - General informational messages (default)
- `warn` or `warning` - Warning messages
- `error` - Error messages
- `dpanic` - Development panic (panics in development, logs in production)
- `panic` - Logs and then panics
- `fatal` - Logs and then calls os.Exit(1)

#### Example Output

```json
{
  "level": "info",
  "timestamp": "2026-01-16T11:37:53.965+0100",
  "caller": "main.go:15",
  "msg": "Application started",
  "version": "1.0"
}
```

## Development

### Building

```bash
make target
```

### Running Tests

```bash
make test
```

### Formatting Code

```bash
make fmt
```

### Updating Dependencies

```bash
make update
```

## Requirements

- Go 1.26 or higher
- Dependencies are managed via Go modules

## License

See LICENSE file for details.
