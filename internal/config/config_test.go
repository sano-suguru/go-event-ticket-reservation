package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// 環境変数をクリア
	envVars := []string{
		"PORT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"DATABASE_URL", "REDIS_URL",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	cfg := Load()

	// Server defaults
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)

	// Database defaults
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "postgres", cfg.Database.Password)
	assert.Equal(t, "ticket_reservation", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	// Redis defaults
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
}

func TestLoad_CustomValues(t *testing.T) {
	// 環境変数を設定
	os.Setenv("PORT", "9090")
	os.Setenv("SERVER_READ_TIMEOUT", "60s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "120s")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "redispass")
	os.Setenv("REDIS_DB", "1")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("SERVER_READ_TIMEOUT")
		os.Unsetenv("SERVER_WRITE_TIMEOUT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
	}()

	cfg := Load()

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 60*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 120*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, "6380", cfg.Redis.Port)
	assert.Equal(t, "redispass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
}

func TestLoad_DatabaseURL(t *testing.T) {
	// DATABASE_URLを設定（Railway形式）
	os.Setenv("DATABASE_URL", "postgres://railwayuser:railwaypass@postgres.railway.app:5432/railway?sslmode=require")
	defer os.Unsetenv("DATABASE_URL")

	cfg := Load()

	assert.Equal(t, "postgres.railway.app", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "railwayuser", cfg.Database.User)
	assert.Equal(t, "railwaypass", cfg.Database.Password)
	assert.Equal(t, "railway", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
}

func TestLoad_DatabaseURL_WithoutSSLMode(t *testing.T) {
	// DATABASE_URLを設定（sslmode なし）
	os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/dbname")
	defer os.Unsetenv("DATABASE_URL")

	cfg := Load()

	assert.Equal(t, "host", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "user", cfg.Database.User)
	assert.Equal(t, "pass", cfg.Database.Password)
	assert.Equal(t, "dbname", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode) // デフォルトで require
}

func TestLoad_RedisURL(t *testing.T) {
	// REDIS_URLを設定（Railway形式）
	os.Setenv("REDIS_URL", "redis://:redispassword@redis.railway.app:6380")
	defer os.Unsetenv("REDIS_URL")

	cfg := Load()

	assert.Equal(t, "redis.railway.app", cfg.Redis.Host)
	assert.Equal(t, "6380", cfg.Redis.Port)
	assert.Equal(t, "redispassword", cfg.Redis.Password)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=postgres")
	assert.Contains(t, dsn, "password=secret")
	assert.Contains(t, dsn, "dbname=testdb")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	addr := cfg.Addr()

	assert.Equal(t, "localhost:6379", addr)
}

func TestGetEnv(t *testing.T) {
	// 環境変数が設定されている場合
	os.Setenv("TEST_ENV_VAR", "custom_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result := getEnv("TEST_ENV_VAR", "default")
	assert.Equal(t, "custom_value", result)

	// 環境変数が設定されていない場合
	result = getEnv("NON_EXISTENT_VAR", "default")
	assert.Equal(t, "default", result)
}

func TestGetIntEnv(t *testing.T) {
	// 有効な整数
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getIntEnv("TEST_INT", 0)
	assert.Equal(t, 42, result)

	// 無効な整数
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_INT")

	result = getIntEnv("TEST_INVALID_INT", 99)
	assert.Equal(t, 99, result)

	// 存在しない変数
	result = getIntEnv("NON_EXISTENT_INT", 100)
	assert.Equal(t, 100, result)
}

func TestGetDurationEnv(t *testing.T) {
	// 有効な期間
	os.Setenv("TEST_DURATION", "5m")
	defer os.Unsetenv("TEST_DURATION")

	result := getDurationEnv("TEST_DURATION", time.Second)
	assert.Equal(t, 5*time.Minute, result)

	// 無効な期間
	os.Setenv("TEST_INVALID_DURATION", "invalid")
	defer os.Unsetenv("TEST_INVALID_DURATION")

	result = getDurationEnv("TEST_INVALID_DURATION", 30*time.Second)
	assert.Equal(t, 30*time.Second, result)

	// 存在しない変数
	result = getDurationEnv("NON_EXISTENT_DURATION", time.Minute)
	assert.Equal(t, time.Minute, result)
}

func TestLoad_InvalidURLs(t *testing.T) {
	// 無効なDATABASE_URL
	os.Setenv("DATABASE_URL", "://invalid-url")
	defer os.Unsetenv("DATABASE_URL")

	cfg := Load()
	require.NotNil(t, cfg)
	// パースに失敗した場合はデフォルト値が使用される
	assert.Equal(t, "localhost", cfg.Database.Host)
}
