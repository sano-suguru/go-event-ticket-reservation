package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupMiddleware は共通ミドルウェアを設定する
func SetupMiddleware(e *echo.Echo) {
	// リクエストID
	e.Use(middleware.RequestID())

	// 構造化リクエストログ（zap）
	e.Use(RequestLogger())

	// パニックリカバリー
	e.Use(middleware.Recover())

	// CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
}
