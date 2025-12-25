package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator はEcho用のカスタムバリデーター
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator は新しいバリデーターを作成する
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate はリクエストのバリデーションを実行する
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}
	return nil
}
