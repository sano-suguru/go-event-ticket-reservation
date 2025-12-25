package e2e

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/api"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/handler"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/middleware"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

var (
	testServer  *TestServer
	testDB      *sqlx.DB
	redisClient *redis.Client
)

// TestMain はE2Eテストのエントリポイント
// パッケージ全体で1回だけサーバーを起動することで高速化
func TestMain(m *testing.M) {
	cfg := config.Load()

	// DB接続
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		os.Exit(0) // DB未起動時はスキップ
	}
	testDB = db

	// Redis接続
	rc, err := redisinfra.NewClient(&redisinfra.Config{
		Host: cfg.Redis.Host, Port: cfg.Redis.Port,
	})
	if err != nil {
		db.Close()
		os.Exit(0) // Redis未起動時はスキップ
	}
	redisClient = rc

	// サービス初期化
	lockManager := redisinfra.NewLockManager(redisClient)
	seatCache := redisinfra.NewSeatCache(redisClient)

	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)
	txManager := postgres.NewTxManager(db)

	eventService := application.NewEventService(eventRepo)
	seatService := application.NewSeatService(seatRepo, eventRepo, seatCache)
	reservationService := application.NewReservationService(txManager, reservationRepo, seatRepo, eventRepo, lockManager, seatCache)

	eventHandler := handler.NewEventHandler(eventService)
	seatHandler := handler.NewSeatHandler(seatService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	healthHandler := handler.NewHealthHandler()

	// Echo セットアップ
	e := echo.New()
	e.Validator = api.NewValidator()
	middleware.SetupMiddleware(e)

	e.GET("/health", healthHandler.Check)

	v1 := e.Group("/api/v1")
	v1.POST("/events", eventHandler.Create)
	v1.GET("/events", eventHandler.List)
	v1.GET("/events/:id", eventHandler.GetByID)
	v1.PUT("/events/:id", eventHandler.Update)
	v1.DELETE("/events/:id", eventHandler.Delete)

	v1.GET("/events/:event_id/seats", seatHandler.GetByEvent)
	v1.POST("/events/:event_id/seats", seatHandler.Create)
	v1.POST("/events/:event_id/seats/bulk", seatHandler.CreateBulk)
	v1.GET("/events/:event_id/seats/available/count", seatHandler.CountAvailable)

	v1.POST("/reservations", reservationHandler.Create)
	v1.GET("/reservations", reservationHandler.GetUserReservations)
	v1.GET("/reservations/:id", reservationHandler.GetByID)
	v1.POST("/reservations/:id/confirm", reservationHandler.Confirm)
	v1.POST("/reservations/:id/cancel", reservationHandler.Cancel)

	testServer = &TestServer{
		Echo:    e,
		Cleanup: func() {}, // 個別テストでは何もしない
	}

	// テスト実行
	code := m.Run()

	// 最終クリーンアップ
	cleanupTables()
	redisClient.Close()
	db.Close()

	os.Exit(code)
}

// cleanupTables はテーブルをクリーンアップ
func cleanupTables() {
	testDB.Exec("TRUNCATE TABLE reservation_seats, reservations, seats, events RESTART IDENTITY CASCADE")
}

// getTestServer は共有サーバーを取得（テスト前にテーブルをクリーンアップ）
func getTestServer(t *testing.T) *TestServer {
	t.Helper()
	if testServer == nil {
		t.Skip("テスト環境が利用できません")
	}
	cleanupTables()
	return testServer
}
