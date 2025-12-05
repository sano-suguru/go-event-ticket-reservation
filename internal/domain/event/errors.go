package event

import "errors"

// Event ドメインのエラー定義
var (
	ErrEventNotFound     = errors.New("イベントが見つかりません")
	ErrEventNameRequired = errors.New("イベント名は必須です")
	ErrInvalidTotalSeats = errors.New("座席数は1以上である必要があります")
	ErrInvalidEventTime  = errors.New("終了時刻は開始時刻より後である必要があります")
)
