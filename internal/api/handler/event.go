package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
)

type EventHandler struct {
	eventService *application.EventService
}

func NewEventHandler(eventService *application.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

type CreateEventRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	StartAt     string `json:"start_at" validate:"required"`
	EndAt       string `json:"end_at" validate:"required"`
	TotalSeats  int    `json:"total_seats" validate:"required,gt=0"`
}

type EventResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at"`
	TotalSeats  int    `json:"total_seats"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toEventResponse(e *event.Event) *EventResponse {
	return &EventResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Venue:       e.Venue,
		StartAt:     e.StartAt.Format(time.RFC3339),
		EndAt:       e.EndAt.Format(time.RFC3339),
		TotalSeats:  e.TotalSeats,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *EventHandler) Create(c echo.Context) error {
	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエストの形式が不正です"})
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "開始時刻の形式が不正です"})
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "終了時刻の形式が不正です"})
	}

	input := application.CreateEventInput{
		Name:        req.Name,
		Description: req.Description,
		Venue:       req.Venue,
		StartAt:     startAt,
		EndAt:       endAt,
		TotalSeats:  req.TotalSeats,
	}

	e, err := h.eventService.CreateEvent(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, toEventResponse(e))
}

func (h *EventHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	e, err := h.eventService.GetEvent(c.Request().Context(), id)
	if err != nil {
		if err == event.ErrEventNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "イベントが見つかりません"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toEventResponse(e))
}

func (h *EventHandler) List(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	events, err := h.eventService.ListEvents(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	responses := make([]*EventResponse, len(events))
	for i, e := range events {
		responses[i] = toEventResponse(e)
	}
	return c.JSON(http.StatusOK, responses)
}

func (h *EventHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエストの形式が不正です"})
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "開始時刻の形式が不正です"})
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "終了時刻の形式が不正です"})
	}

	input := application.UpdateEventInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Venue:       req.Venue,
		StartAt:     startAt,
		EndAt:       endAt,
		TotalSeats:  req.TotalSeats,
	}

	e, err := h.eventService.UpdateEvent(c.Request().Context(), input)
	if err != nil {
		if err == event.ErrEventNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "イベントが見つかりません"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, toEventResponse(e))
}

func (h *EventHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	err := h.eventService.DeleteEvent(c.Request().Context(), id)
	if err != nil {
		if err == event.ErrEventNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "イベントが見つかりません"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
