package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
)

// MockEventRepository はevent.Repositoryのモック
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id string) (*event.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventRepository) List(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventRepository) Update(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewEventService(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)
	assert.NotNil(t, service)
}

func TestEventService_CreateEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	input := CreateEventInput{
		Name:        "テストイベント",
		Description: "テスト説明",
		Venue:       "テスト会場",
		StartAt:     time.Now().Add(24 * time.Hour),
		EndAt:       time.Now().Add(27 * time.Hour),
		TotalSeats:  100,
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)

	result, err := service.CreateEvent(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Description, result.Description)
	assert.Equal(t, input.Venue, result.Venue)
	assert.Equal(t, input.TotalSeats, result.TotalSeats)
	mockRepo.AssertExpectations(t)
}

func TestEventService_CreateEvent_ValidationError(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	// 無効な入力（名前が空）
	input := CreateEventInput{
		Name:        "",
		Description: "テスト説明",
		Venue:       "テスト会場",
		StartAt:     time.Now().Add(24 * time.Hour),
		EndAt:       time.Now().Add(27 * time.Hour),
		TotalSeats:  100,
	}

	result, err := service.CreateEvent(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "バリデーションエラー")
	// CreateはValidationエラーで失敗するので呼ばれない
	mockRepo.AssertNotCalled(t, "Create")
}

func TestEventService_CreateEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	input := CreateEventInput{
		Name:        "テストイベント",
		Description: "テスト説明",
		Venue:       "テスト会場",
		StartAt:     time.Now().Add(24 * time.Hour),
		EndAt:       time.Now().Add(27 * time.Hour),
		TotalSeats:  100,
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*event.Event")).
		Return(errors.New("データベースエラー"))

	result, err := service.CreateEvent(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "イベント作成に失敗しました")
	mockRepo.AssertExpectations(t)
}

func TestEventService_GetEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	expectedEvent := &event.Event{
		ID:   "event-1",
		Name: "テストイベント",
	}

	mockRepo.On("GetByID", mock.Anything, "event-1").Return(expectedEvent, nil)

	result, err := service.GetEvent(context.Background(), "event-1")

	require.NoError(t, err)
	assert.Equal(t, expectedEvent, result)
	mockRepo.AssertExpectations(t)
}

func TestEventService_GetEvent_NotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, event.ErrEventNotFound)

	result, err := service.GetEvent(context.Background(), "non-existent")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, event.ErrEventNotFound)
	mockRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	expectedEvents := []*event.Event{
		{ID: "event-1", Name: "イベント1"},
		{ID: "event-2", Name: "イベント2"},
	}

	mockRepo.On("List", mock.Anything, 20, 0).Return(expectedEvents, nil)

	result, err := service.ListEvents(context.Background(), 0, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_WithLimitAndOffset(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	expectedEvents := []*event.Event{
		{ID: "event-3", Name: "イベント3"},
	}

	mockRepo.On("List", mock.Anything, 10, 20).Return(expectedEvents, nil)

	result, err := service.ListEvents(context.Background(), 10, 20)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_LimitCapped(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	mockRepo.On("List", mock.Anything, 100, 0).Return([]*event.Event{}, nil)

	// limit が 100 を超えると 100 に制限される
	_, err := service.ListEvents(context.Background(), 200, 0)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_NegativeOffset(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	mockRepo.On("List", mock.Anything, 20, 0).Return([]*event.Event{}, nil)

	// 負のoffsetは0に補正される
	_, err := service.ListEvents(context.Background(), 0, -10)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	existingEvent := &event.Event{
		ID:          "event-1",
		Name:        "旧イベント名",
		Description: "旧説明",
		Venue:       "旧会場",
		StartAt:     time.Now().Add(24 * time.Hour),
		EndAt:       time.Now().Add(27 * time.Hour),
		TotalSeats:  50,
	}

	input := UpdateEventInput{
		ID:          "event-1",
		Name:        "新イベント名",
		Description: "新説明",
		Venue:       "新会場",
		StartAt:     time.Now().Add(48 * time.Hour),
		EndAt:       time.Now().Add(51 * time.Hour),
		TotalSeats:  100,
	}

	mockRepo.On("GetByID", mock.Anything, "event-1").Return(existingEvent, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*event.Event")).Return(nil)

	result, err := service.UpdateEvent(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Description, result.Description)
	assert.Equal(t, input.Venue, result.Venue)
	assert.Equal(t, input.TotalSeats, result.TotalSeats)
	mockRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	input := UpdateEventInput{
		ID:         "non-existent",
		Name:       "新イベント名",
		TotalSeats: 100,
	}

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, event.ErrEventNotFound)

	result, err := service.UpdateEvent(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, event.ErrEventNotFound)
	mockRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_ValidationError(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	existingEvent := &event.Event{
		ID:         "event-1",
		Name:       "旧イベント名",
		TotalSeats: 50,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(27 * time.Hour),
	}

	// 無効な入力（名前が空）
	input := UpdateEventInput{
		ID:         "event-1",
		Name:       "",
		TotalSeats: 100,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(27 * time.Hour),
	}

	mockRepo.On("GetByID", mock.Anything, "event-1").Return(existingEvent, nil)

	result, err := service.UpdateEvent(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "バリデーションエラー")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestEventService_DeleteEvent_Success(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	mockRepo.On("Delete", mock.Anything, "event-1").Return(nil)

	err := service.DeleteEvent(context.Background(), "event-1")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEventService_DeleteEvent_NotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	service := NewEventService(mockRepo)

	mockRepo.On("Delete", mock.Anything, "non-existent").Return(event.ErrEventNotFound)

	err := service.DeleteEvent(context.Background(), "non-existent")

	require.Error(t, err)
	assert.ErrorIs(t, err, event.ErrEventNotFound)
	mockRepo.AssertExpectations(t)
}
