package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/handler"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/middleware"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

// TestServer はE2Eテスト用のサーバー
type TestServer struct {
	Echo    *echo.Echo
	Cleanup func()
}

// NewTestServer はテスト用サーバーを作成
func NewTestServer(t *testing.T) *TestServer {
	cfg := config.Load()

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		t.Skipf("DB接続エラー: %v", err)
	}

	redisClient, err := redisinfra.NewClient(&redisinfra.Config{
		Host: cfg.Redis.Host, Port: cfg.Redis.Port,
	})
	if err != nil {
		t.Skipf("Redis接続エラー: %v", err)
	}

	lockManager := redisinfra.NewLockManager(redisClient)
	seatCache := redisinfra.NewSeatCache(redisClient)

	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	eventService := application.NewEventService(eventRepo)
	seatService := application.NewSeatService(db, seatRepo, eventRepo, seatCache)
	reservationService := application.NewReservationService(db, reservationRepo, seatRepo, eventRepo, lockManager, seatCache)

	eventHandler := handler.NewEventHandler(eventService)
	seatHandler := handler.NewSeatHandler(seatService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	healthHandler := handler.NewHealthHandler()

	e := echo.New()
	middleware.SetupMiddleware(e)

	e.GET("/health", healthHandler.Check)

	api := e.Group("/api/v1")
	api.POST("/events", eventHandler.Create)
	api.GET("/events", eventHandler.List)
	api.GET("/events/:id", eventHandler.GetByID)
	api.PUT("/events/:id", eventHandler.Update)
	api.DELETE("/events/:id", eventHandler.Delete)

	api.GET("/events/:event_id/seats", seatHandler.GetByEvent)
	api.POST("/events/:event_id/seats", seatHandler.Create)
	api.POST("/events/:event_id/seats/bulk", seatHandler.CreateBulk)
	api.GET("/events/:event_id/seats/available/count", seatHandler.CountAvailable)

	api.POST("/reservations", reservationHandler.Create)
	api.GET("/reservations", reservationHandler.GetUserReservations)
	api.GET("/reservations/:id", reservationHandler.GetByID)
	api.POST("/reservations/:id/confirm", reservationHandler.Confirm)
	api.POST("/reservations/:id/cancel", reservationHandler.Cancel)

	cleanup := func() {
		db.Exec("DELETE FROM reservation_seats")
		db.Exec("DELETE FROM reservations")
		db.Exec("DELETE FROM seats")
		db.Exec("DELETE FROM events")
		redisClient.Close()
		db.Close()
	}

	return &TestServer{Echo: e, Cleanup: cleanup}
}

// Request はHTTPリクエストを実行
func (s *TestServer) Request(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)
	return rec
}

// TestE2E_HealthCheck はヘルスチェックをテスト
func TestE2E_HealthCheck(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	rec := server.Request("GET", "/health", nil, nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

// TestE2E_CompleteReservationJourney は完全な予約ジャーニーをテスト
func TestE2E_CompleteReservationJourney(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	userID := "e2e-user-yamada"
	var eventID, seatID, reservationID string

	// 1. イベント作成
	t.Run("イベント作成", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "武道館ライブ 2025",
			"venue":       "日本武道館",
			"start_at":    time.Now().Add(14 * 24 * time.Hour).Format(time.RFC3339),
			"end_at":      time.Now().Add(14*24*time.Hour + 3*time.Hour).Format(time.RFC3339),
			"total_seats": 10000,
		}

		rec := server.Request("POST", "/api/v1/events", body, nil)
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		eventID = resp["id"].(string)
		assert.NotEmpty(t, eventID)
	})

	// 2. 座席一括作成
	t.Run("座席一括作成", func(t *testing.T) {
		body := map[string]interface{}{
			"prefix": "A",
			"count":  5,
			"price":  15000,
		}

		path := fmt.Sprintf("/api/v1/events/%s/seats/bulk", eventID)
		rec := server.Request("POST", path, body, nil)
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp []map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Len(t, resp, 5)
		seatID = resp[0]["id"].(string)
	})

	// 3. 空席数確認
	t.Run("空席数確認", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/events/%s/seats/available/count", eventID)
		rec := server.Request("GET", path, nil, nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, float64(5), resp["count"])
	})

	// 4. 予約作成
	t.Run("予約作成", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": "e2e-order-001",
		}

		rec := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": userID,
		})
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		reservationID = resp["id"].(string)
		assert.Equal(t, "pending", resp["status"])
		assert.Equal(t, float64(15000), resp["total_amount"])
	})

	// 5. 予約確定
	t.Run("予約確定", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/reservations/%s/confirm", reservationID)
		rec := server.Request("POST", path, nil, map[string]string{
			"X-User-ID": userID,
		})
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "confirmed", resp["status"])
	})

	// 6. 空席数が減っていることを確認
	t.Run("空席数減少確認", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/events/%s/seats/available/count", eventID)
		rec := server.Request("GET", path, nil, nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, float64(4), resp["count"])
	})

	// 7. 予約詳細確認
	t.Run("予約詳細確認", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/reservations/%s", reservationID)
		rec := server.Request("GET", path, nil, map[string]string{
			"X-User-ID": userID,
		})
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, reservationID, resp["id"])
		assert.Equal(t, "confirmed", resp["status"])
	})
}

// TestE2E_ReservationConflict は予約競合をテスト
func TestE2E_ReservationConflict(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	var eventID, seatID string

	// セットアップ
	body := map[string]interface{}{
		"name":        "競合テストイベント",
		"venue":       "テスト会場",
		"start_at":    time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"end_at":      time.Now().Add(7*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
		"total_seats": 1,
	}
	rec := server.Request("POST", "/api/v1/events", body, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var eventResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &eventResp)
	eventID = eventResp["id"].(string)

	seatBody := map[string]interface{}{"prefix": "VIP", "count": 1, "price": 50000}
	rec = server.Request("POST", fmt.Sprintf("/api/v1/events/%s/seats/bulk", eventID), seatBody, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var seatsResp []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &seatsResp)
	seatID = seatsResp[0]["id"].(string)

	t.Run("ユーザーAが予約成功", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": "conflict-user-a",
		}
		rec := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": "user-A",
		})
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("ユーザーBが同じ座席を予約しようとして失敗", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": "conflict-user-b",
		}
		rec := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": "user-B",
		})
		// 競合エラー（400 または 409）
		assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusConflict,
			"期待: 400 or 409, 実際: %d", rec.Code)
	})
}

// TestE2E_CancelAndRebook はキャンセル後の再予約をテスト
func TestE2E_CancelAndRebook(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	var eventID, seatID, reservationID string

	// セットアップ
	eventBody := map[string]interface{}{
		"name":        "キャンセル再予約テスト",
		"venue":       "テスト会場",
		"start_at":    time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339),
		"end_at":      time.Now().Add(5*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
		"total_seats": 1,
	}
	rec := server.Request("POST", "/api/v1/events", eventBody, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var eventResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &eventResp)
	eventID = eventResp["id"].(string)

	seatBody := map[string]interface{}{"prefix": "S", "count": 1, "price": 10000}
	rec = server.Request("POST", fmt.Sprintf("/api/v1/events/%s/seats/bulk", eventID), seatBody, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var seatsResp []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &seatsResp)
	seatID = seatsResp[0]["id"].(string)

	t.Run("ユーザーAが予約", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": "cancel-rebook-a",
		}
		rec := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": "user-A",
		})
		require.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		reservationID = resp["id"].(string)
	})

	t.Run("ユーザーAがキャンセル", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/reservations/%s/cancel", reservationID)
		rec := server.Request("POST", path, nil, map[string]string{
			"X-User-ID": "user-A",
		})
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "cancelled", resp["status"])
	})

	t.Run("ユーザーBが再予約に成功", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": "cancel-rebook-b",
		}
		rec := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": "user-B",
		})
		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

// TestE2E_IdempotencyKey は冪等性キーをテスト
func TestE2E_IdempotencyKey(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	var eventID, seatID string

	// セットアップ
	eventBody := map[string]interface{}{
		"name":        "冪等性テスト",
		"venue":       "テスト会場",
		"start_at":    time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339),
		"end_at":      time.Now().Add(3*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
		"total_seats": 10,
	}
	rec := server.Request("POST", "/api/v1/events", eventBody, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var eventResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &eventResp)
	eventID = eventResp["id"].(string)

	seatBody := map[string]interface{}{"prefix": "I", "count": 2, "price": 8000}
	rec = server.Request("POST", fmt.Sprintf("/api/v1/events/%s/seats/bulk", eventID), seatBody, nil)
	require.Equal(t, http.StatusCreated, rec.Code)
	var seatsResp []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &seatsResp)
	seatID = seatsResp[0]["id"].(string)

	idempotencyKey := "same-key-12345"
	userID := "user-idem"

	t.Run("同じ冪等性キーで2回リクエスト", func(t *testing.T) {
		body := map[string]interface{}{
			"event_id":        eventID,
			"seat_ids":        []string{seatID},
			"idempotency_key": idempotencyKey,
		}

		// 1回目
		rec1 := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": userID,
		})
		require.Equal(t, http.StatusCreated, rec1.Code)
		var resp1 map[string]interface{}
		json.Unmarshal(rec1.Body.Bytes(), &resp1)
		reservationID1 := resp1["id"].(string)

		// 2回目（同じキー）
		rec2 := server.Request("POST", "/api/v1/reservations", body, map[string]string{
			"X-User-ID": userID,
		})
		require.Equal(t, http.StatusCreated, rec2.Code)
		var resp2 map[string]interface{}
		json.Unmarshal(rec2.Body.Bytes(), &resp2)
		reservationID2 := resp2["id"].(string)

		// 同じ予約IDが返る
		assert.Equal(t, reservationID1, reservationID2, "同じ冪等性キーなら同じ予約IDが返るべき")
	})
}

// TestE2E_EventCRUD はイベントのCRUD操作をテスト
func TestE2E_EventCRUD(t *testing.T) {
	server := NewTestServer(t)
	defer server.Cleanup()

	var eventID string

	t.Run("イベント作成", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "CRUDテストイベント",
			"venue":       "テスト会場",
			"start_at":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			"end_at":      time.Now().Add(10*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
			"total_seats": 50,
		}
		rec := server.Request("POST", "/api/v1/events", body, nil)
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		eventID = resp["id"].(string)
	})

	t.Run("イベント取得", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/events/%s", eventID)
		rec := server.Request("GET", path, nil, nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "CRUDテストイベント", resp["name"])
	})

	t.Run("イベント一覧取得", func(t *testing.T) {
		rec := server.Request("GET", "/api/v1/events", nil, nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp []map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.GreaterOrEqual(t, len(resp), 1)
	})

	t.Run("イベント更新", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "更新後のイベント名",
			"venue":       "新会場",
			"start_at":    time.Now().Add(11 * 24 * time.Hour).Format(time.RFC3339),
			"end_at":      time.Now().Add(11*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
			"total_seats": 60,
		}
		path := fmt.Sprintf("/api/v1/events/%s", eventID)
		rec := server.Request("PUT", path, body, nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "更新後のイベント名", resp["name"])
	})

	t.Run("イベント削除", func(t *testing.T) {
		path := fmt.Sprintf("/api/v1/events/%s", eventID)
		rec := server.Request("DELETE", path, nil, nil)
		require.Equal(t, http.StatusNoContent, rec.Code)

		// 削除後は取得できない
		rec = server.Request("GET", path, nil, nil)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
