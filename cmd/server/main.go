package main

import (
	"log"
	"net/http"

	"antimartingale/internal/config"
	"antimartingale/internal/game"
	"antimartingale/internal/handler"
)

func main() {
	// Initialize and start game
	gameInstance := game.New()
	go gameInstance.Run()

	// Create handlers
	wsHandler := handler.NewWebSocketHandler(gameInstance)
	statsHandler := handler.NewStatsHandler(gameInstance)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/game", wsHandler.HandleConnection)
	mux.HandleFunc("/stats", statsHandler.HandleStats)

	// Start server
	log.Printf("Server starting on %s", config.ServerPort)
	log.Printf("WebSocket endpoint: ws://localhost%s/game", config.ServerPort)
	log.Printf("Statistics endpoint: http://localhost%s/stats", config.ServerPort)

	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
