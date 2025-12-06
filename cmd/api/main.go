package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/handler"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/api/middleware"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
)

func main() {
	cfg := config.Load()

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}
	defer db.Close()

	// Repositories
	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	// Services
	eventService := application.NewEventService(eventRepo)
	seatService := application.NewSeatService(db, seatRepo, eventRepo)
	reservationService := application.NewReservationService(db, reservationRepo, seatRepo, eventRepo)

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

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバー起動エラー: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウンしています...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("シャットダウンエラー: %v", err)
	}
	log.Println("サーバーが正常にシャットダウンしました")
}
