package handler

import (
	"context"
	"encoding/json"
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
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

// MockSeatService はSeatServiceInterfaceのモック
type MockSeatService struct {
	mock.Mock
}

func (m *MockSeatService) CreateSeat(ctx context.Context, input application.CreateSeatInput) (*seat.Seat, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*seat.Seat), args.Error(1)
}

func (m *MockSeatService) CreateBulkSeats(ctx context.Context, input application.CreateBulkSeatsInput) ([]*seat.Seat, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatService) GetSeat(ctx context.Context, id string) (*seat.Seat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*seat.Seat), args.Error(1)
}

func (m *MockSeatService) GetSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatService) GetAvailableSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatService) CountAvailableSeats(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func TestSeatHandler_GetByEvent(t *testing.T) {
	e := echo.New()

	t.Run("全座席を取得できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		now := time.Now()
		seats := []*seat.Seat{
			{ID: "seat-1", EventID: "event-123", SeatNumber: "A-1", Status: seat.StatusAvailable, Price: 5000, CreatedAt: now, UpdatedAt: now},
			{ID: "seat-2", EventID: "event-123", SeatNumber: "A-2", Status: seat.StatusReserved, Price: 5000, CreatedAt: now, UpdatedAt: now},
		}

		mockService.On("GetSeatsByEvent", mock.Anything, "event-123").Return(seats, nil)

		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events/event-123/seats", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")

		err := handler.GetByEvent(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []SeatResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("空席のみを取得できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		now := time.Now()
		seats := []*seat.Seat{
			{ID: "seat-1", EventID: "event-123", SeatNumber: "A-1", Status: seat.StatusAvailable, Price: 5000, CreatedAt: now, UpdatedAt: now},
		}

		mockService.On("GetAvailableSeatsByEvent", mock.Anything, "event-123").Return(seats, nil)

		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events/event-123/seats?available=true", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")
		c.QueryParams().Set("available", "true")

		err := handler.GetByEvent(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []SeatResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "available", resp[0].Status)

		mockService.AssertExpectations(t)
	})
}

func TestSeatHandler_Create(t *testing.T) {
	e := echo.New()

	t.Run("正常に座席を作成できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		now := time.Now()
		expectedSeat := &seat.Seat{
			ID:         "seat-123",
			EventID:    "event-123",
			SeatNumber: "A-1",
			Status:     seat.StatusAvailable,
			Price:      5000,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mockService.On("CreateSeat", mock.Anything, mock.AnythingOfType("application.CreateSeatInput")).
			Return(expectedSeat, nil)

		handler := NewSeatHandler(mockService)

		reqBody := `{"seat_number": "A-1", "price": 5000}`
		req := httptest.NewRequest(http.MethodPost, "/events/event-123/seats", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")

		err := handler.Create(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp SeatResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "seat-123", resp.ID)
		assert.Equal(t, "A-1", resp.SeatNumber)

		mockService.AssertExpectations(t)
	})

	t.Run("不正なリクエストでエラー", func(t *testing.T) {
		mockService := new(MockSeatService)
		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/events/event-123/seats", strings.NewReader("invalid"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")

		err := handler.Create(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestSeatHandler_CreateBulk(t *testing.T) {
	e := echo.New()

	t.Run("正常に一括作成できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		now := time.Now()
		seats := []*seat.Seat{
			{ID: "seat-1", EventID: "event-123", SeatNumber: "A-1", Status: seat.StatusAvailable, Price: 5000, CreatedAt: now, UpdatedAt: now},
			{ID: "seat-2", EventID: "event-123", SeatNumber: "A-2", Status: seat.StatusAvailable, Price: 5000, CreatedAt: now, UpdatedAt: now},
		}

		mockService.On("CreateBulkSeats", mock.Anything, mock.AnythingOfType("application.CreateBulkSeatsInput")).
			Return(seats, nil)

		handler := NewSeatHandler(mockService)

		reqBody := `{"prefix": "A", "count": 2, "price": 5000}`
		req := httptest.NewRequest(http.MethodPost, "/events/event-123/seats/bulk", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")

		err := handler.CreateBulk(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp []SeatResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockService.AssertExpectations(t)
	})
}

func TestSeatHandler_GetByID(t *testing.T) {
	e := echo.New()

	t.Run("正常に座席を取得できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		now := time.Now()
		expectedSeat := &seat.Seat{
			ID:         "seat-123",
			EventID:    "event-123",
			SeatNumber: "A-1",
			Status:     seat.StatusAvailable,
			Price:      5000,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mockService.On("GetSeat", mock.Anything, "seat-123").Return(expectedSeat, nil)

		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/seats/seat-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("seat-123")

		err := handler.GetByID(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("座席が見つからない場合404", func(t *testing.T) {
		mockService := new(MockSeatService)
		mockService.On("GetSeat", mock.Anything, "nonexistent").Return(nil, seat.ErrSeatNotFound)

		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/seats/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := handler.GetByID(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestSeatHandler_CountAvailable(t *testing.T) {
	e := echo.New()

	t.Run("正常に空席数を取得できる", func(t *testing.T) {
		mockService := new(MockSeatService)
		mockService.On("CountAvailableSeats", mock.Anything, "event-123").Return(50, nil)

		handler := NewSeatHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events/event-123/seats/count", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("event_id")
		c.SetParamValues("event-123")

		err := handler.CountAvailable(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "50")

		mockService.AssertExpectations(t)
	})
}
