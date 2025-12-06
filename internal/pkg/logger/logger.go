package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	log = NewLogger("development")
}

func NewLogger(env string) *zap.Logger {
	var config zap.Config
	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(lvl)); err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}

	logger, _ := config.Build()
	return logger
}

func Get() *zap.Logger {
	return log
}

func Set(l *zap.Logger) {
	log = l
}

func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return log.With(fields...)
}

func Sync() error {
	return log.Sync()
}
