package model

import (
	"github.com/gorilla/websocket"
)

// Player represents a player in the game
type Player struct {
	UserID      string
	Nickname    string
	BetAmount   float64
	LockedMulti float64
	IsActive    bool
	Connection  *websocket.Conn
}

// Stats tracks game statistics
type Stats struct {
	Rounds    int
	MultiAcc  float64
	BetAcc    float64
	PayoutAcc float64
	MaxMulti  float64
}

// PlayerInGameInfo represents public player information displayed in the game
type PlayerInGameInfo struct {
	Nickname    string  `json:"nickname"`
	BetAmount   float64 `json:"bet_amount"`
	LockedMulti float64 `json:"locked_multi"`
}

// Message types for WebSocket communication

// LoginMessage represents a login request from a client
type LoginMessage struct {
	Action string `json:"action"`
	ID     string `json:"id"`
}

// BetMessage represents a bet placement from a client
type BetMessage struct {
	Action string  `json:"action"`
	Amount float64 `json:"amount"`
}

// CashoutMessage represents a cashout request from a client
type CashoutMessage struct {
	Action string `json:"action"`
}

// StatsResponse represents the statistics response
type StatsResponse struct {
	Rounds         string `json:"rounds"`
	MeanMultiplier string `json:"mean_multiplier"`
	MaxMultiplier  string `json:"max_multiplier"`
	SumBets        string `json:"sum_bets"`
	SumPayouts     string `json:"sum_payouts"`
	HouseEdge      string `json:"house_edge"`
}
