package config

import (
	"os"
	"strconv"
	"time"
)

// Phase represents the different game phases
type Phase int

const (
	// BettingPhase is when players can place bets
	BettingPhase Phase = iota
	// CashoutPhase is when the multiplier increases and players can cash out
	CashoutPhase
	// ConfiscatePhase is when all remaining bets are confiscated
	ConfiscatePhase
)

// Game configuration constants
const (
	// PhaseDuration is the duration of betting and confiscate phases
	PhaseDuration = 10 * time.Second

	// MultiplierUpdateInterval is how often the multiplier updates during cashout phase
	MultiplierUpdateInterval = 100 * time.Millisecond

	// MultiplierIncrement is how much the multiplier increases per update
	MultiplierIncrement = 0.01

	// InitialMultiplier is the starting multiplier in cashout phase
	InitialMultiplier = 1.0

	// MinimumBetAmount is the minimum amount a player can bet
	MinimumBetAmount = 1.0

	// DefaultBalance is the initial balance for new players
	DefaultBalance = 10000.0

	// ServerPort is the port the server listens on
	ServerPort = ":8080"
)

// Probability distribution constants for game duration
const (
	// ProbabilityImmediateEnd is the chance of immediate game end (15%)
	ProbabilityImmediateEnd = 0.15

	// ProbabilityNormalRange is the chance of normal duration (65%)
	ProbabilityNormalRange = 0.8

	// ProbabilityShortGame is the chance of short game within normal range (30%)
	ProbabilityShortGame = 0.3
)

// Environment variable helpers
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
} {
	return struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "antimartingale"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// GetServerPort returns the server port from environment or default
func GetServerPort() string {
	return getEnv("SERVER_PORT", ":8080")
}
