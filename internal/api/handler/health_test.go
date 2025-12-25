package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

func TestHealthHandler_Check(t *testing.T) {
	// Setup
	e := NewTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHealthHandler()

	// Act
	err := h.Check(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	assert.Contains(t, rec.Body.String(), `"timestamp"`)
}

func TestNewHealthHandler(t *testing.T) {
	h := NewHealthHandler()
	assert.NotNil(t, h)
}

func TestToEventResponse(t *testing.T) {
	now := time.Now()
	e := &event.Event{
		ID:          "event-123",
		Name:        "テストイベント",
		Description: "テスト説明",
		Venue:       "テスト会場",
		StartAt:     now,
		EndAt:       now.Add(3 * time.Hour),
		TotalSeats:  100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	resp := toEventResponse(e)

	assert.Equal(t, e.ID, resp.ID)
	assert.Equal(t, e.Name, resp.Name)
	assert.Equal(t, e.Description, resp.Description)
	assert.Equal(t, e.Venue, resp.Venue)
	assert.Equal(t, e.TotalSeats, resp.TotalSeats)
	assert.Equal(t, e.StartAt.Format(time.RFC3339), resp.StartAt)
	assert.Equal(t, e.EndAt.Format(time.RFC3339), resp.EndAt)
	assert.Equal(t, e.CreatedAt.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, e.UpdatedAt.Format(time.RFC3339), resp.UpdatedAt)
}

func TestToSeatResponse(t *testing.T) {
	now := time.Now()
	s := &seat.Seat{
		ID:         "seat-123",
		EventID:    "event-456",
		SeatNumber: "A-1",
		Status:     seat.StatusAvailable,
		Price:      5000,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp := toSeatResponse(s)

	assert.Equal(t, s.ID, resp.ID)
	assert.Equal(t, s.EventID, resp.EventID)
	assert.Equal(t, s.SeatNumber, resp.SeatNumber)
	assert.Equal(t, string(s.Status), resp.Status)
	assert.Equal(t, s.Price, resp.Price)
}

func TestToReservationResponse(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute)
	r := &reservation.Reservation{
		ID:             "res-123",
		EventID:        "event-456",
		UserID:         "user-789",
		SeatIDs:        []string{"seat-1", "seat-2"},
		Status:         reservation.StatusPending,
		TotalAmount:    10000,
		IdempotencyKey: "idem-key",
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	resp := toReservationResponse(r)

	assert.Equal(t, r.ID, resp.ID)
	assert.Equal(t, r.EventID, resp.EventID)
	assert.Equal(t, r.UserID, resp.UserID)
	assert.Equal(t, r.SeatIDs, resp.SeatIDs)
	assert.Equal(t, string(r.Status), resp.Status)
	assert.Equal(t, r.TotalAmount, resp.TotalAmount)
	assert.Equal(t, r.ExpiresAt, resp.ExpiresAt)
	assert.Equal(t, r.CreatedAt, resp.CreatedAt)
}
