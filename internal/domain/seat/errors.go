package seat

import "errors"

// Seat ドメインのエラー定義
var (
	ErrSeatNotFound           = errors.New("座席が見つかりません")
	ErrSeatNotAvailable       = errors.New("座席は予約できません")
	ErrSeatNotReserved        = errors.New("座席は予約されていません")
	ErrSeatAlreadyReserved    = errors.New("座席は既に予約されています")
	ErrEventIDRequired        = errors.New("イベントIDは必須です")
	ErrSeatNumberRequired     = errors.New("座席番号は必須です")
	ErrInvalidPrice           = errors.New("価格は0以上である必要があります")
	ErrOptimisticLockConflict = errors.New("楽観的ロックの競合が発生しました")
)
