# Anti-Martingale Backend - Refactoring Documentation

## Overview

This document describes the refactoring performed to improve the project structure, readability, and maintainability.

## Refactoring Goals

1. **Separation of Concerns**: Split monolithic code into logical packages
2. **Improved Maintainability**: Easier to understand, test, and modify
3. **Better Code Organization**: Clear package structure following Go best practices
4. **Enhanced Readability**: Well-documented code with clear responsibilities
5. **Error Handling**: Improved error handling throughout the codebase

## New Project Structure

```
anti-martingale-backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration constants and enums
│   ├── model/
│   │   └── models.go         # Data models and structs
│   ├── game/
│   │   ├── game.go           # Core game logic and state
│   │   ├── phases.go         # Phase management (betting, cashout, confiscate)
│   │   ├── player.go         # Player connection and message handling
│   │   └── messaging.go      # WebSocket messaging functions
│   ├── handler/
│   │   ├── websocket.go      # WebSocket HTTP handler
│   │   └── http.go           # Statistics HTTP handler
│   └── util/
│       ├── nickname.go       # Nickname generation utility
│       └── provably_fair.go  # Provably fair game utilities
├── bin/                       # Build output (gitignored)
├── go.mod
├── go.sum
├── Makefile                   # Build automation
└── README.md                  # Original README

```

## Package Descriptions

### cmd/server
Contains the application entry point. Initializes the game and sets up HTTP routing.

### internal/config
Defines all configuration constants, enums, and configuration-related code. Includes:
- Phase enums (BettingPhase, CashoutPhase, ConfiscatePhase)
- Game timing constants
- Server configuration

### internal/model
Contains all data structures used throughout the application:
- Player struct
- Stats struct
- PlayerInGameInfo struct
- Message types for WebSocket communication

### internal/game
Core game logic separated into multiple files:
- **game.go**: Game struct definition and initialization
- **phases.go**: Phase management logic for all three game phases
- **player.go**: Player connection, disconnection, and message handling
- **messaging.go**: WebSocket communication functions

### internal/handler
HTTP request handlers:
- **websocket.go**: WebSocket upgrade and connection handling
- **http.go**: Statistics endpoint handler with CORS support

### internal/util
Utility functions:
- **nickname.go**: Random Chinese nickname generation
- **provably_fair.go**: Provably fair game duration calculation and hashing

## Key Improvements

### 1. Separation of Concerns
- Configuration is isolated in `config` package
- Data models are separated from business logic
- Network handling is separated from game logic
- Utilities are in dedicated packages

### 2. Improved Testability
- Interfaces allow for easy mocking (GameConnection, GameStats)
- Smaller, focused functions are easier to test
- Dependencies are injected rather than hardcoded

### 3. Better Error Handling
- WebSocket write errors are now logged
- HTTP errors return proper status codes
- Connection errors are handled gracefully

### 4. Code Documentation
- All exported types and functions have documentation comments
- Package-level documentation describes purpose
- Complex logic includes inline comments

### 5. Maintainability
- Related code is grouped together
- Each file has a single, clear responsibility
- Constants are named and documented
- Magic numbers are eliminated

## Building and Running

### Using Makefile (Recommended)

```bash
# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean

# Format code
make fmt

# Run tests
make test
```

### Using Go directly

```bash
# Build
go build -o bin/server ./cmd/server

# Run
go run ./cmd/server
```

## Migration from Old Code

The original `main.go` has been refactored into the new structure. No functionality has been changed, only reorganized. The API endpoints remain the same:

- WebSocket: `ws://localhost:8080/game`
- Statistics: `http://localhost:8080/stats`

## Future Improvements

Potential areas for further enhancement:

1. **Testing**: Add unit tests and integration tests
2. **Configuration**: Move hardcoded values to configuration files
3. **Database**: Add persistent storage for user data
4. **Security**: Implement proper origin checking and authentication
5. **Monitoring**: Add metrics and health check endpoints
6. **Rate Limiting**: Prevent abuse of WebSocket connections
7. **Graceful Shutdown**: Handle server shutdown properly
8. **Context Usage**: Add context for cancellation and timeouts

## Backward Compatibility

The refactored code maintains full backward compatibility with the original implementation. All client code should continue to work without modification.
