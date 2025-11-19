package game

import (
	"log"
	"sync"
	"time"

	"antimartingale/internal/config"
	"antimartingale/internal/model"

	"github.com/gorilla/websocket"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetBalance(userID string) (float64, error)
	UpdateBalance(userID string, newBalance float64) error
	GetNickname(userID string) (string, error)
	CreateUser(userID, nickname string, balance float64) error
	UserExists(userID string) (bool, error)
	GetOrCreateUser(userID, nickname string, defaultBalance float64) (balance float64, finalNickname string, err error)
}

// Game manages the game state and players
type Game struct {
	mutex            sync.Mutex
	phase            config.Phase
	multiplier       float64
	players          map[string]*model.Player
	connections      map[*websocket.Conn]string
	phaseTimer       *time.Timer
	confiscateTimer  *time.Timer
	phaseEndTime     time.Time
	statistics       model.Stats
	onlinePlayerList map[string]*model.PlayerInGameInfo
	userRepo         UserRepository

	// Provably fair fields
	serverSeed       string
	clientSeed       string
	timestamp        int64
	provablyFairHash string
}

// New creates and initializes a new Game instance
func New(userRepo UserRepository) *Game {
	return &Game{
		players:          make(map[string]*model.Player),
		connections:      make(map[*websocket.Conn]string),
		onlinePlayerList: make(map[string]*model.PlayerInGameInfo),
		userRepo:         userRepo,
	}
}

// Run starts the game loop
func (g *Game) Run() {
	g.StartBettingPhase()
}

// getBalance returns the balance for a given user ID
func (g *Game) getBalance(userID string) float64 {
	balance, err := g.userRepo.GetBalance(userID)
	if err != nil {
		log.Printf("Error getting balance for user %s: %v", userID, err)
		return 0
	}
	return balance
}

// updateBalance updates the balance for a given user ID
func (g *Game) updateBalance(userID string, newBalance float64) error {
	return g.userRepo.UpdateBalance(userID, newBalance)
}

// GetStatistics returns the current game statistics
func (g *Game) GetStatistics() model.Stats {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.statistics
}
