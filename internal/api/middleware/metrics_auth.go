package middleware

import (
	"crypto/subtle"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// MetricsBasicAuth は /metrics エンドポイント用の Basic 認証ミドルウェア
// 環境変数 METRICS_USER と METRICS_PASSWORD が設定されている場合のみ認証を要求
// 設定されていない場合は認証をスキップ（ローカル開発用）
func MetricsBasicAuth() echo.MiddlewareFunc {
	expectedUser := os.Getenv("METRICS_USER")
	expectedPass := os.Getenv("METRICS_PASSWORD")

	// 認証設定がない場合はスキップ（パススルー）
	if expectedUser == "" || expectedPass == "" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// タイミング攻撃を防ぐため ConstantTimeCompare を使用
		userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1

		return userMatch && passMatch, nil
	})
}

// MetricsConfig はメトリクス認証の設定
type MetricsConfig struct {
	User     string
	Password string
}

// LoadMetricsConfig は環境変数からメトリクス認証設定を読み込む
func LoadMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		User:     os.Getenv("METRICS_USER"),
		Password: os.Getenv("METRICS_PASSWORD"),
	}
}

// IsEnabled は認証が有効かどうかを返す
func (c *MetricsConfig) IsEnabled() bool {
	return c.User != "" && c.Password != ""
}
