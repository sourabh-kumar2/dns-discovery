// Package logger provides structured logging functionality for the dns-discovery project.
// It encapsulates the initialization and cleanup of a production-ready logger using Uber's zap package.
// Consumers of this package should initialize the logger via InitLogger and use the global Logger for logging.
package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type contextKey string

const (
	requestIDKey     contextKey = "requestId"
	transactionIDKey contextKey = "transactionId"
	logFilePath                 = "dns-server.log"
)

// logger is the globally accessible zap.Logger instance.
// It must be initialized by calling InitLogger before any logging occurs.
var logger *zap.Logger

// InitLogger initializes the production logger using zap.NewProduction and assigns it to Logger.
// It returns an error if the logger cannot be initialized.
//
// Example usage:
//
//	err := logger.InitLogger()
//	if err != nil {
//	    // Handle error appropriately
//	}
func InitLogger(debugMode bool) error {
	// Configure common encoder settings.
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	logFile, err := getLogFile(debugMode)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)
	if debugMode {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(logFile),
			zap.DebugLevel,
		)
	}

	logger = zap.New(zapcore.NewTee(core), zap.AddCaller())
	Log(zap.InfoLevel, "Initialized logger", zap.Bool("debugMode", debugMode))
	return nil
}

func getLogFile(debugMode bool) (*os.File, error) {
	if debugMode {
		return os.Stdout, nil
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return logFile, nil
}

// SyncLogger flushes any buffered log entries.
// It should be called before the application exits to ensure that all logs are written.
// For example, you can defer this function in your main function:
//
//	defer logger.SyncLogger()
func SyncLogger() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// WithRequestID attaches a request ID to the context.
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

// WithTransactionID attaches a transaction ID to the context.
func WithTransactionID(ctx context.Context, txID uint16) context.Context {
	return context.WithValue(ctx, transactionIDKey, txID)
}

// RequestIDFromContext retrieves the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// TransactionIDFromContext retrieves the transaction ID from context.
func TransactionIDFromContext(ctx context.Context) uint16 {
	if v, ok := ctx.Value(transactionIDKey).(uint16); ok {
		return v
	}
	return 0
}

// Log wrapper for zap.Log.
func Log(level zapcore.Level, msg string, fields ...zap.Field) {
	callerSkipLogger := logger.WithOptions(zap.AddCallerSkip(1))
	callerSkipLogger.Log(level, msg, fields...)
}

// LogWithContext automatically extracts request and transaction IDs and logs them.
func LogWithContext(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	callerSkipLogger := logger.WithOptions(zap.AddCallerSkip(1))

	if reqID := RequestIDFromContext(ctx); reqID != "" {
		fields = append(fields, zap.Any("requestId", reqID))
	}
	if txID := TransactionIDFromContext(ctx); txID != 0 {
		fields = append(fields, zap.Any("transactionId", txID))
	}

	callerSkipLogger.Log(level, msg, fields...)
}

// CaptureLogs captures log output for testing.
func CaptureLogs(f func()) []observer.LoggedEntry {
	// Create an observer core to capture logs
	core, logs := observer.New(zapcore.DebugLevel)
	testLogger := zap.New(core)

	// Swap global logger with test logger
	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	// Run the function that generates logs
	f()

	// Return captured logs
	return logs.All()
}

// InitTestLogger for testing.
func InitTestLogger() {
	logger = zap.NewNop()
}
