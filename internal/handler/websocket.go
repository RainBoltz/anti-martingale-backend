package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// GameConnection interface defines what handlers need from the game
type GameConnection interface {
	HandleConnection(conn *websocket.Conn)
}

// WebSocketHandler handles WebSocket upgrade and connections
type WebSocketHandler struct {
	game     GameConnection
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(game GameConnection) *WebSocketHandler {
	return &WebSocketHandler{
		game: game,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Implement proper origin checking in production
			},
		},
	}
}

// HandleConnection upgrades HTTP connection to WebSocket and handles the game connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	h.game.HandleConnection(conn)
}
