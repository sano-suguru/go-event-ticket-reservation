package reservation

import "errors"

// Reservation ドメインのエラー定義
var (
	ErrReservationNotFound         = errors.New("予約が見つかりません")
	ErrReservationNotPending       = errors.New("予約は保留中ではありません")
	ErrReservationExpired          = errors.New("予約の有効期限が切れています")
	ErrReservationAlreadyCancelled = errors.New("予約は既にキャンセルされています")
	ErrReservationAlreadyConfirmed = errors.New("予約は既に確定されています")
	ErrEventIDRequired             = errors.New("イベントIDは必須です")
	ErrUserIDRequired              = errors.New("ユーザーIDは必須です")
	ErrSeatIDsRequired             = errors.New("座席IDは必須です")
	ErrIdempotencyKeyRequired      = errors.New("冪等性キーは必須です")
	ErrIdempotencyKeyAlreadyExists = errors.New("同じ冪等性キーの予約が既に存在します")
)
