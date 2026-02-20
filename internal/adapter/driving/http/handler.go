package http

import (
	"net/http"

	"github.com/Wyydra/ya/internal/adapter/driven/gateway/ws"
	"github.com/Wyydra/ya/internal/core/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	ChatService *service.ChatService
	CallService *service.CallService
	Hub         *ws.Hub
}

func NewHandler(chatService *service.ChatService, callService *service.CallService, hub *ws.Hub) *Handler {
	return &Handler{
		ChatService: chatService,
		CallService: callService,
		Hub:         hub,
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
