package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/metrics"
)

// PrometheusMiddleware はHTTPメトリクスを収集するミドルウェア
func PrometheusMiddleware(m *metrics.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start).Seconds()
			status := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				}
			}

			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}

			method := c.Request().Method
			statusCode := strconv.Itoa(status)

			// メトリクス記録
			m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
			m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

			return err
		}
	}
}
