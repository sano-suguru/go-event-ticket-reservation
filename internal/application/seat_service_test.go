package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/transaction"
)

// MockSeatRepository is a mock implementation of seat.Repository
type MockSeatRepository struct {
	mock.Mock
}

func (m *MockSeatRepository) Create(ctx context.Context, s *seat.Seat) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSeatRepository) CreateBulk(ctx context.Context, seats []*seat.Seat) error {
	args := m.Called(ctx, seats)
	return args.Error(0)
}

func (m *MockSeatRepository) GetByID(ctx context.Context, id string) (*seat.Seat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*seat.Seat), args.Error(1)
}

func (m *MockSeatRepository) GetByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatRepository) GetAvailableByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatRepository) CountAvailableByEventID(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *MockSeatRepository) ReserveSeats(ctx context.Context, tx transaction.Tx, ids []string, reservationID string) error {
	args := m.Called(ctx, tx, ids, reservationID)
	return args.Error(0)
}

func (m *MockSeatRepository) ConfirmSeats(ctx context.Context, tx transaction.Tx, ids []string) error {
	args := m.Called(ctx, tx, ids)
	return args.Error(0)
}

func (m *MockSeatRepository) ReleaseSeats(ctx context.Context, tx transaction.Tx, ids []string) error {
	args := m.Called(ctx, tx, ids)
	return args.Error(0)
}

func TestNewSeatService(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockEventRepo := new(MockEventRepository)
	mockCache := new(MockSeatCache)

	service := NewSeatService(mockSeatRepo, mockEventRepo, mockCache)

	assert.NotNil(t, service)
}

func TestSeatService_CreateSeat(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateSeatInput
		setupMocks  func(sr *MockSeatRepository, er *MockEventRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "正常に座席が作成される",
			input: CreateSeatInput{
				EventID:    "event-123",
				SeatNumber: "A-1",
				Price:      5000,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "event-123").Return(&event.Event{ID: "event-123"}, nil)
				sr.On("Create", mock.Anything, mock.AnythingOfType("*seat.Seat")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "イベントが存在しない場合エラー",
			input: CreateSeatInput{
				EventID:    "nonexistent",
				SeatNumber: "A-1",
				Price:      5000,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "nonexistent").Return(nil, event.ErrEventNotFound)
			},
			expectError: true,
			errorMsg:    "イベント取得に失敗",
		},
		{
			name: "バリデーションエラー - 価格が負",
			input: CreateSeatInput{
				EventID:    "event-123",
				SeatNumber: "A-1",
				Price:      -1,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "event-123").Return(&event.Event{ID: "event-123"}, nil)
			},
			expectError: true,
		},
		{
			name: "リポジトリエラー",
			input: CreateSeatInput{
				EventID:    "event-123",
				SeatNumber: "A-1",
				Price:      5000,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "event-123").Return(&event.Event{ID: "event-123"}, nil)
				sr.On("Create", mock.Anything, mock.AnythingOfType("*seat.Seat")).Return(errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSeatRepo := new(MockSeatRepository)
			mockEventRepo := new(MockEventRepository)
			tt.setupMocks(mockSeatRepo, mockEventRepo)

			// SeatServiceはcacheなしで作成（nilで問題ない）
			service := &SeatService{
				seatRepo:  mockSeatRepo,
				eventRepo: mockEventRepo,
				cache:     nil,
			}

			result, err := service.CreateSeat(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.EventID, result.EventID)
				assert.Equal(t, tt.input.SeatNumber, result.SeatNumber)
				assert.Equal(t, tt.input.Price, result.Price)
			}

			mockSeatRepo.AssertExpectations(t)
			mockEventRepo.AssertExpectations(t)
		})
	}
}

func TestSeatService_CreateBulkSeats(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateBulkSeatsInput
		setupMocks  func(sr *MockSeatRepository, er *MockEventRepository)
		expectError bool
		expectCount int
	}{
		{
			name: "正常に一括作成される",
			input: CreateBulkSeatsInput{
				EventID: "event-123",
				Prefix:  "A",
				Count:   3,
				Price:   5000,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "event-123").Return(&event.Event{ID: "event-123"}, nil)
				sr.On("CreateBulk", mock.Anything, mock.AnythingOfType("[]*seat.Seat")).Return(nil)
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name: "イベントが存在しない",
			input: CreateBulkSeatsInput{
				EventID: "nonexistent",
				Prefix:  "A",
				Count:   5,
				Price:   5000,
			},
			setupMocks: func(sr *MockSeatRepository, er *MockEventRepository) {
				er.On("GetByID", mock.Anything, "nonexistent").Return(nil, event.ErrEventNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSeatRepo := new(MockSeatRepository)
			mockEventRepo := new(MockEventRepository)
			tt.setupMocks(mockSeatRepo, mockEventRepo)

			service := &SeatService{
				seatRepo:  mockSeatRepo,
				eventRepo: mockEventRepo,
				cache:     nil,
			}

			result, err := service.CreateBulkSeats(context.Background(), tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectCount)
				for _, s := range result {
					assert.Contains(t, s.SeatNumber, tt.input.Prefix)
					assert.Equal(t, tt.input.Price, s.Price)
				}
			}

			mockSeatRepo.AssertExpectations(t)
			mockEventRepo.AssertExpectations(t)
		})
	}
}

func TestSeatService_GetSeat(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	expectedSeat := &seat.Seat{ID: "seat-123", EventID: "event-123", SeatNumber: "A-1"}
	mockSeatRepo.On("GetByID", mock.Anything, "seat-123").Return(expectedSeat, nil)

	service := &SeatService{seatRepo: mockSeatRepo}

	result, err := service.GetSeat(context.Background(), "seat-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedSeat, result)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_GetSeat_NotFound(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockSeatRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, seat.ErrSeatNotFound)

	service := &SeatService{seatRepo: mockSeatRepo}

	result, err := service.GetSeat(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, seat.ErrSeatNotFound)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_GetSeatsByEvent(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	expectedSeats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-123"},
		{ID: "seat-2", EventID: "event-123"},
	}
	mockSeatRepo.On("GetByEventID", mock.Anything, "event-123").Return(expectedSeats, nil)

	service := &SeatService{seatRepo: mockSeatRepo}

	result, err := service.GetSeatsByEvent(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_GetAvailableSeatsByEvent(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	expectedSeats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-123", Status: seat.StatusAvailable},
	}
	mockSeatRepo.On("GetAvailableByEventID", mock.Anything, "event-123").Return(expectedSeats, nil)

	service := &SeatService{seatRepo: mockSeatRepo}

	result, err := service.GetAvailableSeatsByEvent(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, seat.StatusAvailable, result[0].Status)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_CountAvailableSeats_NoCacheHit(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockSeatRepo.On("CountAvailableByEventID", mock.Anything, "event-123").Return(50, nil)

	service := &SeatService{
		seatRepo: mockSeatRepo,
		cache:    nil, // キャッシュなし
	}

	count, err := service.CountAvailableSeats(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Equal(t, 50, count)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_CountAvailableSeats_DBError(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockSeatRepo.On("CountAvailableByEventID", mock.Anything, "event-123").Return(0, errors.New("db error"))

	service := &SeatService{
		seatRepo: mockSeatRepo,
		cache:    nil,
	}

	count, err := service.CountAvailableSeats(context.Background(), "event-123")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	mockSeatRepo.AssertExpectations(t)
}

// MockSeatCache はSeatCacheInterfaceのモック
type MockSeatCache struct {
	mock.Mock
}

func (m *MockSeatCache) GetAvailableCount(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *MockSeatCache) SetAvailableCount(ctx context.Context, eventID string, count int, ttl time.Duration) error {
	args := m.Called(ctx, eventID, count, ttl)
	return args.Error(0)
}

func (m *MockSeatCache) Invalidate(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func TestSeatService_CountAvailableSeats_CacheHit(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockCache := new(MockSeatCache)
	mockCache.On("GetAvailableCount", mock.Anything, "event-123").Return(42, nil)

	service := &SeatService{
		seatRepo: mockSeatRepo,
		cache:    mockCache,
	}

	count, err := service.CountAvailableSeats(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Equal(t, 42, count)
	// DBは呼ばれない
	mockSeatRepo.AssertNotCalled(t, "CountAvailableByEventID")
	mockCache.AssertExpectations(t)
}

func TestSeatService_CountAvailableSeats_CacheMiss(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockCache := new(MockSeatCache)
	mockCache.On("GetAvailableCount", mock.Anything, "event-123").Return(0, errors.New("cache miss"))
	mockSeatRepo.On("CountAvailableByEventID", mock.Anything, "event-123").Return(50, nil)
	mockCache.On("SetAvailableCount", mock.Anything, "event-123", 50, mock.Anything).Return(nil)

	service := &SeatService{
		seatRepo: mockSeatRepo,
		cache:    mockCache,
	}

	count, err := service.CountAvailableSeats(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Equal(t, 50, count)
	mockCache.AssertExpectations(t)
	mockSeatRepo.AssertExpectations(t)
}

func TestSeatService_InvalidateCache(t *testing.T) {
	mockCache := new(MockSeatCache)
	mockCache.On("Invalidate", mock.Anything, "event-123").Return(nil)

	service := &SeatService{
		cache: mockCache,
	}

	service.InvalidateCache(context.Background(), "event-123")

	mockCache.AssertExpectations(t)
}

func TestSeatService_InvalidateCache_NoCache(t *testing.T) {
	service := &SeatService{
		cache: nil,
	}

	// パニックしないことを確認
	service.InvalidateCache(context.Background(), "event-123")
}

func TestSeatService_InvalidateCache_Error(t *testing.T) {
	mockCache := new(MockSeatCache)
	mockCache.On("Invalidate", mock.Anything, "event-123").Return(errors.New("cache error"))

	service := &SeatService{
		cache: mockCache,
	}

	// エラーが発生してもパニックしないことを確認
	service.InvalidateCache(context.Background(), "event-123")

	mockCache.AssertExpectations(t)
}

func TestSeatService_CountAvailableSeats_CacheSetError(t *testing.T) {
	mockSeatRepo := new(MockSeatRepository)
	mockCache := new(MockSeatCache)
	mockCache.On("GetAvailableCount", mock.Anything, "event-123").Return(0, errors.New("cache miss"))
	mockSeatRepo.On("CountAvailableByEventID", mock.Anything, "event-123").Return(50, nil)
	mockCache.On("SetAvailableCount", mock.Anything, "event-123", 50, mock.Anything).Return(errors.New("cache set error"))

	service := &SeatService{
		seatRepo: mockSeatRepo,
		cache:    mockCache,
	}

	// キャッシュ保存エラーでも正常な結果が返ることを確認
	count, err := service.CountAvailableSeats(context.Background(), "event-123")

	assert.NoError(t, err)
	assert.Equal(t, 50, count)
	mockCache.AssertExpectations(t)
	mockSeatRepo.AssertExpectations(t)
}
