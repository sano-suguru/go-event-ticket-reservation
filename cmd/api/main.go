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
	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/handler"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/middleware"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/worker"
)

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
	if redisClient != nil {
		lockManager = redisinfra.NewLockManager(redisClient)
		defer redisClient.Close()
		logger.Info("Redis接続成功")
	}

	// Repositories
	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	// Services
	eventService := application.NewEventService(eventRepo)
	seatService := application.NewSeatService(db, seatRepo, eventRepo)
	reservationService := application.NewReservationService(db, reservationRepo, seatRepo, eventRepo, lockManager)

	// Handlers
	eventHandler := handler.NewEventHandler(eventService)
	seatHandler := handler.NewSeatHandler(seatService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	healthHandler := handler.NewHealthHandler()

	e := echo.New()
	middleware.SetupMiddleware(e)

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
