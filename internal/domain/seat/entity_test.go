package seat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSeat(t *testing.T) {
	eventID := "event-123"
	seatNumber := "A-1"
	price := 5000

	seat := NewSeat(eventID, seatNumber, price)

	assert.Equal(t, eventID, seat.EventID)
	assert.Equal(t, seatNumber, seat.SeatNumber)
	assert.Equal(t, price, seat.Price)
	assert.Equal(t, StatusAvailable, seat.Status)
	assert.Nil(t, seat.ReservedBy)
	assert.Nil(t, seat.ReservedAt)
	assert.Equal(t, 0, seat.Version)
}

func TestSeat_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"利用可能", StatusAvailable, true},
		{"予約済み", StatusReserved, false},
		{"確定済み", StatusConfirmed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seat := &Seat{Status: tt.status}
			assert.Equal(t, tt.expected, seat.IsAvailable())
		})
	}
}

func TestSeat_Reserve(t *testing.T) {
	t.Run("利用可能な座席を予約できる", func(t *testing.T) {
		seat := NewSeat("event-123", "A-1", 5000)
		reservationID := "reservation-456"

		err := seat.Reserve(reservationID)

		require.NoError(t, err)
		assert.Equal(t, StatusReserved, seat.Status)
		assert.NotNil(t, seat.ReservedBy)
		assert.Equal(t, reservationID, *seat.ReservedBy)
		assert.NotNil(t, seat.ReservedAt)
	})

	t.Run("予約済みの座席は予約できない", func(t *testing.T) {
		seat := NewSeat("event-123", "A-1", 5000)
		seat.Status = StatusReserved

		err := seat.Reserve("reservation-456")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSeatNotAvailable)
	})

	t.Run("確定済みの座席は予約できない", func(t *testing.T) {
		seat := NewSeat("event-123", "A-1", 5000)
		seat.Status = StatusConfirmed

		err := seat.Reserve("reservation-456")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSeatNotAvailable)
	})
}

func TestSeat_Confirm(t *testing.T) {
	t.Run("予約済みの座席を確定できる", func(t *testing.T) {
		seat := NewSeat("event-123", "A-1", 5000)
		seat.Reserve("reservation-456")

		err := seat.Confirm()

		require.NoError(t, err)
		assert.Equal(t, StatusConfirmed, seat.Status)
	})

	t.Run("利用可能な座席は確定できない", func(t *testing.T) {
		seat := NewSeat("event-123", "A-1", 5000)

		err := seat.Confirm()

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSeatNotReserved)
	})
}

func TestSeat_Release(t *testing.T) {
	seat := NewSeat("event-123", "A-1", 5000)
	seat.Reserve("reservation-456")

	seat.Release()

	assert.Equal(t, StatusAvailable, seat.Status)
	assert.Nil(t, seat.ReservedBy)
	assert.Nil(t, seat.ReservedAt)
}

func TestSeat_Validate(t *testing.T) {
	tests := []struct {
		name        string
		seat        *Seat
		expectedErr error
	}{
		{
			name:        "有効な座席",
			seat:        &Seat{EventID: "event-123", SeatNumber: "A-1", Price: 5000},
			expectedErr: nil,
		},
		{
			name:        "イベントIDが空",
			seat:        &Seat{EventID: "", SeatNumber: "A-1", Price: 5000},
			expectedErr: ErrEventIDRequired,
		},
		{
			name:        "座席番号が空",
			seat:        &Seat{EventID: "event-123", SeatNumber: "", Price: 5000},
			expectedErr: ErrSeatNumberRequired,
		},
		{
			name:        "価格が負",
			seat:        &Seat{EventID: "event-123", SeatNumber: "A-1", Price: -100},
			expectedErr: ErrInvalidPrice,
		},
		{
			name:        "価格が0は有効",
			seat:        &Seat{EventID: "event-123", SeatNumber: "A-1", Price: 0},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.seat.Validate()
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
