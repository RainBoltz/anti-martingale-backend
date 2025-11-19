package game

import (
	"encoding/json"
	"log"

	"antimartingale/internal/config"
	"antimartingale/internal/model"
	"antimartingale/internal/util"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// HandleConnection manages WebSocket connections from clients
func (g *Game) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			g.handleDisconnection(conn)
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err == nil {
			g.HandleMessage(conn, data)
		}
	}
}

// handleDisconnection cleans up when a player disconnects
func (g *Game) handleDisconnection(conn *websocket.Conn) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	userID, hasUserID := g.connections[conn]
	if !hasUserID {
		log.Println("Connection closed: player not logged in")
		return
	}

	player, exists := g.players[userID]
	if exists {
		if _, isOnline := g.onlinePlayerList[player.Nickname]; isOnline {
			delete(g.onlinePlayerList, player.Nickname)
			log.Println("Player removed from online list:", player.Nickname)
		}

		// Only remove player if they're not active in a game
		if !player.IsActive {
			delete(g.players, userID)
			log.Println("Inactive player removed:", userID)
		}
	}

	delete(g.connections, conn)
	log.Println("Player disconnected")

	g.broadcast("online_player_list", map[string]interface{}{"list": g.onlinePlayerList})
}

// HandleMessage processes incoming messages from clients
func (g *Game) HandleMessage(conn *websocket.Conn, data map[string]interface{}) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	serverUserID, playerIsOnline := g.connections[conn]
	player, playerExists := g.players[serverUserID]

	action, ok := data["action"].(string)
	if !ok {
		return
	}

	switch action {
	case "login":
		g.handleLogin(conn, data, playerIsOnline, playerExists, player)
	case "bet":
		g.handleBet(conn, data, playerExists, player)
	case "cashout":
		g.handleCashout(conn, playerExists, player)
	}
}

// handleLogin processes player login requests
func (g *Game) handleLogin(conn *websocket.Conn, data map[string]interface{}, playerIsOnline, playerExists bool, player *model.Player) {
	log.Println("Processing login request")

	// Handle connection refresh
	if playerIsOnline && playerExists {
		log.Println("Refreshing existing connection")
		delete(g.connections, conn)
		g.connections[conn] = player.UserID
		g.players[player.UserID].Connection = conn

		g.sendLoginConfirmed(conn, player.UserID, player.Nickname)
		balance := g.getBalance(player.UserID)
		g.sendBalance(conn, balance)
		return
	}

	clientUserID, ok := data["id"].(string)
	if !ok {
		return
	}

	// Handle reconnection
	if clientUserID != "" {
		if existingPlayer, exists := g.players[clientUserID]; exists {
			log.Println("Player reconnecting:", clientUserID)
			g.connections[conn] = clientUserID
			existingPlayer.Connection = conn

			g.sendLoginConfirmed(conn, clientUserID, existingPlayer.Nickname)
			balance := g.getBalance(clientUserID)
			g.sendBalance(conn, balance)

			// Restore betting state if in betting phase
			if g.phase == config.BettingPhase && existingPlayer.IsActive {
				g.sendBetAmount(conn, existingPlayer.BetAmount)
			}

			// Restore locked multiplier if in cashout phase
			if g.phase == config.CashoutPhase && existingPlayer.LockedMulti != 0 {
				g.sendLockMulti(conn, existingPlayer.LockedMulti)
			}

			return
		}
	}

	// Handle new login
	log.Println("New player login")
	userID := clientUserID
	if clientUserID == "" {
		userID = uuid.NewString()
	}

	// Get or create user in database
	nickname := util.GenerateNickname()
	balance, finalNickname, err := g.userRepo.GetOrCreateUser(userID, nickname, config.DefaultBalance)
	if err != nil {
		log.Printf("Error getting or creating user: %v", err)
		return
	}

	// Create new player in memory
	g.players[userID] = &model.Player{
		UserID:     userID,
		Nickname:   finalNickname,
		IsActive:   false,
		Connection: conn,
	}
	g.connections[conn] = userID

	g.sendLoginConfirmed(conn, userID, finalNickname)
	g.sendBalance(conn, balance)

	g.onlinePlayerList[finalNickname] = &model.PlayerInGameInfo{
		Nickname:    finalNickname,
		BetAmount:   0.0,
		LockedMulti: 0.0,
	}
	g.broadcast("online_player_list", map[string]interface{}{"list": g.onlinePlayerList})
}

// handleBet processes player betting requests
func (g *Game) handleBet(conn *websocket.Conn, data map[string]interface{}, playerExists bool, player *model.Player) {
	if g.phase != config.BettingPhase || !playerExists {
		return
	}

	amount, ok := data["amount"].(float64)
	balance := g.getBalance(player.UserID)

	if !player.IsActive {
		// Validate bet initialization
		if !ok || balance < amount || amount < config.MinimumBetAmount {
			return
		}
		// Initialize player for this round
		player.BetAmount = 0.0
		player.IsActive = true
	}

	// Process bet
	newBalance := balance - amount
	if err := g.updateBalance(player.UserID, newBalance); err != nil {
		log.Printf("Error updating balance for bet: %v", err)
		return
	}

	player.BetAmount += amount
	g.sendBalance(conn, newBalance)
	g.sendBetAmount(conn, player.BetAmount)

	// Update online player list
	g.onlinePlayerList[player.Nickname].BetAmount += amount
	g.broadcast("online_player_list", map[string]interface{}{"list": g.onlinePlayerList})
}

// handleCashout processes player cashout requests
func (g *Game) handleCashout(conn *websocket.Conn, playerExists bool, player *model.Player) {
	if !playerExists || !player.IsActive || g.phase != config.CashoutPhase {
		return
	}

	if player.LockedMulti == 0 {
		player.LockedMulti = g.multiplier
		g.sendLockMulti(conn, player.LockedMulti)
	}

	// Update online player list
	g.onlinePlayerList[player.Nickname].LockedMulti = player.LockedMulti
	g.broadcast("online_player_list", map[string]interface{}{"list": g.onlinePlayerList})
}
