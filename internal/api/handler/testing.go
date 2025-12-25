package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/api"
)

// NewTestEcho はテスト用のEchoインスタンスを作成する
func NewTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = api.NewValidator()
	return e
}
