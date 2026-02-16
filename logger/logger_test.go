package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectError bool
		expectedLvl zapcore.Level
	}{
		{"debug level", "debug", false, zapcore.DebugLevel},
		{"info level", "info", false, zapcore.InfoLevel},
		{"warn level", "warn", false, zapcore.WarnLevel},
		{"warning level", "warning", false, zapcore.WarnLevel},
		{"error level", "error", false, zapcore.ErrorLevel},
		{"dpanic level", "dpanic", false, zapcore.DPanicLevel},
		{"panic level", "panic", false, zapcore.PanicLevel},
		{"fatal level", "fatal", false, zapcore.FatalLevel},
		{"case insensitive", "DEBUG", false, zapcore.DebugLevel},
		{"invalid level", "invalid", true, zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to Info level before each test
			atomicLevel.SetLevel(zapcore.InfoLevel)

			err := SetLogLevel(tt.level)

			if tt.expectError && err == nil {
				t.Errorf("expected error for level %q, got nil", tt.level)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for level %q: %v", tt.level, err)
			}
			if atomicLevel.Level() != tt.expectedLvl {
				t.Errorf("expected level %v, got %v", tt.expectedLvl, atomicLevel.Level())
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	// Set to a known level
	atomicLevel.SetLevel(zapcore.DebugLevel)
	level := GetLogLevel()

	if !strings.Contains(strings.ToLower(level), "debug") {
		t.Errorf("expected level string to contain 'debug', got %q", level)
	}

	// Test another level
	atomicLevel.SetLevel(zapcore.ErrorLevel)
	level = GetLogLevel()

	if !strings.Contains(strings.ToLower(level), "error") {
		t.Errorf("expected level string to contain 'error', got %q", level)
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected zapcore.Level
	}{
		{"env debug", "debug", zapcore.DebugLevel},
		{"env info", "info", zapcore.InfoLevel},
		{"env warn", "warn", zapcore.WarnLevel},
		{"env empty", "", zapcore.InfoLevel}, // Should keep default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to Info level
			atomicLevel.SetLevel(zapcore.InfoLevel)

			if tt.envValue != "" {
				t.Setenv("LOG_LEVEL", tt.envValue)
			} else {
				t.Setenv("LOG_LEVEL", "")
			}

			Init()

			if atomicLevel.Level() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, atomicLevel.Level())
			}
		})
	}
}

func TestInitWithInvalidLevel(t *testing.T) {
	// Reset to Info level
	atomicLevel.SetLevel(zapcore.InfoLevel)

	t.Setenv("LOG_LEVEL", "invalid_level")

	Init()

	// Should remain at Info level due to invalid value
	if atomicLevel.Level() != zapcore.InfoLevel {
		t.Errorf("expected level to remain at Info, got %v", atomicLevel.Level())
	}
}

func TestLogOutput(t *testing.T) {
	// Save original logger
	originalLog := log
	originalLevel := atomicLevel
	defer func() {
		log = originalLog
		atomicLevel = originalLevel
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.DebugLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		atomicLevel,
	)

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// Test logging at different levels
	Debug("debug message", zap.String("key", "value"))
	Info("info message", zap.Int("count", 42))
	Warn("warn message")
	Error("error message")

	output := buf.String()

	// Verify each log level appears in output
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected info message in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("expected error message in output")
	}

	// Verify JSON structure and fields for each log line
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedEntries := []struct {
		level  string
		msg    string
		fields map[string]interface{}
	}{
		{"debug", "debug message", map[string]interface{}{"key": "value"}},
		{"info", "info message", map[string]interface{}{"count": float64(42)}},
		{"warn", "warn message", nil},
		{"error", "error message", nil},
	}

	if len(lines) != len(expectedEntries) {
		t.Fatalf("expected %d log lines, got %d", len(expectedEntries), len(lines))
	}

	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("line %d: failed to parse log line as JSON: %v", i, err)
			continue
		}
		if _, ok := logEntry["timestamp"]; !ok {
			t.Errorf("line %d: expected 'timestamp' field in log output", i)
		}
		if _, ok := logEntry["caller"]; !ok {
			t.Errorf("line %d: expected 'caller' field in log output", i)
		}
		if level, ok := logEntry["level"]; !ok {
			t.Errorf("line %d: expected 'level' field in log output", i)
		} else if level != expectedEntries[i].level {
			t.Errorf("line %d: expected level %q, got %q", i, expectedEntries[i].level, level)
		}
		if msg, ok := logEntry["msg"]; !ok {
			t.Errorf("line %d: expected 'msg' field in log output", i)
		} else if msg != expectedEntries[i].msg {
			t.Errorf("line %d: expected msg %q, got %q", i, expectedEntries[i].msg, msg)
		}
		for field, expectedVal := range expectedEntries[i].fields {
			if val, ok := logEntry[field]; !ok {
				t.Errorf("line %d: expected field %q in log output", i, field)
			} else if val != expectedVal {
				t.Errorf("line %d: expected field %q to be %v, got %v", i, field, expectedVal, val)
			}
		}
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Save original logger
	originalLog := log
	originalLevel := atomicLevel
	defer func() {
		log = originalLog
		atomicLevel = originalLevel
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.WarnLevel) // Set to Warn level

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		atomicLevel,
	)

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// Log at different levels
	Debug("debug message - should not appear")
	Info("info message - should not appear")
	Warn("warn message - should appear")
	Error("error message - should appear")

	output := buf.String()

	// Debug and Info should be filtered out
	if strings.Contains(output, "debug message") {
		t.Error("debug message should be filtered at Warn level")
	}
	if strings.Contains(output, "info message") {
		t.Error("info message should be filtered at Warn level")
	}

	// Warn and Error should appear
	if !strings.Contains(output, "warn message") {
		t.Error("expected warn message in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("expected error message in output")
	}
}
