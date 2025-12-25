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
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
)

// MockEventService はEventServiceInterfaceのモック
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) CreateEvent(ctx context.Context, input application.CreateEventInput) (*event.Event, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventService) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventService) ListEvents(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventService) UpdateEvent(ctx context.Context, input application.UpdateEventInput) (*event.Event, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventService) DeleteEvent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestEventHandler_Create(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にイベントを作成できる", func(t *testing.T) {
		mockService := new(MockEventService)
		now := time.Now()
		expectedEvent := &event.Event{
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

		mockService.On("CreateEvent", mock.Anything, mock.AnythingOfType("application.CreateEventInput")).
			Return(expectedEvent, nil)

		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"description": "テスト説明",
			"venue": "テスト会場",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp EventResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "event-123", resp.ID)
		assert.Equal(t, "テストイベント", resp.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("不正なリクエスト形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
	})

	t.Run("不正な開始時刻形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "invalid-date",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "開始時刻")
	})

	t.Run("不正な終了時刻形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "invalid-date",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "終了時刻")
	})
}

func TestEventHandler_GetByID(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にイベントを取得できる", func(t *testing.T) {
		mockService := new(MockEventService)
		now := time.Now()
		expectedEvent := &event.Event{
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

		mockService.On("GetEvent", mock.Anything, "event-123").Return(expectedEvent, nil)

		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events/event-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.GetByID(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp EventResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "event-123", resp.ID)

		mockService.AssertExpectations(t)
	})

	t.Run("イベントが見つからない場合404", func(t *testing.T) {
		mockService := new(MockEventService)
		mockService.On("GetEvent", mock.Anything, "nonexistent").Return(nil, event.ErrEventNotFound)

		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events/nonexistent", nil)
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

func TestEventHandler_List(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にイベント一覧を取得できる", func(t *testing.T) {
		mockService := new(MockEventService)
		now := time.Now()
		events := []*event.Event{
			{ID: "event-1", Name: "イベント1", StartAt: now, EndAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
			{ID: "event-2", Name: "イベント2", StartAt: now, EndAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
		}

		mockService.On("ListEvents", mock.Anything, 0, 0).Return(events, nil)

		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.List(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []*EventResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)

		mockService.AssertExpectations(t)
	})
}

func TestEventHandler_Delete(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にイベントを削除できる", func(t *testing.T) {
		mockService := new(MockEventService)
		mockService.On("DeleteEvent", mock.Anything, "event-123").Return(nil)

		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodDelete, "/events/event-123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Delete(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("イベントが見つからない場合404", func(t *testing.T) {
		mockService := new(MockEventService)
		mockService.On("DeleteEvent", mock.Anything, "nonexistent").Return(event.ErrEventNotFound)

		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodDelete, "/events/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent")

		err := handler.Delete(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)

		mockService.AssertExpectations(t)
	})
}

func TestEventHandler_Update(t *testing.T) {
	e := NewTestEcho()

	t.Run("正常にイベントを更新できる", func(t *testing.T) {
		mockService := new(MockEventService)
		now := time.Now()
		expectedEvent := &event.Event{
			ID:          "event-123",
			Name:        "更新されたイベント",
			Description: "更新された説明",
			Venue:       "更新された会場",
			StartAt:     now,
			EndAt:       now.Add(3 * time.Hour),
			TotalSeats:  200,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockService.On("UpdateEvent", mock.Anything, mock.AnythingOfType("application.UpdateEventInput")).
			Return(expectedEvent, nil)

		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "更新されたイベント",
			"description": "更新された説明",
			"venue": "更新された会場",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 200
		}`
		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp EventResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "更新されたイベント", resp.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("不正なリクエスト形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
	})

	t.Run("不正な開始時刻形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "invalid-date",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "開始時刻")
	})

	t.Run("不正な終了時刻形式でエラー", func(t *testing.T) {
		mockService := new(MockEventService)
		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "invalid-date",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "終了時刻")
	})

	t.Run("イベントが見つからない場合404", func(t *testing.T) {
		mockService := new(MockEventService)
		mockService.On("UpdateEvent", mock.Anything, mock.AnythingOfType("application.UpdateEventInput")).
			Return(nil, event.ErrEventNotFound)

		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, he.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("その他のエラーで400", func(t *testing.T) {
		mockService := new(MockEventService)
		mockService.On("UpdateEvent", mock.Anything, mock.AnythingOfType("application.UpdateEventInput")).
			Return(nil, assert.AnError)

		handler := NewEventHandler(mockService)

		reqBody := `{
			"name": "テストイベント",
			"start_at": "2025-12-31T18:00:00+09:00",
			"end_at": "2025-12-31T21:00:00+09:00",
			"total_seats": 100
		}`
		req := httptest.NewRequest(http.MethodPut, "/events/event-123", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("event-123")

		err := handler.Update(c)

		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)

		mockService.AssertExpectations(t)
	})
}
