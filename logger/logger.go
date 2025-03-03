// Package logger provides structured logging functionality for the dns-discovery project.
// It encapsulates the initialization and cleanup of a production-ready logger using Uber's zap package.
// Consumers of this package should initialize the logger via InitLogger and use the global Logger for logging.
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Logger is the globally accessible zap.Logger instance.
// It must be initialized by calling InitLogger before any logging occurs.
var Logger *zap.Logger

var encoder = os.Getenv("ENCODER")

// InitLogger initializes the production logger using zap.NewProduction and assigns it to Logger.
// It returns an error if the logger cannot be initialized.
//
// Example usage:
//
//	err := logger.InitLogger()
//	if err != nil {
//	    // Handle error appropriately
//	}
func InitLogger() error {
	var err error
	cfg := zap.NewProductionConfig()
	// Set the key for timestamps in the output.
	cfg.EncoderConfig.TimeKey = "timestamp"

	if encoder == "console" {
		cfg.Encoding = "console"
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	// Use ISO8601 format for the time encoder.
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Build the logger.
	Logger, err = cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}
	return nil
}

// SyncLogger flushes any buffered log entries.
// It should be called before the application exits to ensure that all logs are written.
// For example, you can defer this function in your main function:
//
//	defer logger.SyncLogger()
func SyncLogger() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// CaptureLogs captures log output for testing.
func CaptureLogs(f func()) []observer.LoggedEntry {
	// Create an observer core to capture logs
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := zap.New(core)

	// Swap global logger with test logger
	oldLogger := Logger
	Logger = testLogger
	defer func() { Logger = oldLogger }()

	// Run the function that generates logs
	f()

	// Return captured logs
	return logs.All()
}
