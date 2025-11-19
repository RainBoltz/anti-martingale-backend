package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"antimartingale/internal/model"
)

// GameStats interface defines what handlers need from the game for stats
type GameStats interface {
	GetStatistics() model.Stats
}

// StatsHandler handles HTTP requests for game statistics
type StatsHandler struct {
	game GameStats
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(game GameStats) *StatsHandler {
	return &StatsHandler{
		game: game,
	}
}

// HandleStats responds with current game statistics
func (h *StatsHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := h.game.GetStatistics()
	response := h.buildStatsResponse(stats)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// buildStatsResponse constructs the statistics response
func (h *StatsHandler) buildStatsResponse(stats model.Stats) map[string]string {
	if stats.Rounds == 0 {
		return map[string]string{
			"rounds":          "0",
			"mean_multiplier": "0.0",
			"max_multiplier":  "0.0",
			"sum_bets":        "0.0",
			"sum_payouts":     "0.0",
			"house_edge":      "0.0",
		}
	}

	var houseEdge float64
	if stats.BetAcc == 0 {
		houseEdge = 0.0
	} else {
		houseEdge = (stats.BetAcc - stats.PayoutAcc) / stats.BetAcc
	}

	return map[string]string{
		"rounds":          strconv.FormatInt(int64(stats.Rounds), 10),
		"mean_multiplier": strconv.FormatFloat(stats.MultiAcc/float64(stats.Rounds), 'f', 2, 64),
		"max_multiplier":  strconv.FormatFloat(stats.MaxMulti, 'f', 2, 64),
		"sum_bets":        strconv.FormatFloat(stats.BetAcc, 'f', 2, 64),
		"sum_payouts":     strconv.FormatFloat(stats.PayoutAcc, 'f', 2, 64),
		"house_edge":      strconv.FormatFloat(houseEdge, 'f', 4, 64),
	}
}
