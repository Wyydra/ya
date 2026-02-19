package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wyydra/ya/internal/adapter/handler"
	"github.com/Wyydra/ya/internal/core/service"
	"github.com/rs/zerolog"
)

const PORT = ":8080"

func main() {
	w := zerolog.ConsoleWriter{Out: os.Stdout}
	l := zerolog.New(w).With().Timestamp().Caller().Logger()

	roomService := service.NewRoomService()
	go roomService.Run()

	h := handler.NewHandler(roomService)

	// Router is now encapsulated in the adapter layer
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

	roomService.Stop()
	l.Info().Msg("Server exited")
}
