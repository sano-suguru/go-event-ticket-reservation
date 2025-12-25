package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
)

// MockReservationService はReservationServiceInterfaceのモック
type MockReservationService struct {
	mock.Mock
}

func (m *MockReservationService) CreateReservation(ctx context.Context, input application.CreateReservationInput) (*reservation.Reservation, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationService) GetReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationService) GetUserReservations(ctx context.Context, userID string, limit, offset int) ([]*reservation.Reservation, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reservation.Reservation), args.Error(1)
}

func (m *MockReservationService) ConfirmReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationService) CancelReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationService) CancelExpiredReservations(ctx context.Context, expireAfter time.Duration) (int, error) {
	args := m.Called(ctx, expireAfter)
	return args.Int(0), args.Error(1)
}

func TestReservationHandler_Create(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常に予約を作成できる", func(t *testing.T) {
		mockService := new(MockReservationService)
		now := time.Now()
		expectedReservation := &reservation.Reservation{
			ID:             "res-123",
			EventID:        "event-123",
			UserID:         "user-123",
			SeatIDs:        []string{"seat-1", "seat-2"},
			Status:         reservation.StatusPending,
			TotalAmount:    10000,
			IdempotencyKey: "idem-key",
			ExpiresAt:      now.Add(15 * time.Minute),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		mockService.On("CreateReservation", mock.Anything, mock.AnythingOfType("application.CreateReservationInput")).
			Return(expectedReservation, nil)

		handler := NewReservationHandler(mockService)

		reqBody := `{
			"event_id": "event-123",
			"seat_ids": ["seat-1", "seat-2"],
			"idempotency_key": "idem-key"
		}`
		req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp ReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "res-123", resp.ID)
		assert.Equal(t, "pending", resp.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("ユーザーIDがない場合401", func(t *testing.T) {
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		reqBody := `{"event_id": "event-123", "seat_ids": ["seat-1"], "idempotency_key": "idem-key"}`
		req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// X-User-ID ヘッダーなし
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("不正なリクエストでエラー", func(t *testing.T) {
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations", strings.NewReader("invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
	})
}

func TestReservationHandler_GetByID(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常に予約を取得できる", func(t *testing.T) {
		mockService := new(MockReservationService)
		now := time.Now()
		expectedReservation := &reservation.Reservation{
			ID:          "res-123",
			EventID:     "event-123",
			UserID:      "user-123",
			SeatIDs:     []string{"seat-1"},
			Status:      reservation.StatusPending,
			TotalAmount: 5000,
			ExpiresAt:   now.Add(15 * time.Minute),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("GetReservation", mock.Anything, "res-123").Return(expectedReservation, nil)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reservations/res-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("res-123")

		err := handler.GetByID(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("予約が見つからない場合404", func(t *testing.T) {
		mockService := new(MockReservationService)
		mockService.On("GetReservation", mock.Anything, "nonexistent").Return(nil, reservation.ErrReservationNotFound)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reservations/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := handler.GetByID(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)

		mockService.AssertExpectations(t)
	})
}

func TestReservationHandler_GetUserReservations(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にユーザーの予約一覧を取得できる", func(t *testing.T) {
		mockService := new(MockReservationService)
		now := time.Now()
		reservations := []*reservation.Reservation{
			{ID: "res-1", EventID: "event-1", UserID: "user-123", SeatIDs: []string{"seat-1"}, Status: reservation.StatusPending, ExpiresAt: now.Add(15 * time.Minute), CreatedAt: now, UpdatedAt: now},
			{ID: "res-2", EventID: "event-2", UserID: "user-123", SeatIDs: []string{"seat-2"}, Status: reservation.StatusConfirmed, ExpiresAt: now.Add(15 * time.Minute), CreatedAt: now, UpdatedAt: now},
		}

		mockService.On("GetUserReservations", mock.Anything, "user-123", 0, 0).Return(reservations, nil)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reservations", nil)
		req.Header.Set("X-User-ID", "user-123")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.GetUserReservations(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []ReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("ユーザーIDがない場合401", func(t *testing.T) {
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/reservations", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.GetUserReservations(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})
}

func TestReservationHandler_Confirm(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常に予約を確定できる", func(t *testing.T) {
		mockService := new(MockReservationService)
		now := time.Now()
		confirmedAt := now
		expectedReservation := &reservation.Reservation{
			ID:          "res-123",
			EventID:     "event-123",
			UserID:      "user-123",
			SeatIDs:     []string{"seat-1"},
			Status:      reservation.StatusConfirmed,
			TotalAmount: 5000,
			ExpiresAt:   now.Add(15 * time.Minute),
			ConfirmedAt: &confirmedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("ConfirmReservation", mock.Anything, "res-123").Return(expectedReservation, nil)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations/res-123/confirm", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("res-123")

		err := handler.Confirm(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp ReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "confirmed", resp.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("予約が見つからない場合404", func(t *testing.T) {
		mockService := new(MockReservationService)
		mockService.On("ConfirmReservation", mock.Anything, "nonexistent").Return(nil, reservation.ErrReservationNotFound)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations/nonexistent/confirm", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := handler.Confirm(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("確定できない状態の場合400", func(t *testing.T) {
		mockService := new(MockReservationService)
		mockService.On("ConfirmReservation", mock.Anything, "res-123").Return(nil, errors.New("予約がpending状態ではありません"))

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations/res-123/confirm", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("res-123")

		err := handler.Confirm(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)

		mockService.AssertExpectations(t)
	})
}

func TestReservationHandler_Cancel(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常に予約をキャンセルできる", func(t *testing.T) {
		mockService := new(MockReservationService)
		now := time.Now()
		expectedReservation := &reservation.Reservation{
			ID:          "res-123",
			EventID:     "event-123",
			UserID:      "user-123",
			SeatIDs:     []string{"seat-1"},
			Status:      reservation.StatusCancelled,
			TotalAmount: 5000,
			ExpiresAt:   now.Add(15 * time.Minute),
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("CancelReservation", mock.Anything, "res-123").Return(expectedReservation, nil)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations/res-123/cancel", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("res-123")

		err := handler.Cancel(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp ReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", resp.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("予約が見つからない場合404", func(t *testing.T) {
		mockService := new(MockReservationService)
		mockService.On("CancelReservation", mock.Anything, "nonexistent").Return(nil, reservation.ErrReservationNotFound)

		handler := NewReservationHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/reservations/nonexistent/cancel", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := handler.Cancel(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)

		mockService.AssertExpectations(t)
	})
}
