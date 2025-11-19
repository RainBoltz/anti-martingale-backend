package game

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// sendBalance sends the player's current balance
func (g *Game) sendBalance(conn *websocket.Conn, balance float64) {
	if err := conn.WriteJSON(map[string]interface{}{
		"event": "balance",
		"data":  map[string]float64{"value": balance},
	}); err != nil {
		log.Printf("Error sending balance: %v", err)
	}
}

// sendResult sends the game result to a player
func (g *Game) sendResult(conn *websocket.Conn, profit float64) {
	if err := conn.WriteJSON(map[string]interface{}{
		"event": "result",
		"data":  map[string]float64{"profit": profit},
	}); err != nil {
		log.Printf("Error sending result: %v", err)
	}
}

// sendLockMulti sends the locked multiplier to a player
func (g *Game) sendLockMulti(conn *websocket.Conn, multi float64) {
	if err := conn.WriteJSON(map[string]interface{}{
		"event": "lock_multi",
		"data":  map[string]float64{"multi": multi},
	}); err != nil {
		log.Printf("Error sending lock_multi: %v", err)
	}
}

// sendBetAmount sends the current bet amount to a player
func (g *Game) sendBetAmount(conn *websocket.Conn, betAmount float64) {
	if err := conn.WriteJSON(map[string]interface{}{
		"event": "bet_amount",
		"data":  map[string]float64{"value": betAmount},
	}); err != nil {
		log.Printf("Error sending bet_amount: %v", err)
	}
}

// sendLoginConfirmed sends login confirmation to a player
func (g *Game) sendLoginConfirmed(conn *websocket.Conn, userID, nickname string) {
	if err := conn.WriteJSON(map[string]interface{}{
		"event": "login_confirmed",
		"data": map[string]string{
			"id":   userID,
			"name": nickname,
		},
	}); err != nil {
		log.Printf("Error sending login_confirmed: %v", err)
	}
}

// broadcast sends a message to all connected players
func (g *Game) broadcast(event string, data map[string]interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"event": event,
		"data":  data,
	})
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	for _, player := range g.players {
		if player.Connection != nil {
			if err := player.Connection.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("Error broadcasting to player %s: %v", player.UserID, err)
			}
		}
	}
}
