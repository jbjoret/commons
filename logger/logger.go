// Package logger provides a structured logging wrapper around go.uber.org/zap.
//
// The logger is initialized with sensible defaults (JSON output, ISO8601 timestamps,
// Info level) and can be configured via environment variables or programmatically.
//
// Usage:
//
//	import "bitbucket.org/jbjoret/commons/logger"
//
//	// Initialize from environment (optional, reads LOG_LEVEL env var)
//	logger.Init()
//
//	// Log messages
//	logger.Info("Application started", zap.String("version", "1.0"))
//	logger.Debug("Debug information", zap.Int("count", 42))
//	logger.Error("Something went wrong", zap.Error(err))
//
//	// Change log level programmatically
//	logger.SetLogLevel("debug")
//
//	// Get current log level
//	level := logger.GetLogLevel()
package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log         *zap.Logger
	atomicLevel zap.AtomicLevel
)

func init() {
	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.InfoLevel) // Default to Info level

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	)

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Init initializes the logger's level from the environment.
// It should be called after loading the .env file.
func Init() {
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		if err := SetLogLevel(levelStr); err != nil {
			Warn("Invalid LOG_LEVEL provided, using default.", zap.String("value", levelStr), zap.Error(err))
		} else {
			Info("Log level set from environment.", zap.String("level", levelStr))
		}
	} else {
		Info("Log level is not set, using default 'info'.")
	}
}

// Debug prints a message at debug level.
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Info prints a message at info level.
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn prints a message at warn level.
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error prints a message at error level.
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal prints a message at fatal level, then calls os.Exit(1).
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

// SetLogLevel sets the log level based on a string.
func SetLogLevel(level string) error {
	var newLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		newLevel = zapcore.DebugLevel
	case "info":
		newLevel = zapcore.InfoLevel
	case "warn", "warning":
		newLevel = zapcore.WarnLevel
	case "error":
		newLevel = zapcore.ErrorLevel
	case "dpanic":
		newLevel = zapcore.DPanicLevel
	case "panic":
		newLevel = zapcore.PanicLevel
	case "fatal":
		newLevel = zapcore.FatalLevel
	default:
		if err := newLevel.Set(level); err != nil {
			return fmt.Errorf("invalid log level string: %s", level)
		}
	}
	atomicLevel.SetLevel(newLevel)
	return nil
}

// GetLogLevel returns the log level
func GetLogLevel() string {
	return atomicLevel.String()
}
