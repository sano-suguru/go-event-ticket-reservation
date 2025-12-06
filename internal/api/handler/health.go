package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthHandler はヘルスチェックハンドラー
type HealthHandler struct{}

// NewHealthHandler はHealthHandlerを作成する
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse はヘルスチェックのレスポンス
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Check はヘルスチェックを行う
// @Summary ヘルスチェック
// @Description アプリケーションの健全性を確認する
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Check(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
