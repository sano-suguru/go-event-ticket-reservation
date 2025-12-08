package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/metrics"
)

func TestSetupMiddleware(t *testing.T) {
	e := echo.New()

	// ミドルウェア設定が正常に動作することを確認
	SetupMiddleware(e)

	// ミドルウェアが設定されていることを確認
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}

func TestRequestLogger(t *testing.T) {
	e := echo.New()

	// RequestLogger ミドルウェアを適用
	e.Use(RequestLogger())

	// テスト用ハンドラー
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_WithError(t *testing.T) {
	e := echo.New()

	e.Use(RequestLogger())

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRequestLogger_ServerError(t *testing.T) {
	e := echo.New()

	e.Use(RequestLogger())

	e.GET("/server-error", func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "internal error")
	})

	req := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRequestIDMiddleware(t *testing.T) {
	e := echo.New()

	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// リクエストIDなしのリクエスト
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// レスポンスにリクエストIDが設定されていることを確認
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))
}

func TestRequestIDMiddleware_WithExistingRequestID(t *testing.T) {
	e := echo.New()

	e.Use(RequestIDMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// 既存のリクエストIDを持つリクエスト
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(echo.HeaderXRequestID, "existing-request-id")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// 既存のリクエストIDが維持されていることを確認
	assert.Equal(t, "existing-request-id", rec.Header().Get(echo.HeaderXRequestID))
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
}

func TestRandomString(t *testing.T) {
	s1 := randomString(8)
	s2 := randomString(8)

	assert.Len(t, s1, 8)
	assert.Len(t, s2, 8)
}

func TestRandomString_Length(t *testing.T) {
	tests := []int{4, 8, 16, 32}

	for _, length := range tests {
		s := randomString(length)
		assert.Len(t, s, length)
	}
}

func TestPrometheusMiddleware(t *testing.T) {
	e := echo.New()

	// テスト用のメトリクスを作成
	reg := prometheus.NewRegistry()
	m := metrics.NewWithRegistry(reg)

	e.Use(PrometheusMiddleware(m))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// メトリクスが記録されているか確認
	families, err := reg.Gather()
	assert.NoError(t, err)

	var foundRequests, foundDuration bool
	for _, f := range families {
		if f.GetName() == "http_requests_total" {
			foundRequests = true
		}
		if f.GetName() == "http_request_duration_seconds" {
			foundDuration = true
		}
	}
	assert.True(t, foundRequests, "http_requests_total should be recorded")
	assert.True(t, foundDuration, "http_request_duration_seconds should be recorded")
}

func TestPrometheusMiddleware_WithError(t *testing.T) {
	e := echo.New()

	reg := prometheus.NewRegistry()
	m := metrics.NewWithRegistry(reg)

	e.Use(PrometheusMiddleware(m))

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
