package logger

import (
	"bytes"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestLoggerInitialization ensures that the logger initializes correctly.
func TestLoggerInitialization(t *testing.T) {
	err := InitLogger(true)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
}
func TestLoggerOutput(t *testing.T) {
	// Capture logs using a buffer
	var buf bytes.Buffer

	// Create a custom core that writes to the buffer
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:     "timestamp",
			LevelKey:    "level",
			MessageKey:  "message",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
		}),
		zapcore.AddSync(&buf), // Write logs to the buffer
		zapcore.DebugLevel,    // Log everything
	)

	logger := zap.New(core)
	logger.Info("Hello, Logger!")

	// Read buffer output and validate
	logOutput := buf.String()
	if len(logOutput) == 0 {
		t.Fatal("Expected log output, but got empty string")
	}
	if !contains(logOutput, "Hello, Logger!") {
		t.Errorf("Expected log message to contain 'Hello, Logger!', but got: %s", logOutput)
	}
}

// Helper function to check if a string contains another string
func contains(str, substr string) bool {
	return bytes.Contains([]byte(str), []byte(substr))
}

func TestLogging(t *testing.T) {
	logs := CaptureLogs(func() {
		logger.Info("Test log", zap.String("key", "value"))
	})

	found := false
	for _, log := range logs {
		if log.Message == "Test log" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected log message not found")
	}
}
