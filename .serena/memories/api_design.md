# API設計

## ベースパス
`/api/v1`

## 認証
ヘッダー: `X-User-ID`（デモ用、本番はJWT推奨）

## エンドポイント一覧

### ヘルスチェック
- `GET /health` - ルートレベル（Railway/K8s対応）
- `GET /api/v1/health` - APIレベル

### イベント
- `POST /api/v1/events` - イベント作成
- `GET /api/v1/events` - イベント一覧
- `GET /api/v1/events/:id` - イベント詳細
- `PUT /api/v1/events/:id` - イベント更新
- `DELETE /api/v1/events/:id` - イベント削除

### 座席
- `GET /api/v1/events/:event_id/seats` - 座席一覧
- `POST /api/v1/events/:event_id/seats` - 座席作成
- `POST /api/v1/events/:event_id/seats/bulk` - 座席一括作成
- `GET /api/v1/events/:event_id/seats/available/count` - 空席数
- `GET /api/v1/seats/:id` - 座席詳細

### 予約
- `POST /api/v1/reservations` - 予約作成（冪等性キー必須）
- `GET /api/v1/reservations` - ユーザー予約一覧
- `GET /api/v1/reservations/:id` - 予約詳細
- `POST /api/v1/reservations/:id/confirm` - 予約確定（15分以内）
- `POST /api/v1/reservations/:id/cancel` - 予約キャンセル

### 監視
- `GET /metrics` - Prometheusメトリクス（認証なし、意図的に公開）
- `GET /swagger/*` - Swagger UI

## Swaggerアノテーション例
```go
// @Summary イベント作成
// @Description 新しいイベントを作成します
// @Tags events
// @Accept json
// @Produce json
// @Param request body CreateEventRequest true "イベント作成リクエスト"
// @Success 201 {object} EventResponse
// @Failure 400 {object} ErrorResponse
// @Router /events [post]
func (h *EventHandler) Create(c echo.Context) error {
```

## メトリクス
- `http_requests_total` - リクエスト数（method, path, status_code）
- `http_request_duration_seconds` - レスポンス時間
- `reservations_total` - 予約数（status: success/conflict/lock_failed）
- `distributed_lock_duration_seconds` - ロック取得時間
- `active_reservations` - アクティブ予約数
