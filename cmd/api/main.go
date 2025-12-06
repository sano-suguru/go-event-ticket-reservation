package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	_ "github.com/sanosuguru/go-event-ticket-reservation/docs"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/handler"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/middleware"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/metrics"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/worker"
)

// @title イベントチケット予約システム API
// @version 1.0
// @description 高並行性イベントチケット予約システムのREST API
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-User-ID

func main() {
	cfg := config.Load()

	// ロガー初期化
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	log := logger.NewLogger(env)
	logger.Set(log)
	defer logger.Sync()

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("DB接続エラー", zap.Error(err))
	}
	defer db.Close()
	logger.Info("データベース接続成功")

	// マイグレーション実行（起動時に自動実行）
	if migrationErr := postgres.RunMigrations(db.DB, "db/migrations"); migrationErr != nil {
		logger.Fatal("マイグレーションエラー", zap.Error(migrationErr))
	}
	logger.Info("マイグレーション完了")

	// Redis接続
	redisClient, err := redisinfra.NewClient(&redisinfra.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Warn("Redis接続エラー（分散ロック無効）", zap.Error(err))
	}
	var lockManager *redisinfra.LockManager
	var seatCache *redisinfra.SeatCache
	if redisClient != nil {
		lockManager = redisinfra.NewLockManager(redisClient)
		seatCache = redisinfra.NewSeatCache(redisClient)
		defer redisClient.Close()
		logger.Info("Redis接続成功")
	}

	// Repositories
	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	// Services
	eventService := application.NewEventService(eventRepo)
	seatService := application.NewSeatService(db, seatRepo, eventRepo, seatCache)
	reservationService := application.NewReservationService(db, reservationRepo, seatRepo, eventRepo, lockManager, seatCache)

	// Handlers
	eventHandler := handler.NewEventHandler(eventService)
	seatHandler := handler.NewSeatHandler(seatService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	healthHandler := handler.NewHealthHandler()

	// Prometheusメトリクス初期化
	appMetrics := metrics.Init()

	e := echo.New()
	middleware.SetupMiddleware(e)

	// Prometheusミドルウェア追加
	e.Use(middleware.PrometheusMiddleware(appMetrics))

	// メトリクスエンドポイント
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// ヘルスチェック（ルートレベル - Railway/K8s対応）
	e.GET("/health", healthHandler.Check)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")
	api.GET("/health", healthHandler.Check)

	// Events
	api.POST("/events", eventHandler.Create)
	api.GET("/events", eventHandler.List)
	api.GET("/events/:id", eventHandler.GetByID)
	api.PUT("/events/:id", eventHandler.Update)
	api.DELETE("/events/:id", eventHandler.Delete)

	// Seats
	api.GET("/events/:event_id/seats", seatHandler.GetByEvent)
	api.POST("/events/:event_id/seats", seatHandler.Create)
	api.POST("/events/:event_id/seats/bulk", seatHandler.CreateBulk)
	api.GET("/events/:event_id/seats/available/count", seatHandler.CountAvailable)
	api.GET("/seats/:id", seatHandler.GetByID)

	// Reservations
	api.POST("/reservations", reservationHandler.Create)
	api.GET("/reservations", reservationHandler.GetUserReservations)
	api.GET("/reservations/:id", reservationHandler.GetByID)
	api.POST("/reservations/:id/confirm", reservationHandler.Confirm)
	api.POST("/reservations/:id/cancel", reservationHandler.Cancel)

	// 期限切れ予約クリーナーを開始
	ctx, cancel := context.WithCancel(context.Background())
	cleaner := worker.NewExpiredReservationCleaner(
		reservationService,
		1*time.Minute,  // 1分ごとにチェック
		15*time.Minute, // 15分以上経過した予約をキャンセル
	)
	go cleaner.Start(ctx)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		logger.Info("サーバー起動", zap.String("addr", addr))
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Fatal("サーバー起動エラー", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("シャットダウン開始...")

	// ワーカーを停止
	cancel()
	cleaner.Stop()
	logger.Info("バックグラウンドワーカー停止完了")

	// サーバーをシャットダウン
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("サーバーシャットダウンエラー", zap.Error(err))
	}
	logger.Info("サーバーが正常にシャットダウンしました")
}
