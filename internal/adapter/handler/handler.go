package handler

import (
	"net/http"

	"github.com/Wyydra/ya/internal/core/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	RoomService *service.RoomService
}

func NewHandler(roomService *service.RoomService) *Handler {
	return &Handler{
		RoomService: roomService,
	}
}

func (h *Handler) NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/*", fs)

	r.Get("/ws", h.ServeWS)

	return r
}
