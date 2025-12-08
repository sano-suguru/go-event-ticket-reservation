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
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/transaction"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

// === Mock implementations ===

// MockTxManager implements transaction.Manager
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) Begin(ctx context.Context) (transaction.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(transaction.Tx), args.Error(1)
}

// MockTx implements transaction.Tx
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockReservationRepository implements reservation.Repository
type MockReservationRepository struct {
	mock.Mock
}

func (m *MockReservationRepository) Create(ctx context.Context, tx transaction.Tx, r *reservation.Reservation) error {
	args := m.Called(ctx, tx, r)
	return args.Error(0)
}

func (m *MockReservationRepository) GetByID(ctx context.Context, id string) (*reservation.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*reservation.Reservation, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reservation.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetByIdempotencyKey(ctx context.Context, key string) (*reservation.Reservation, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reservation.Reservation), args.Error(1)
}

func (m *MockReservationRepository) Update(ctx context.Context, tx transaction.Tx, r *reservation.Reservation) error {
	args := m.Called(ctx, tx, r)
	return args.Error(0)
}

func (m *MockReservationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReservationRepository) GetExpiredPending(ctx context.Context, olderThan time.Duration) ([]*reservation.Reservation, error) {
	args := m.Called(ctx, olderThan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reservation.Reservation), args.Error(1)
}

// MockSeatRepositoryUnit implements seat.Repository for unit tests
type MockSeatRepositoryUnit struct {
	mock.Mock
}

func (m *MockSeatRepositoryUnit) Create(ctx context.Context, s *seat.Seat) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSeatRepositoryUnit) CreateBulk(ctx context.Context, seats []*seat.Seat) error {
	args := m.Called(ctx, seats)
	return args.Error(0)
}

func (m *MockSeatRepositoryUnit) GetByID(ctx context.Context, id string) (*seat.Seat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*seat.Seat), args.Error(1)
}

func (m *MockSeatRepositoryUnit) GetByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatRepositoryUnit) GetAvailableByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*seat.Seat), args.Error(1)
}

func (m *MockSeatRepositoryUnit) CountAvailableByEventID(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *MockSeatRepositoryUnit) ReserveSeats(ctx context.Context, tx transaction.Tx, ids []string, reservationID string) error {
	args := m.Called(ctx, tx, ids, reservationID)
	return args.Error(0)
}

func (m *MockSeatRepositoryUnit) ConfirmSeats(ctx context.Context, tx transaction.Tx, ids []string) error {
	args := m.Called(ctx, tx, ids)
	return args.Error(0)
}

func (m *MockSeatRepositoryUnit) ReleaseSeats(ctx context.Context, tx transaction.Tx, ids []string) error {
	args := m.Called(ctx, tx, ids)
	return args.Error(0)
}

// MockEventRepositoryUnit implements event.Repository for unit tests
type MockEventRepositoryUnit struct {
	mock.Mock
}

func (m *MockEventRepositoryUnit) Create(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockEventRepositoryUnit) GetByID(ctx context.Context, id string) (*event.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockEventRepositoryUnit) List(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*event.Event), args.Error(1)
}

func (m *MockEventRepositoryUnit) Update(ctx context.Context, e *event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockEventRepositoryUnit) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockLockManager implements redisinfra.LockManagerInterface
type MockLockManager struct {
	mock.Mock
}

func (m *MockLockManager) AcquireLock(ctx context.Context, key string, ttl time.Duration) (redisinfra.Lock, error) {
	args := m.Called(ctx, key, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(redisinfra.Lock), args.Error(1)
}

func (m *MockLockManager) AcquireLockWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (redisinfra.Lock, error) {
	args := m.Called(ctx, key, ttl, maxRetries, retryInterval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(redisinfra.Lock), args.Error(1)
}

// MockLock implements redisinfra.Lock
type MockLock struct {
	mock.Mock
}

func (m *MockLock) Release(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLock) Extend(ctx context.Context, ttl time.Duration) error {
	args := m.Called(ctx, ttl)
	return args.Error(0)
}

// MockSeatCacheUnit implements redisinfra.SeatCacheInterface
type MockSeatCacheUnit struct {
	mock.Mock
}

func (m *MockSeatCacheUnit) GetAvailableCount(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *MockSeatCacheUnit) SetAvailableCount(ctx context.Context, eventID string, count int, ttl time.Duration) error {
	args := m.Called(ctx, eventID, count, ttl)
	return args.Error(0)
}

func (m *MockSeatCacheUnit) Invalidate(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// === Test helper ===
type testDeps struct {
	txManager   *MockTxManager
	tx          *MockTx
	resRepo     *MockReservationRepository
	seatRepo    *MockSeatRepositoryUnit
	eventRepo   *MockEventRepositoryUnit
	lockManager *MockLockManager
	lock        *MockLock
	seatCache   *MockSeatCacheUnit
	service     *ReservationService
}

func newTestDeps() *testDeps {
	txm := new(MockTxManager)
	tx := new(MockTx)
	resRepo := new(MockReservationRepository)
	seatRepo := new(MockSeatRepositoryUnit)
	eventRepo := new(MockEventRepositoryUnit)
	lockManager := new(MockLockManager)
	lock := new(MockLock)
	seatCache := new(MockSeatCacheUnit)

	service := NewReservationService(txm, resRepo, seatRepo, eventRepo, lockManager, seatCache)

	return &testDeps{
		txManager:   txm,
		tx:          tx,
		resRepo:     resRepo,
		seatRepo:    seatRepo,
		eventRepo:   eventRepo,
		lockManager: lockManager,
		lock:        lock,
		seatCache:   seatCache,
		service:     service,
	}
}

// === Tests ===

func TestReservationService_CreateReservation_Success(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1", "seat-2"},
		IdempotencyKey: "idempotency-key-1",
	}

	// Setup mocks
	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	// IsBookingOpen() returns true when now.Before(StartAt)
	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour), // Future start = booking open
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
		{ID: "seat-2", EventID: "event-1", Status: seat.StatusAvailable, Price: 2000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)
	deps.tx.On("Commit").Return(nil)

	deps.resRepo.On("Create", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)
	deps.seatRepo.On("ReserveSeats", ctx, deps.tx, input.SeatIDs, mock.AnythingOfType("string")).Return(nil)

	deps.seatCache.On("Invalidate", ctx, "event-1").Return(nil)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "event-1", result.EventID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, 3000, result.TotalAmount)
	assert.Equal(t, reservation.StatusPending, result.Status)

	deps.txManager.AssertExpectations(t)
	deps.resRepo.AssertExpectations(t)
	deps.seatRepo.AssertExpectations(t)
	deps.eventRepo.AssertExpectations(t)
	deps.lockManager.AssertExpectations(t)
}

func TestReservationService_CreateReservation_IdempotencyHit(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "existing-key",
	}

	existingRes := &reservation.Reservation{
		ID:             "existing-res",
		EventID:        "event-1",
		UserID:         "user-1",
		IdempotencyKey: "existing-key",
		Status:         reservation.StatusPending,
	}
	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).Return(existingRes, nil)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "existing-res", result.ID)
	deps.resRepo.AssertExpectations(t)
	// No other mocks should be called
	deps.lockManager.AssertNotCalled(t, "AcquireLockWithRetry")
}

func TestReservationService_CreateReservation_LockFailed(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(nil, redisinfra.ErrLockNotAcquired)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "他のユーザーによって処理中")
}

func TestReservationService_CreateReservation_EventNotOpen(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	// Event with past start time (IsBookingOpen returns false)
	closedEvent := &event.Event{
		ID:      "event-1",
		Name:    "Past Event",
		StartAt: time.Now().Add(-1 * time.Hour), // Past start = booking closed
		EndAt:   time.Now().Add(1 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(closedEvent, nil)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, event.ErrEventNotOpen))
}

func TestReservationService_CreateReservation_SeatNotFound(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1", "nonexistent-seat"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	// Only one seat exists
	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, seat.ErrSeatNotFound))
}

func TestReservationService_CreateReservation_SeatAlreadyReserved(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	// Seat is already reserved
	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusReserved, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	// Execute
	result, err := deps.service.CreateReservation(ctx, input)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, seat.ErrSeatAlreadyReserved))
}

func TestReservationService_GetReservation(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	expected := &reservation.Reservation{
		ID:      "res-1",
		EventID: "event-1",
		UserID:  "user-1",
	}
	deps.resRepo.On("GetByID", ctx, "res-1").Return(expected, nil)

	result, err := deps.service.GetReservation(ctx, "res-1")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestReservationService_GetUserReservations(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	expected := []*reservation.Reservation{
		{ID: "res-1", UserID: "user-1"},
		{ID: "res-2", UserID: "user-1"},
	}
	deps.resRepo.On("GetByUserID", ctx, "user-1", 20, 0).Return(expected, nil)

	result, err := deps.service.GetUserReservations(ctx, "user-1", 0, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestReservationService_ConfirmReservation_Success(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	res := &reservation.Reservation{
		ID:        "res-1",
		EventID:   "event-1",
		UserID:    "user-1",
		SeatIDs:   []string{"seat-1"},
		Status:    reservation.StatusPending,
		ExpiresAt: time.Now().Add(10 * time.Minute), // Not expired
	}
	deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)
	deps.tx.On("Commit").Return(nil)
	deps.seatRepo.On("ConfirmSeats", ctx, deps.tx, res.SeatIDs).Return(nil)
	deps.resRepo.On("Update", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)

	result, err := deps.service.ConfirmReservation(ctx, "res-1")

	require.NoError(t, err)
	assert.Equal(t, reservation.StatusConfirmed, result.Status)
}

func TestReservationService_ConfirmReservation_NotFound(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	deps.resRepo.On("GetByID", ctx, "nonexistent").Return(nil, reservation.ErrReservationNotFound)

	result, err := deps.service.ConfirmReservation(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, reservation.ErrReservationNotFound))
}

func TestReservationService_CancelReservation_Success(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	res := &reservation.Reservation{
		ID:      "res-1",
		EventID: "event-1",
		UserID:  "user-1",
		SeatIDs: []string{"seat-1"},
		Status:  reservation.StatusPending,
	}
	deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)
	deps.tx.On("Commit").Return(nil)
	deps.seatRepo.On("ReleaseSeats", ctx, deps.tx, res.SeatIDs).Return(nil)
	deps.resRepo.On("Update", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)
	deps.seatCache.On("Invalidate", ctx, "event-1").Return(nil)

	result, err := deps.service.CancelReservation(ctx, "res-1")

	require.NoError(t, err)
	assert.Equal(t, reservation.StatusCancelled, result.Status)
}

func TestReservationService_CancelExpiredReservations(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	expired := []*reservation.Reservation{
		{
			ID:      "res-1",
			EventID: "event-1",
			UserID:  "user-1",
			SeatIDs: []string{"seat-1"},
			Status:  reservation.StatusPending,
		},
		{
			ID:      "res-2",
			EventID: "event-2",
			UserID:  "user-2",
			SeatIDs: []string{"seat-2"},
			Status:  reservation.StatusPending,
		},
	}
	expireAfter := 15 * time.Minute

	deps.resRepo.On("GetExpiredPending", ctx, expireAfter).Return(expired, nil)

	// First reservation succeeds
	tx1 := new(MockTx)
	deps.txManager.On("Begin", ctx).Return(tx1, nil).Once()
	tx1.On("Rollback").Return(nil)
	tx1.On("Commit").Return(nil)
	deps.seatRepo.On("ReleaseSeats", ctx, tx1, []string{"seat-1"}).Return(nil).Once()
	deps.resRepo.On("Update", ctx, tx1, mock.AnythingOfType("*reservation.Reservation")).Return(nil).Once()

	// Second reservation succeeds
	tx2 := new(MockTx)
	deps.txManager.On("Begin", ctx).Return(tx2, nil).Once()
	tx2.On("Rollback").Return(nil)
	tx2.On("Commit").Return(nil)
	deps.seatRepo.On("ReleaseSeats", ctx, tx2, []string{"seat-2"}).Return(nil).Once()
	deps.resRepo.On("Update", ctx, tx2, mock.AnythingOfType("*reservation.Reservation")).Return(nil).Once()

	count, err := deps.service.CancelExpiredReservations(ctx, expireAfter)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestReservationService_CancelExpiredReservations_Errors(t *testing.T) {
	t.Run("GetExpiredPending失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		deps.resRepo.On("GetExpiredPending", ctx, 15*time.Minute).Return(nil, errors.New("db error"))

		count, err := deps.service.CancelExpiredReservations(ctx, 15*time.Minute)

		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "期限切れ予約取得に失敗")
	})

	t.Run("一部の予約でエラー発生", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		expired := []*reservation.Reservation{
			{
				ID:      "res-1",
				EventID: "event-1",
				UserID:  "user-1",
				SeatIDs: []string{"seat-1"},
				Status:  reservation.StatusPending,
			},
			{
				ID:      "res-2",
				EventID: "event-2",
				UserID:  "user-2",
				SeatIDs: []string{"seat-2"},
				Status:  reservation.StatusPending,
			},
		}

		deps.resRepo.On("GetExpiredPending", ctx, 15*time.Minute).Return(expired, nil)

		// First reservation: Begin fails
		deps.txManager.On("Begin", ctx).Return(nil, errors.New("begin error")).Once()

		// Second reservation succeeds
		tx2 := new(MockTx)
		deps.txManager.On("Begin", ctx).Return(tx2, nil).Once()
		tx2.On("Rollback").Return(nil)
		tx2.On("Commit").Return(nil)
		deps.seatRepo.On("ReleaseSeats", ctx, tx2, []string{"seat-2"}).Return(nil).Once()
		deps.resRepo.On("Update", ctx, tx2, mock.AnythingOfType("*reservation.Reservation")).Return(nil).Once()

		count, err := deps.service.CancelExpiredReservations(ctx, 15*time.Minute)

		require.NoError(t, err)
		assert.Equal(t, 1, count) // Only one succeeded
	})

	t.Run("既にキャンセル済みの予約をスキップ", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		expired := []*reservation.Reservation{
			{
				ID:      "res-1",
				EventID: "event-1",
				UserID:  "user-1",
				SeatIDs: []string{"seat-1"},
				Status:  reservation.StatusCancelled, // Already cancelled
			},
		}

		deps.resRepo.On("GetExpiredPending", ctx, 15*time.Minute).Return(expired, nil)

		count, err := deps.service.CancelExpiredReservations(ctx, 15*time.Minute)

		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("コミット失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		expired := []*reservation.Reservation{
			{
				ID:      "res-1",
				EventID: "event-1",
				UserID:  "user-1",
				SeatIDs: []string{"seat-1"},
				Status:  reservation.StatusPending,
			},
		}

		deps.resRepo.On("GetExpiredPending", ctx, 15*time.Minute).Return(expired, nil)

		tx := new(MockTx)
		deps.txManager.On("Begin", ctx).Return(tx, nil)
		tx.On("Rollback").Return(nil)
		tx.On("Commit").Return(errors.New("commit error"))
		deps.seatRepo.On("ReleaseSeats", ctx, tx, []string{"seat-1"}).Return(nil)
		deps.resRepo.On("Update", ctx, tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)

		count, err := deps.service.CancelExpiredReservations(ctx, 15*time.Minute)

		require.NoError(t, err)
		assert.Equal(t, 0, count) // None succeeded due to commit failure
	})
}

func TestReservationService_CreateReservation_TransactionBeginFailed(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	deps.txManager.On("Begin", ctx).Return(nil, errors.New("db connection failed"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "トランザクション開始に失敗")
}

func TestReservationService_ConfirmReservation_TransactionErrors(t *testing.T) {
	t.Run("Begin失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:        "res-1",
			EventID:   "event-1",
			UserID:    "user-1",
			SeatIDs:   []string{"seat-1"},
			Status:    reservation.StatusPending,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(nil, errors.New("db error"))

		result, err := deps.service.ConfirmReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "トランザクション開始に失敗")
	})

	t.Run("ConfirmSeats失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:        "res-1",
			EventID:   "event-1",
			UserID:    "user-1",
			SeatIDs:   []string{"seat-1"},
			Status:    reservation.StatusPending,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
		deps.tx.On("Rollback").Return(nil)
		deps.seatRepo.On("ConfirmSeats", ctx, deps.tx, res.SeatIDs).Return(errors.New("seat confirm error"))

		result, err := deps.service.ConfirmReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Update失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:        "res-1",
			EventID:   "event-1",
			UserID:    "user-1",
			SeatIDs:   []string{"seat-1"},
			Status:    reservation.StatusPending,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
		deps.tx.On("Rollback").Return(nil)
		deps.seatRepo.On("ConfirmSeats", ctx, deps.tx, res.SeatIDs).Return(nil)
		deps.resRepo.On("Update", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(errors.New("update error"))

		result, err := deps.service.ConfirmReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Commit失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:        "res-1",
			EventID:   "event-1",
			UserID:    "user-1",
			SeatIDs:   []string{"seat-1"},
			Status:    reservation.StatusPending,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
		deps.tx.On("Rollback").Return(nil)
		deps.tx.On("Commit").Return(errors.New("commit error"))
		deps.seatRepo.On("ConfirmSeats", ctx, deps.tx, res.SeatIDs).Return(nil)
		deps.resRepo.On("Update", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)

		result, err := deps.service.ConfirmReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "コミットに失敗")
	})
}

func TestReservationService_CancelReservation_Errors(t *testing.T) {
	t.Run("予約が見つからない", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		deps.resRepo.On("GetByID", ctx, "nonexistent").Return(nil, reservation.ErrReservationNotFound)

		result, err := deps.service.CancelReservation(ctx, "nonexistent")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, reservation.ErrReservationNotFound))
	})

	t.Run("Begin失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:      "res-1",
			EventID: "event-1",
			UserID:  "user-1",
			SeatIDs: []string{"seat-1"},
			Status:  reservation.StatusPending,
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(nil, errors.New("db error"))

		result, err := deps.service.CancelReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "トランザクション開始に失敗")
	})

	t.Run("ReleaseSeats失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:      "res-1",
			EventID: "event-1",
			UserID:  "user-1",
			SeatIDs: []string{"seat-1"},
			Status:  reservation.StatusPending,
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
		deps.tx.On("Rollback").Return(nil)
		deps.seatRepo.On("ReleaseSeats", ctx, deps.tx, res.SeatIDs).Return(errors.New("release error"))

		result, err := deps.service.CancelReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Commit失敗", func(t *testing.T) {
		deps := newTestDeps()
		ctx := context.Background()

		res := &reservation.Reservation{
			ID:      "res-1",
			EventID: "event-1",
			UserID:  "user-1",
			SeatIDs: []string{"seat-1"},
			Status:  reservation.StatusPending,
		}
		deps.resRepo.On("GetByID", ctx, "res-1").Return(res, nil)
		deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
		deps.tx.On("Rollback").Return(nil)
		deps.tx.On("Commit").Return(errors.New("commit error"))
		deps.seatRepo.On("ReleaseSeats", ctx, deps.tx, res.SeatIDs).Return(nil)
		deps.resRepo.On("Update", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)
		deps.seatCache.On("Invalidate", ctx, "event-1").Return(nil)

		result, err := deps.service.CancelReservation(ctx, "res-1")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "コミットに失敗")
	})
}

func TestReservationService_CreateReservation_SeatReserveFailed(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)

	deps.resRepo.On("Create", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)
	deps.seatRepo.On("ReserveSeats", ctx, deps.tx, input.SeatIDs, mock.AnythingOfType("string")).Return(errors.New("seat reserve failed"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationService_CreateReservation_CommitFailed(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)
	deps.tx.On("Commit").Return(errors.New("commit failed"))

	deps.resRepo.On("Create", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(nil)
	deps.seatRepo.On("ReserveSeats", ctx, deps.tx, input.SeatIDs, mock.AnythingOfType("string")).Return(nil)

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "コミットに失敗")
}

func TestReservationService_CreateReservation_IdempotencyCheckDBError(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	// Return a DB error that is not ErrReservationNotFound
	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, errors.New("db connection error"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "冪等性チェックに失敗")
}

func TestReservationService_CreateReservation_EventGetError(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	deps.eventRepo.On("GetByID", ctx, "event-1").Return(nil, errors.New("event not found"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "イベント取得に失敗")
}

func TestReservationService_CreateReservation_SeatGetError(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(nil, errors.New("db error"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "座席取得に失敗")
}

func TestReservationService_CreateReservation_LockGenericError(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	// Return a generic error (not ErrLockNotAcquired)
	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(nil, errors.New("redis connection error"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ロック取得に失敗")
}

func TestReservationService_CreateReservation_ReservationCreateError(t *testing.T) {
	deps := newTestDeps()
	ctx := context.Background()

	input := CreateReservationInput{
		EventID:        "event-1",
		UserID:         "user-1",
		SeatIDs:        []string{"seat-1"},
		IdempotencyKey: "key-1",
	}

	deps.resRepo.On("GetByIdempotencyKey", ctx, input.IdempotencyKey).
		Return(nil, reservation.ErrReservationNotFound)

	deps.lockManager.On("AcquireLockWithRetry", ctx, mock.AnythingOfType("string"), 10*time.Second, 3, 100*time.Millisecond).
		Return(deps.lock, nil)
	deps.lock.On("Release", ctx).Return(nil)

	openEvent := &event.Event{
		ID:      "event-1",
		Name:    "Test Event",
		StartAt: time.Now().Add(1 * time.Hour),
		EndAt:   time.Now().Add(2 * time.Hour),
	}
	deps.eventRepo.On("GetByID", ctx, "event-1").Return(openEvent, nil)

	seats := []*seat.Seat{
		{ID: "seat-1", EventID: "event-1", Status: seat.StatusAvailable, Price: 1000},
	}
	deps.seatRepo.On("GetByEventID", ctx, "event-1").Return(seats, nil)

	deps.txManager.On("Begin", ctx).Return(deps.tx, nil)
	deps.tx.On("Rollback").Return(nil)

	deps.resRepo.On("Create", ctx, deps.tx, mock.AnythingOfType("*reservation.Reservation")).Return(errors.New("create error"))

	result, err := deps.service.CreateReservation(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
}
