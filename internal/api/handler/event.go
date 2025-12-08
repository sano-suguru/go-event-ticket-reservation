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
	eventService EventServiceInterface
}

func NewEventHandler(eventService EventServiceInterface) *EventHandler {
	return &EventHandler{eventService: eventService}
}

type CreateEventRequest struct {
	Name        string `json:"name" validate:"required" example:"東京ドームコンサート2025"`
	Description string `json:"description" example:"年末スペシャルコンサート"`
	Venue       string `json:"venue" example:"東京ドーム"`
	StartAt     string `json:"start_at" validate:"required" example:"2025-12-31T18:00:00+09:00"`
	EndAt       string `json:"end_at" validate:"required" example:"2025-12-31T21:00:00+09:00"`
	TotalSeats  int    `json:"total_seats" validate:"required,gt=0" example:"50000"`
}

type EventResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"東京ドームコンサート2025"`
	Description string `json:"description" example:"年末スペシャルコンサート"`
	Venue       string `json:"venue" example:"東京ドーム"`
	StartAt     string `json:"start_at" example:"2025-12-31T18:00:00+09:00"`
	EndAt       string `json:"end_at" example:"2025-12-31T21:00:00+09:00"`
	TotalSeats  int    `json:"total_seats" example:"50000"`
	CreatedAt   string `json:"created_at" example:"2025-12-06T10:00:00+09:00"`
	UpdatedAt   string `json:"updated_at" example:"2025-12-06T10:00:00+09:00"`
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

// Create godoc
// @Summary イベントを作成
// @Description 新しいイベントを作成します
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "イベント情報"
// @Success 201 {object} EventResponse
// @Failure 400 {object} map[string]string
// @Router /events [post]
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

// GetByID godoc
// @Summary イベントを取得
// @Description 指定IDのイベントを取得します
// @Tags events
// @Produce json
// @Param id path string true "イベントID"
// @Success 200 {object} EventResponse
// @Failure 404 {object} map[string]string
// @Router /events/{id} [get]
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

// List godoc
// @Summary イベント一覧を取得
// @Description イベントの一覧を取得します
// @Tags events
// @Produce json
// @Param limit query int false "取得件数" default(20)
// @Param offset query int false "オフセット" default(0)
// @Success 200 {array} EventResponse
// @Router /events [get]
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

// Update godoc
// @Summary イベントを更新
// @Description 指定IDのイベントを更新します
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "イベントID"
// @Param request body CreateEventRequest true "イベント情報"
// @Success 200 {object} EventResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /events/{id} [put]
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

// Delete godoc
// @Summary イベントを削除
// @Description 指定IDのイベントを削除します
// @Tags events
// @Param id path string true "イベントID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /events/{id} [delete]
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
