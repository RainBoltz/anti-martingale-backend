package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"antimartingale/internal/config"
	"antimartingale/internal/database"
	"antimartingale/internal/game"
	"antimartingale/internal/handler"
)

func main() {
	// Get database configuration
	dbConfig := config.GetDatabaseConfig()

	// Initialize database connection
	db, err := database.New(database.Config{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: dbConfig.Password,
		DBName:   dbConfig.DBName,
		SSLMode:  dbConfig.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create user repository
	userRepo := database.NewUserRepository(db)

	// Initialize and start game
	gameInstance := game.New(userRepo)
	go gameInstance.Run()

	// Create handlers
	wsHandler := handler.NewWebSocketHandler(gameInstance)
	statsHandler := handler.NewStatsHandler(gameInstance)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/game", wsHandler.HandleConnection)
	mux.HandleFunc("/stats", statsHandler.HandleStats)

	// Get server port from config
	serverPort := config.GetServerPort()

	// Setup graceful shutdown
	go func() {
		log.Printf("Server starting on %s", serverPort)
		log.Printf("WebSocket endpoint: ws://localhost%s/game", serverPort)
		log.Printf("Statistics endpoint: http://localhost%s/stats", serverPort)

		if err := http.ListenAndServe(serverPort, mux); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
