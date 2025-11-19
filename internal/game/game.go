package game

import (
	"sync"
	"time"

	"antimartingale/internal/config"
	"antimartingale/internal/model"

	"github.com/gorilla/websocket"
)

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
	userBalances     sync.Map
	userNicknames    sync.Map
	onlinePlayerList map[string]*model.PlayerInGameInfo

	// Provably fair fields
	serverSeed       string
	clientSeed       string
	timestamp        int64
	provablyFairHash string
}

// New creates and initializes a new Game instance
func New() *Game {
	return &Game{
		players:          make(map[string]*model.Player),
		connections:      make(map[*websocket.Conn]string),
		onlinePlayerList: make(map[string]*model.PlayerInGameInfo),
	}
}

// Run starts the game loop
func (g *Game) Run() {
	g.StartBettingPhase()
}

// GetBalance returns the balance for a given user ID
func (g *Game) getBalance(userID string) float64 {
	balance, _ := g.userBalances.Load(userID)
	return balance.(float64)
}

// GetStatistics returns the current game statistics
func (g *Game) GetStatistics() model.Stats {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.statistics
}
