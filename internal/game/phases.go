package game

import (
	"math"
	"time"

	"antimartingale/internal/config"
	"antimartingale/internal/util"

	"github.com/google/uuid"
)

// StartBettingPhase initializes and starts the betting phase
func (g *Game) StartBettingPhase() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.broadcast("online_player_list", map[string]interface{}{"list": g.onlinePlayerList})

	g.phase = config.BettingPhase
	g.multiplier = 0.0
	g.phaseEndTime = time.Now().Add(config.PhaseDuration)

	// Reset player status
	for _, player := range g.players {
		player.BetAmount = 0
		player.LockedMulti = 0
		player.IsActive = false
	}

	g.broadcast("phase", map[string]interface{}{
		"phase":      "betting",
		"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
		"multiplier": g.multiplier,
		"multi":      0.0,
	})

	// Start countdown ticker
	ticker := time.NewTicker(100 * time.Millisecond)
	go g.bettingPhaseTicker(ticker)

	g.statistics.Rounds++
	g.phaseTimer = time.AfterFunc(config.PhaseDuration, g.StartCashoutPhase)
}

// bettingPhaseTicker sends regular updates during the betting phase
func (g *Game) bettingPhaseTicker(ticker *time.Ticker) {
	for range ticker.C {
		g.mutex.Lock()
		if g.phase == config.BettingPhase && time.Now().Before(g.phaseEndTime) {
			g.broadcast("phase", map[string]interface{}{
				"phase":      "betting",
				"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
				"multiplier": g.multiplier,
				"multi":      0.0,
			})
			g.mutex.Unlock()
		} else {
			ticker.Stop()
			g.mutex.Unlock()
			return
		}
	}
}

// StartCashoutPhase initializes and starts the cashout phase
func (g *Game) StartCashoutPhase() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.phase = config.CashoutPhase
	g.multiplier = config.InitialMultiplier
	tagTime := time.Now()

	// Setup provably fair system
	g.clientSeed = uuid.New().String()
	g.timestamp = tagTime.Unix()
	g.provablyFairHash = util.GenerateProvablyFairHash(g.serverSeed, g.clientSeed, g.timestamp)
	g.broadcast("provably_fair", map[string]interface{}{
		"client_seed": g.clientSeed,
		"timestamp":   g.timestamp,
		"hash":        g.provablyFairHash,
	})

	// Calculate game duration
	gameDuration, serverSeed := util.CalculateGameDuration()
	g.serverSeed = serverSeed
	g.phaseEndTime = tagTime.Add(gameDuration)

	g.broadcast("phase", map[string]interface{}{
		"phase":      "cashout",
		"countdown":  time.Now().UnixMilli() - tagTime.UnixMilli(),
		"multiplier": g.multiplier,
	})

	// Start multiplier update ticker
	ticker := time.NewTicker(config.MultiplierUpdateInterval)
	go g.cashoutPhaseTicker(ticker, tagTime)

	g.confiscateTimer = time.AfterFunc(time.Until(g.phaseEndTime), g.StartConfiscatePhase)
}

// cashoutPhaseTicker updates the multiplier during the cashout phase
func (g *Game) cashoutPhaseTicker(ticker *time.Ticker, tagTime time.Time) {
	for range ticker.C {
		g.mutex.Lock()
		if g.phase == config.CashoutPhase && time.Now().Before(g.phaseEndTime) {
			g.multiplier += config.MultiplierIncrement
			g.broadcast("phase", map[string]interface{}{
				"phase":      "cashout",
				"countdown":  time.Now().UnixMilli() - tagTime.UnixMilli(),
				"multiplier": g.multiplier,
			})
			g.mutex.Unlock()
		} else {
			ticker.Stop()
			g.mutex.Unlock()
			return
		}
	}
}

// StartConfiscatePhase initializes and starts the confiscate phase
func (g *Game) StartConfiscatePhase() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.phase = config.ConfiscatePhase
	g.phaseEndTime = time.Now().Add(config.PhaseDuration)
	g.broadcast("phase", map[string]interface{}{
		"phase":      "confiscate",
		"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
		"multiplier": g.multiplier,
	})

	// Update statistics
	g.statistics.MultiAcc += g.multiplier
	g.statistics.MaxMulti = math.Max(g.statistics.MaxMulti, g.multiplier)

	// Settlement
	var winnerNickname string
	var winnerPayout float64

	for _, player := range g.players {
		if player.IsActive {
			// Calculate profit
			profit := player.BetAmount * player.LockedMulti
			balance := g.getBalance(player.UserID)
			g.userBalances.Store(player.UserID, balance+profit)

			// Send results to connected player
			playerID, playerIsConnected := g.connections[player.Connection]
			if playerIsConnected && player.UserID == playerID {
				g.sendResult(player.Connection, profit)
				g.sendBalance(player.Connection, balance+profit)
				g.sendBetAmount(player.Connection, 0)
			}

			// Update statistics
			g.statistics.BetAcc += player.BetAmount
			g.statistics.PayoutAcc += profit

			// Track winner
			if profit > winnerPayout {
				winnerNickname = player.Nickname
				winnerPayout = profit
			}

			// Reset player's online status
			if info, exists := g.onlinePlayerList[player.Nickname]; exists {
				info.LockedMulti = 0.0
				info.BetAmount = 0.0
			}
		}
		player.IsActive = false
	}

	g.broadcast("winner", map[string]interface{}{
		"nickname": winnerNickname,
		"payout":   winnerPayout,
	})

	// Start confiscate phase countdown ticker
	ticker := time.NewTicker(time.Second)
	go g.confiscatePhaseTicker(ticker)

	time.AfterFunc(config.PhaseDuration, g.StartBettingPhase)
}

// confiscatePhaseTicker sends regular updates during the confiscate phase
func (g *Game) confiscatePhaseTicker(ticker *time.Ticker) {
	for range ticker.C {
		g.mutex.Lock()
		if g.phase == config.ConfiscatePhase && time.Now().Before(g.phaseEndTime) {
			g.broadcast("phase", map[string]interface{}{
				"phase":      "confiscate",
				"countdown":  g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(),
				"multiplier": g.multiplier,
			})
			g.mutex.Unlock()
		} else {
			ticker.Stop()
			g.mutex.Unlock()
			return
		}
	}
}
