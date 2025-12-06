package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
)

// RequestLogger はリクエストの構造化ログを出力するミドルウェア
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			// リクエストIDを生成または取得
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = res.Header().Get(echo.HeaderXRequestID)
			}

			// リクエスト処理
			err := next(c)

			// レスポンス後のログ
			latency := time.Since(start)

			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("query", req.URL.RawQuery),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.Duration("latency", latency),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("request failed", fields...)
			} else if res.Status >= 500 {
				logger.Error("server error", fields...)
			} else if res.Status >= 400 {
				logger.Warn("client error", fields...)
			} else {
				logger.Info("request completed", fields...)
			}

			return err
		}
	}
}

// RequestIDMiddleware はリクエストIDを生成・付与するミドルウェア
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = generateRequestID()
			}
			res.Header().Set(echo.HeaderXRequestID, requestID)

			return next(c)
		}
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
