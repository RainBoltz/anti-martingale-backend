package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"antimartingale/internal/config"
)

// GenerateProvablyFairHash creates a hash from server seed, client seed, and timestamp
// This allows players to verify game fairness
func GenerateProvablyFairHash(serverSeed, clientSeed string, timestamp int64) string {
	// Concatenate all inputs
	input := serverSeed + clientSeed + fmt.Sprintf("%d", timestamp)

	// Generate a SHA256 hash
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)

	// Return the hex string of the hash
	return hex.EncodeToString(hashBytes)
}

// CalculateGameDuration determines how long the cashout phase should last
// Uses probability distributions to create varied game outcomes
func CalculateGameDuration() (time.Duration, string) {
	var randomResult float64
	selectProbability := rand.Float64()

	if selectProbability < config.ProbabilityImmediateEnd {
		// 15% chance: immediately end
		randomResult = 0.0 + generateAlmostZeroWithLongMantissa()
	} else if selectProbability < config.ProbabilityNormalRange {
		// 65% chance: normal range
		if rand.Float64() < config.ProbabilityShortGame {
			// 30% of normal range: (1 + [0.0, 1.5])x
			randomResult = 15*rand.Float64() + generateAlmostZeroWithLongMantissa()
		} else {
			// 70% of normal range: (1 + [0.0, 0.5])x
			randomResult = 5*rand.Float64() + generateAlmostZeroWithLongMantissa()
		}
	} else {
		// 20% chance: super long time
		randomResult = rand.ExpFloat64()*30 + generateAlmostZeroWithLongMantissa()
	}

	serverSeed := fmt.Sprintf("%.20f", randomResult)
	duration := time.Duration(randomResult) * time.Second

	return duration, serverSeed
}

// generateAlmostZeroWithLongMantissa creates a very small random number
// This adds precision to the game duration calculation
func generateAlmostZeroWithLongMantissa() float64 {
	base := 1e-9
	randomFraction := rand.Float64() + 0.1
	return base * randomFraction
}
