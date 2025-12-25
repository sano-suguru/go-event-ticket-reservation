# コードスタイルとコンベンション

## 言語・コメント
- コメントは日本語
- エラーメッセージは日本語（例: `errors.New("座席が見つかりません")`）
- Swaggerアノテーションは日本語

## フォーマット
- `gofmt` と `goimports` を使用
- ローカルパッケージプレフィックス: `github.com/sanosuguru/go-event-ticket-reservation`

## 関数シグネチャ
- I/O関数の第一引数は必ず `context.Context`
- SQLプレースホルダ: `$1, $2` (PostgreSQL形式)
- 配列パラメータ: `pq.Array()`

## エラー定義
各ドメインパッケージに `errors.go` を配置:
```go
var (
    ErrSeatNotFound     = errors.New("座席が見つかりません")
    ErrSeatNotAvailable = errors.New("座席は予約できません")
)
```
比較には `errors.Is()` を使用

## エラーラップ
```go
fmt.Errorf("予約作成に失敗: %w", err)
```

## ロギング
```go
logger.Info("メッセージ", zap.String("key", value))
logger.Error("エラー", zap.Error(err))
```

## ドメインパッケージ構成
```
domain/seat/
  entity.go       # エンティティと状態定数
  repository.go   # リポジトリインターフェース
  errors.go       # ドメインエラー定義
  entity_test.go  # ユニットテスト
```

## トランザクション境界
アプリケーション層のみでトランザクション管理:
```go
tx, _ := s.db.BeginTxx(ctx, nil)
defer tx.Rollback()  // 必須
s.reservationRepo.Create(ctx, tx, res)
s.seatRepo.ReserveSeats(ctx, tx, seatIDs, res.ID)
tx.Commit()
```

## 命名規則
- エンティティ: パッケージ名と同じ（`seat.Seat`, `event.Event`）
- リポジトリ: `Repository` インターフェース（`seat.Repository`）
- サービス: `*Service`（`application.ReservationService`）
- ハンドラ: `*Handler`（`handler.EventHandler`）
