package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
)

// ErrorResponse はエラーレスポンスの統一フォーマット
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// CustomHTTPErrorHandler はカスタムエラーハンドラー
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var (
		code    = http.StatusInternalServerError
		message = "内部サーバーエラー"
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		} else {
			message = http.StatusText(code)
		}
	}

	// エラーログを出力（5xx エラーの場合）
	if code >= 500 {
		logger.Error("サーバーエラー",
			zap.Int("status", code),
			zap.String("path", c.Request().URL.Path),
			zap.Error(err),
		)
	}

	// JSONレスポンスを返す
	if err := c.JSON(code, ErrorResponse{
		Error: message,
		Code:  code,
	}); err != nil {
		logger.Error("エラーレスポンス送信失敗", zap.Error(err))
	}
}
