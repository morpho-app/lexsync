package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger is a singleton instance of zap.Logger
var (
	instance *zap.Logger
	once     sync.Once
)

// GetLogger returns the singleton zap.Logger instance
func GetLogger() *zap.Logger {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		logger, err := config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
		instance = logger
	})
	return instance
}

// SyncLogger syncs any buffered log entries
func SyncLogger() {
	if instance != nil {
		instance.Sync()
	}
}
