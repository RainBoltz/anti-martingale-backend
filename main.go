package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	BettingPhase = iota
	CashoutPhase
	ConfiscatePhase
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Game struct {
	phase           int
	multiplier      float64
	players         map[*websocket.Conn]*Player
	mutex           sync.Mutex
	phaseTimer      *time.Timer
	confiscateTimer *time.Timer
	phaseEndTime    time.Time
	statistics      Stats
}

type Player struct {
	UserID      string
	BetAmount   float64
	LockedMulti float64
	IsActive    bool
}

type Stats struct {
	rounds    int
	multiAcc  float64
	betAcc    float64
	payoutAcc float64
	maxMulti  float64
}

var (
	userBalances sync.Map
	gameInstance = NewGame()
)

func NewGame() *Game {
	return &Game{
		players: make(map[*websocket.Conn]*Player),
	}
}

func (g *Game) Run() {
	g.StartBettingPhase()
}

func (g *Game) StartBettingPhase() {
	const PHASE_DURATION_SEC = 10

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.phase = BettingPhase
	g.multiplier = 0.0
	g.phaseEndTime = time.Now().Add(PHASE_DURATION_SEC * time.Second)

	// reset player status
	for _, player := range g.players {
		player.BetAmount = 0
		player.LockedMulti = 0
		player.IsActive = false
	}

	g.Broadcast("phase", map[string]interface{}{
		"phase":      "betting",
		"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
		"multiplier": g.multiplier,
		"multi":      0.0,
	})

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			g.mutex.Lock()
			if g.phase == BettingPhase && time.Now().Before(g.phaseEndTime) {
				g.Broadcast("phase", map[string]interface{}{
					"phase":      "betting",
					"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
					"multiplier": g.multiplier,
					"multi":      0.0,
				})
			} else {
				ticker.Stop()
				g.mutex.Unlock()
				return
			}
			g.mutex.Unlock()
		}
	}()

	g.statistics.rounds += 1
	g.phaseTimer = time.AfterFunc(PHASE_DURATION_SEC*time.Second, g.StartCashoutPhase)
}

func (g *Game) StartCashoutPhase() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.phase = CashoutPhase
	g.multiplier = 1.0
	tagTime := time.Now()
	randomDuration := 0 + rand.Float64()*9 // 1x ~ 10x
	g.phaseEndTime = tagTime.Add(time.Duration(randomDuration) * time.Second)

	g.Broadcast("phase", map[string]interface{}{
		"phase":      "cashout",
		"countdown":  time.Now().UnixMilli() - tagTime.UnixMilli(),
		"multiplier": g.multiplier,
	})

	// multiplier update
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			g.mutex.Lock()
			if g.phase == CashoutPhase && time.Now().Before(g.phaseEndTime) {
				g.multiplier += 0.01
				g.Broadcast("phase", map[string]interface{}{
					"phase":      "cashout",
					"countdown":  time.Now().UnixMilli() - tagTime.UnixMilli(),
					"multiplier": g.multiplier,
				})
			} else {
				ticker.Stop()
				g.mutex.Unlock()
				return
			}
			g.mutex.Unlock()
		}
	}()

	g.confiscateTimer = time.AfterFunc(time.Until(g.phaseEndTime), g.StartConfiscatePhase)
}

func (g *Game) StartConfiscatePhase() {
	const PHASE_DURATION_SEC = 10

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.phase = ConfiscatePhase
	g.phaseEndTime = time.Now().Add(PHASE_DURATION_SEC * time.Second)
	g.Broadcast("phase", map[string]interface{}{
		"phase":      "confiscate",
		"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
		"multiplier": g.multiplier,
	})
	g.statistics.multiAcc += g.multiplier
	g.statistics.maxMulti = math.Max(g.statistics.maxMulti, g.multiplier)

	// settlement
	for conn, player := range g.players {
		if player.IsActive {
			profit := player.BetAmount * player.LockedMulti
			balance := getBalance(player.UserID)
			userBalances.Store(player.UserID, balance+profit)
			g.SendResult(conn, profit)
			g.SendBalance(conn, balance+profit)

			g.statistics.betAcc += player.BetAmount
			g.statistics.payoutAcc += profit
		}
		player.IsActive = false
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			g.mutex.Lock()
			if g.phase == ConfiscatePhase && time.Now().Before(g.phaseEndTime) {
				g.Broadcast("phase", map[string]interface{}{
					"phase":      "confiscate",
					"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
					"multiplier": g.multiplier,
				})
			} else {
				ticker.Stop()
				g.mutex.Unlock()
				return
			}
			g.mutex.Unlock()
		}
	}()

	time.AfterFunc(PHASE_DURATION_SEC*time.Second, g.StartBettingPhase)
}

func (g *Game) HandleConnection(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			g.mutex.Lock()
			delete(g.players, conn)
			g.mutex.Unlock()
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err == nil {
			g.HandleMessage(conn, data)
		}
	}
}

func (g *Game) HandleMessage(conn *websocket.Conn, data map[string]interface{}) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	player, playerExists := g.players[conn]

	switch data["action"] {
	case "login":
		clientUserID, ok := data["id"].(string)
		if !ok {
			return
		}

		userID := clientUserID
		if clientUserID == "" { // never login before
			userID = uuid.NewString()
		}

		// create player
		g.players[conn] = &Player{
			UserID:   userID,
			IsActive: false,
		}

		g.SendLoginConfirmed(conn, userID)

		balance, _ := userBalances.LoadOrStore(userID, 100.0)
		g.SendBalance(conn, balance.(float64))

	case "bet":
		if g.phase == BettingPhase && !player.IsActive && playerExists {
			amount, ok := data["amount"].(float64)
			if !ok {
				return
			}

			if balance := getBalance(player.UserID); balance >= amount {
				userBalances.Store(player.UserID, balance-amount)
				player.BetAmount = amount
				player.IsActive = true
				g.SendBalance(conn, balance-amount)
			}
		}

	case "cashout":
		if player.IsActive && g.phase == CashoutPhase && playerExists {
			if player.LockedMulti == 0 {
				player.LockedMulti = g.multiplier
				g.SendLockMulti(conn, player.LockedMulti)
			}
		}
	}
}

func (g *Game) HandleStatsRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	var output string
	if g.statistics.rounds == 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"rounds":          "0",
			"mean_multiplier": "0.0",
			"max_multiplier":  "0.0",
			"total_bets":      "0.0",
			"total_payouts":   "0.0",
			"house_edge":      "0.0",
		})
		output = string(msg)
	} else {
		var houseEdge float64
		if g.statistics.betAcc == 0 {
			houseEdge = 0.0
		} else {
			houseEdge = (g.statistics.betAcc - g.statistics.payoutAcc) / g.statistics.betAcc
		}
		msg, _ := json.Marshal(map[string]interface{}{
			"rounds":          strconv.FormatInt(int64(g.statistics.rounds), 10),
			"mean_multiplier": strconv.FormatFloat(g.statistics.multiAcc/float64(g.statistics.rounds), 'f', 2, 64),
			"max_multiplier":  strconv.FormatFloat(g.statistics.maxMulti, 'f', 2, 64),
			"sum_bets":        strconv.FormatFloat(g.statistics.betAcc, 'f', 2, 64),
			"sum_payouts":     strconv.FormatFloat(g.statistics.payoutAcc, 'f', 2, 64),
			"house_edge":      strconv.FormatFloat(houseEdge, 'f', 4, 64),
		})
		output = string(msg)
	}

	fmt.Fprint(w, string(output))
}

func getBalance(userID string) float64 {
	balance, _ := userBalances.Load(userID)
	return balance.(float64)
}

func (g *Game) SendBalance(conn *websocket.Conn, balance float64) {
	conn.WriteJSON(map[string]interface{}{
		"event": "balance",
		"data":  map[string]float64{"value": balance},
	})
}

func (g *Game) SendResult(conn *websocket.Conn, profit float64) {
	conn.WriteJSON(map[string]interface{}{
		"event": "result",
		"data":  map[string]float64{"profit": profit},
	})
}

func (g *Game) SendLockMulti(conn *websocket.Conn, multi float64) {
	conn.WriteJSON(map[string]interface{}{
		"event": "lock_multi",
		"data":  map[string]float64{"multi": multi},
	})
}

func (g *Game) SendLoginConfirmed(conn *websocket.Conn, userId string) {
	conn.WriteJSON(map[string]interface{}{
		"event": "login_confirmed",
		"data":  map[string]string{"id": userId},
	})
}

func (g *Game) Broadcast(event string, data map[string]interface{}) {
	msg, _ := json.Marshal(map[string]interface{}{
		"event": event,
		"data":  data,
	})

	for conn := range g.players {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	// Create and start game instance
	gameInstance = NewGame()
	go gameInstance.Run()

	// Configure CORS
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true // TODO: remove in production
	}

	// Create a new mux for routing
	mux := http.NewServeMux()
	mux.HandleFunc("/game", gameInstance.HandleConnection)
	mux.HandleFunc("/stats", gameInstance.HandleStatsRequest)

	// // Start the server
	fmt.Println("Server starting on :8080")
	fmt.Println("(see stats at http://localhost:8080/stats)")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
