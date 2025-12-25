package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
)

type ReservationHandler struct {
	service ReservationServiceInterface
}

func NewReservationHandler(s ReservationServiceInterface) *ReservationHandler {
	return &ReservationHandler{service: s}
}

type CreateReservationRequest struct {
	EventID        string   `json:"event_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeatIDs        []string `json:"seat_ids" validate:"required,min=1" example:"seat-A1,seat-A2"`
	IdempotencyKey string   `json:"idempotency_key" validate:"required" example:"order-2025-001"`
}

type ReservationResponse struct {
	ID          string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventID     string     `json:"event_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID      string     `json:"user_id" example:"user-123"`
	SeatIDs     []string   `json:"seat_ids" example:"seat-A1,seat-A2"`
	Status      string     `json:"status" example:"pending"`
	TotalAmount int        `json:"total_amount" example:"10000"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toReservationResponse(r *reservation.Reservation) ReservationResponse {
	return ReservationResponse{
		ID: r.ID, EventID: r.EventID, UserID: r.UserID,
		SeatIDs: r.SeatIDs, Status: string(r.Status),
		TotalAmount: r.TotalAmount, ExpiresAt: r.ExpiresAt,
		ConfirmedAt: r.ConfirmedAt, CreatedAt: r.CreatedAt,
	}
}

// Create godoc
// @Summary 予約を作成
// @Description 座席を仮押さえします（15分間有効）
// @Tags reservations
// @Accept json
// @Produce json
// @Param X-User-ID header string true "ユーザーID"
// @Param request body CreateReservationRequest true "予約情報"
// @Success 201 {object} ReservationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string "座席が既に予約済み"
// @Router /reservations [post]
func (h *ReservationHandler) Create(c echo.Context) error {
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "ユーザーIDが必要です")
	}
	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエスト")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	r, err := h.service.CreateReservation(c.Request().Context(), application.CreateReservationInput{
		EventID: req.EventID, UserID: userID, SeatIDs: req.SeatIDs, IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, toReservationResponse(r))
}

// GetByID godoc
// @Summary 予約を取得
// @Description 指定IDの予約を取得します
// @Tags reservations
// @Produce json
// @Param id path string true "予約ID"
// @Success 200 {object} ReservationResponse
// @Failure 404 {object} map[string]string
// @Router /reservations/{id} [get]
func (h *ReservationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.GetReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}

// GetUserReservations godoc
// @Summary ユーザーの予約一覧を取得
// @Description ログインユーザーの予約一覧を取得します
// @Tags reservations
// @Produce json
// @Param X-User-ID header string true "ユーザーID"
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} ReservationResponse
// @Failure 401 {object} map[string]string
// @Router /reservations [get]
func (h *ReservationHandler) GetUserReservations(c echo.Context) error {
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "ユーザーIDが必要です")
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	reservations, err := h.service.GetUserReservations(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	resp := make([]ReservationResponse, len(reservations))
	for i, r := range reservations {
		resp[i] = toReservationResponse(r)
	}
	return c.JSON(http.StatusOK, resp)
}

// Confirm godoc
// @Summary 予約を確定
// @Description 仮押さえ中の予約を確定します
// @Tags reservations
// @Produce json
// @Param id path string true "予約ID"
// @Success 200 {object} ReservationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /reservations/{id}/confirm [post]
func (h *ReservationHandler) Confirm(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.ConfirmReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}

// Cancel godoc
// @Summary 予約をキャンセル
// @Description 予約をキャンセルし、座席を解放します
// @Tags reservations
// @Produce json
// @Param id path string true "予約ID"
// @Success 200 {object} ReservationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /reservations/{id}/cancel [post]
func (h *ReservationHandler) Cancel(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.CancelReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}
