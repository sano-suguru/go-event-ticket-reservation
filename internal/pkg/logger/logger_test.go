package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLogger_Development(t *testing.T) {
	logger := NewLogger("development")
	require.NotNil(t, logger)

	// 開発環境のロガーが正常に動作することを確認
	logger.Info("test message")
}

func TestNewLogger_Production(t *testing.T) {
	logger := NewLogger("production")
	require.NotNil(t, logger)

	logger.Info("test message")
}

func TestNewLogger_WithLogLevel(t *testing.T) {
	// LOG_LEVELを設定
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	logger := NewLogger("development")
	require.NotNil(t, logger)
}

func TestNewLogger_WithInvalidLogLevel(t *testing.T) {
	// 無効なLOG_LEVELを設定
	os.Setenv("LOG_LEVEL", "invalid_level")
	defer os.Unsetenv("LOG_LEVEL")

	// 無効なレベルでも正常に動作することを確認
	logger := NewLogger("development")
	require.NotNil(t, logger)
}

func TestGet(t *testing.T) {
	logger := Get()
	require.NotNil(t, logger)
}

func TestSet(t *testing.T) {
	originalLogger := Get()
	defer Set(originalLogger) // テスト後に元に戻す

	newLogger := zap.NewNop()
	Set(newLogger)

	assert.Equal(t, newLogger, Get())
}

func TestInfo(t *testing.T) {
	// ログ関数がパニックしないことを確認
	assert.NotPanics(t, func() {
		Info("test info message")
	})
}

func TestError(t *testing.T) {
	assert.NotPanics(t, func() {
		Error("test error message")
	})
}

func TestDebug(t *testing.T) {
	assert.NotPanics(t, func() {
		Debug("test debug message")
	})
}

func TestWarn(t *testing.T) {
	assert.NotPanics(t, func() {
		Warn("test warn message")
	})
}

func TestWith(t *testing.T) {
	logger := With(zap.String("key", "value"))
	require.NotNil(t, logger)
}

func TestSync(t *testing.T) {
	// Syncはエラーを返す可能性があるが、パニックしないことを確認
	assert.NotPanics(t, func() {
		_ = Sync()
	})
}

func TestInfo_WithFields(t *testing.T) {
	assert.NotPanics(t, func() {
		Info("test message",
			zap.String("string_field", "value"),
			zap.Int("int_field", 42),
			zap.Bool("bool_field", true),
		)
	})
}

func TestError_WithFields(t *testing.T) {
	assert.NotPanics(t, func() {
		Error("test error",
			zap.String("error_code", "E001"),
			zap.Int("status", 500),
		)
	})
}
