package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReservation(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		userID         string
		idempotencyKey string
		seatIDs        []string
		totalAmount    int
		wantErr        bool
		errExpected    error
	}{
		{
			name: "正常な予約作成", eventID: "event-456", userID: "user-123",
			idempotencyKey: "idem-key-1", seatIDs: []string{"seat-1", "seat-2"}, totalAmount: 20000,
			wantErr: false,
		},
		{
			name: "イベントID未指定", eventID: "", userID: "user-123",
			idempotencyKey: "idem-key-1", seatIDs: []string{"seat-1"}, totalAmount: 10000,
			wantErr: true, errExpected: ErrEventIDRequired,
		},
		{
			name: "ユーザーID未指定", eventID: "event-456", userID: "",
			idempotencyKey: "idem-key-1", seatIDs: []string{"seat-1"}, totalAmount: 10000,
			wantErr: true, errExpected: ErrUserIDRequired,
		},
		{
			name: "座席未選択", eventID: "event-456", userID: "user-123",
			idempotencyKey: "idem-key-1", seatIDs: []string{}, totalAmount: 10000,
			wantErr: true, errExpected: ErrSeatIDsRequired,
		},
		{
			name: "冪等性キー未指定", eventID: "event-456", userID: "user-123",
			idempotencyKey: "", seatIDs: []string{"seat-1"}, totalAmount: 10000,
			wantErr: true, errExpected: ErrIdempotencyKeyRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReservation(tt.eventID, tt.userID, tt.idempotencyKey, tt.seatIDs, tt.totalAmount)
			err := r.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.eventID, r.EventID)
			assert.Equal(t, tt.userID, r.UserID)
			assert.Equal(t, StatusPending, r.Status)
			assert.Equal(t, tt.totalAmount, r.TotalAmount)
		})
	}
}

func TestReservation_Confirm(t *testing.T) {
	r := createTestReservation(t)
	err := r.Confirm()
	require.NoError(t, err)
	assert.Equal(t, StatusConfirmed, r.Status)
	assert.NotNil(t, r.ConfirmedAt)
}

func TestReservation_Confirm_NotPending(t *testing.T) {
	r := createTestReservation(t)
	r.Status = StatusCancelled
	err := r.Confirm()
	assert.ErrorIs(t, err, ErrReservationNotPending)
}

func TestReservation_Confirm_Expired(t *testing.T) {
	r := createTestReservation(t)
	r.ExpiresAt = time.Now().Add(-1 * time.Minute)
	err := r.Confirm()
	assert.ErrorIs(t, err, ErrReservationExpired)
}

func TestReservation_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		wantErr error
	}{
		{"Pending状態からキャンセル", StatusPending, nil},
		{"Cancelled状態からキャンセル", StatusCancelled, ErrReservationAlreadyCancelled},
		{"Confirmed状態からキャンセル", StatusConfirmed, ErrReservationAlreadyConfirmed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := createTestReservation(t)
			r.Status = tt.status
			err := r.Cancel()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, StatusCancelled, r.Status)
			}
		})
	}
}

func TestReservation_IsExpired(t *testing.T) {
	r := createTestReservation(t)
	r.ExpiresAt = time.Now().Add(-1 * time.Minute)
	assert.True(t, r.IsExpired())
	r.ExpiresAt = time.Now().Add(10 * time.Minute)
	assert.False(t, r.IsExpired())
}

func TestReservation_IsPending(t *testing.T) {
	r := createTestReservation(t)
	assert.True(t, r.IsPending())
	r.Status = StatusConfirmed
	assert.False(t, r.IsPending())
}

func createTestReservation(t *testing.T) *Reservation {
	r := NewReservation("event-456", "user-123", "idem-key-1", []string{"seat-1"}, 10000)
	require.NoError(t, r.Validate())
	return r
}
