package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsBasicAuth_NoCredentials(t *testing.T) {
	// 認証設定がない場合はスキップ
	os.Unsetenv("METRICS_USER")
	os.Unsetenv("METRICS_PASSWORD")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := MetricsBasicAuth()(func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "metrics", rec.Body.String())
}

func TestMetricsBasicAuth_ValidCredentials(t *testing.T) {
	// 認証設定あり
	os.Setenv("METRICS_USER", "testuser")
	os.Setenv("METRICS_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("METRICS_USER")
		os.Unsetenv("METRICS_PASSWORD")
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	// Basic認証ヘッダーを設定
	auth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
	req.Header.Set("Authorization", "Basic "+auth)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := MetricsBasicAuth()(func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsBasicAuth_InvalidCredentials(t *testing.T) {
	// 認証設定あり
	os.Setenv("METRICS_USER", "testuser")
	os.Setenv("METRICS_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("METRICS_USER")
		os.Unsetenv("METRICS_PASSWORD")
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	// 間違った認証情報
	auth := base64.StdEncoding.EncodeToString([]byte("wronguser:wrongpass"))
	req.Header.Set("Authorization", "Basic "+auth)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := MetricsBasicAuth()(func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	err := handler(c)
	// Basic認証失敗時はHTTPErrorが返る
	if err != nil {
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	} else {
		// エラーがない場合はレスポンスコードをチェック
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestMetricsBasicAuth_NoAuthHeader(t *testing.T) {
	// 認証設定あり
	os.Setenv("METRICS_USER", "testuser")
	os.Setenv("METRICS_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("METRICS_USER")
		os.Unsetenv("METRICS_PASSWORD")
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	// Authorization ヘッダーなし

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := MetricsBasicAuth()(func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	err := handler(c)
	// Basic認証ヘッダーがない場合は401
	if err != nil {
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	}
}

func TestLoadMetricsConfig(t *testing.T) {
	tests := []struct {
		name        string
		user        string
		password    string
		wantEnabled bool
	}{
		{
			name:        "両方設定あり",
			user:        "user",
			password:    "pass",
			wantEnabled: true,
		},
		{
			name:        "ユーザーのみ",
			user:        "user",
			password:    "",
			wantEnabled: false,
		},
		{
			name:        "パスワードのみ",
			user:        "",
			password:    "pass",
			wantEnabled: false,
		},
		{
			name:        "両方なし",
			user:        "",
			password:    "",
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != "" {
				os.Setenv("METRICS_USER", tt.user)
			} else {
				os.Unsetenv("METRICS_USER")
			}
			if tt.password != "" {
				os.Setenv("METRICS_PASSWORD", tt.password)
			} else {
				os.Unsetenv("METRICS_PASSWORD")
			}
			defer func() {
				os.Unsetenv("METRICS_USER")
				os.Unsetenv("METRICS_PASSWORD")
			}()

			cfg := LoadMetricsConfig()
			assert.Equal(t, tt.wantEnabled, cfg.IsEnabled())
		})
	}
}
