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

	eventRepo := postgres.NewEventRepository(db)
	eventService := application.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)
	healthHandler := handler.NewHealthHandler()

	e := echo.New()
	middleware.SetupMiddleware(e)

	api := e.Group("/api/v1")
	api.GET("/health", healthHandler.Check)
	api.POST("/events", eventHandler.Create)
	api.GET("/events", eventHandler.List)
	api.GET("/events/:id", eventHandler.GetByID)
	api.PUT("/events/:id", eventHandler.Update)
	api.DELETE("/events/:id", eventHandler.Delete)

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
