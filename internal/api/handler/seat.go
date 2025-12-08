package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

type SeatHandler struct {
	service SeatServiceInterface
}

func NewSeatHandler(s SeatServiceInterface) *SeatHandler {
	return &SeatHandler{service: s}
}

type CreateSeatRequest struct {
	SeatNumber string `json:"seat_number" validate:"required"`
	Price      int    `json:"price" validate:"required,min=0"`
}

type CreateBulkSeatsRequest struct {
	Prefix string `json:"prefix" validate:"required"`
	Count  int    `json:"count" validate:"required,min=1,max=1000"`
	Price  int    `json:"price" validate:"required,min=0"`
}

type SeatResponse struct {
	ID         string  `json:"id"`
	EventID    string  `json:"event_id"`
	SeatNumber string  `json:"seat_number"`
	Status     string  `json:"status"`
	Price      int     `json:"price"`
	ReservedBy *string `json:"reserved_by,omitempty"`
}

func toSeatResponse(s *seat.Seat) SeatResponse {
	return SeatResponse{
		ID: s.ID, EventID: s.EventID, SeatNumber: s.SeatNumber,
		Status: string(s.Status), Price: s.Price, ReservedBy: s.ReservedBy,
	}
}

func (h *SeatHandler) GetByEvent(c echo.Context) error {
	eventID := c.Param("event_id")
	availableOnly := c.QueryParam("available") == "true"
	var seats []*seat.Seat
	var err error
	if availableOnly {
		seats, err = h.service.GetAvailableSeatsByEvent(c.Request().Context(), eventID)
	} else {
		seats, err = h.service.GetSeatsByEvent(c.Request().Context(), eventID)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	resp := make([]SeatResponse, len(seats))
	for i, s := range seats {
		resp[i] = toSeatResponse(s)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *SeatHandler) Create(c echo.Context) error {
	eventID := c.Param("event_id")
	var req CreateSeatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "無効なリクエスト"})
	}
	s, err := h.service.CreateSeat(c.Request().Context(), application.CreateSeatInput{
		EventID: eventID, SeatNumber: req.SeatNumber, Price: req.Price,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, toSeatResponse(s))
}

func (h *SeatHandler) CreateBulk(c echo.Context) error {
	eventID := c.Param("event_id")
	var req CreateBulkSeatsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "無効なリクエスト"})
	}
	seats, err := h.service.CreateBulkSeats(c.Request().Context(), application.CreateBulkSeatsInput{
		EventID: eventID, Prefix: req.Prefix, Count: req.Count, Price: req.Price,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := make([]SeatResponse, len(seats))
	for i, s := range seats {
		resp[i] = toSeatResponse(s)
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *SeatHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	s, err := h.service.GetSeat(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toSeatResponse(s))
}

func (h *SeatHandler) CountAvailable(c echo.Context) error {
	eventID := c.Param("event_id")
	count, err := h.service.CountAvailableSeats(c.Request().Context(), eventID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]int{"count": count})
}
