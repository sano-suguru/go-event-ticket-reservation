package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	// Arrange
	name := "テストコンサート"
	description := "素晴らしいコンサート"
	venue := "東京ドーム"
	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(3 * time.Hour)
	totalSeats := 100

	// Act
	event := NewEvent(name, description, venue, startAt, endAt, totalSeats)

	// Assert
	assert.Equal(t, name, event.Name)
	assert.Equal(t, description, event.Description)
	assert.Equal(t, venue, event.Venue)
	assert.Equal(t, startAt, event.StartAt)
	assert.Equal(t, endAt, event.EndAt)
	assert.Equal(t, totalSeats, event.TotalSeats)
	assert.Equal(t, 0, event.Version)
	assert.NotZero(t, event.CreatedAt)
	assert.NotZero(t, event.UpdatedAt)
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name        string
		event       *Event
		expectedErr error
	}{
		{
			name: "有効なイベント",
			event: &Event{
				Name:       "テストイベント",
				TotalSeats: 100,
				StartAt:    time.Now(),
				EndAt:      time.Now().Add(1 * time.Hour),
			},
			expectedErr: nil,
		},
		{
			name: "イベント名が空",
			event: &Event{
				Name:       "",
				TotalSeats: 100,
				StartAt:    time.Now(),
				EndAt:      time.Now().Add(1 * time.Hour),
			},
			expectedErr: ErrEventNameRequired,
		},
		{
			name: "座席数が0",
			event: &Event{
				Name:       "テストイベント",
				TotalSeats: 0,
				StartAt:    time.Now(),
				EndAt:      time.Now().Add(1 * time.Hour),
			},
			expectedErr: ErrInvalidTotalSeats,
		},
		{
			name: "座席数が負",
			event: &Event{
				Name:       "テストイベント",
				TotalSeats: -1,
				StartAt:    time.Now(),
				EndAt:      time.Now().Add(1 * time.Hour),
			},
			expectedErr: ErrInvalidTotalSeats,
		},
		{
			name: "終了時刻が開始時刻より前",
			event: &Event{
				Name:       "テストイベント",
				TotalSeats: 100,
				StartAt:    time.Now().Add(1 * time.Hour),
				EndAt:      time.Now(),
			},
			expectedErr: ErrInvalidEventTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
