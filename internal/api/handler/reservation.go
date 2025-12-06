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
	service *application.ReservationService
}

func NewReservationHandler(s *application.ReservationService) *ReservationHandler {
	return &ReservationHandler{service: s}
}

type CreateReservationRequest struct {
	EventID        string   `json:"event_id" validate:"required"`
	SeatIDs        []string `json:"seat_ids" validate:"required,min=1"`
	IdempotencyKey string   `json:"idempotency_key" validate:"required"`
}

type ReservationResponse struct {
	ID          string     `json:"id"`
	EventID     string     `json:"event_id"`
	UserID      string     `json:"user_id"`
	SeatIDs     []string   `json:"seat_ids"`
	Status      string     `json:"status"`
	TotalAmount int        `json:"total_amount"`
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

func (h *ReservationHandler) Create(c echo.Context) error {
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "ユーザーIDが必要です"})
	}
	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "無効なリクエスト"})
	}
	r, err := h.service.CreateReservation(c.Request().Context(), application.CreateReservationInput{
		EventID: req.EventID, UserID: userID, SeatIDs: req.SeatIDs, IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toReservationResponse(r))
}

func (h *ReservationHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.GetReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}

func (h *ReservationHandler) GetUserReservations(c echo.Context) error {
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "ユーザーIDが必要です"})
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	reservations, err := h.service.GetUserReservations(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	resp := make([]ReservationResponse, len(reservations))
	for i, r := range reservations {
		resp[i] = toReservationResponse(r)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *ReservationHandler) Confirm(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.ConfirmReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	id := c.Param("id")
	r, err := h.service.CancelReservation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, reservation.ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toReservationResponse(r))
}
