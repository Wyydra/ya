package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wyydra/ya/internal/adapter/driven/gateway/ws"
	"github.com/Wyydra/ya/internal/adapter/driven/media/pion"
	repo "github.com/Wyydra/ya/internal/adapter/driven/persistence/memory"
	handler "github.com/Wyydra/ya/internal/adapter/driving/http"
	"github.com/Wyydra/ya/internal/core/service"
	"github.com/rs/zerolog"
)

const PORT = ":8080"

func main() {
	w := zerolog.ConsoleWriter{Out: os.Stdout}
	l := zerolog.New(w).With().Timestamp().Caller().Logger()

	repo := repo.NewMessageRepository()
	hub := ws.NewHub()

	mediaEngine := pion.NewPionAdapter()
	
	chatService := service.NewChatService(repo, hub)
	callService := service.NewCallService(mediaEngine, hub)
	h := handler.NewHandler(chatService, callService, hub)

	go hub.Run()

	r := h.NewRouter()

	srv := &http.Server{
		Addr:    PORT,
		Handler: r,
	}

	go func() {
		l.Info().Str("port", PORT).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	l.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Error().Err(err).Msg("Server forced to shutdown")
	}

	hub.Stop()
	l.Info().Msg("Server exited")
}
