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

## Recent Updates

### Database Migration (PostgreSQL)

**Date**: 2025-11-21

Migrated from in-memory cache to PostgreSQL for persistent data storage:

**Changes:**
- Created `internal/database/` package with connection management and repositories
- Implemented `UserRepository` with methods for user data operations
- Added automatic database migrations on startup
- Updated game logic to use database instead of `sync.Map`
- Added environment variable configuration for database settings

**Benefits:**
- Data persists across server restarts
- Scalable user management
- Better data integrity
- Production-ready persistence layer

**Files Added:**
- `internal/database/db.go`: Database connection and migrations
- `internal/database/user_repository.go`: User data repository
- `.env.example`: Environment variable template

### CI/CD Integration

**Date**: 2025-11-21

Added comprehensive CI/CD support for automated testing and deployment:

**Makefile Enhancements:**
- `make ci`: Complete CI pipeline (lint + test + build)
- `make ci-test`: Tests with coverage reporting
- `make ci-lint`: Linting and format checks
- `make ci-build`: Optimized production build
- `make coverage`: HTML coverage report generation
- `make deploy-docker`: Docker Compose deployment
- `make deploy-production`: Production binary build
- `make help`: Show all available commands

**CI/CD Workflows:**
- `.github/workflows/ci.yml`: GitHub Actions pipeline
- `.gitlab-ci.yml`: GitLab CI configuration
- Support for Jenkins, CircleCI, Travis CI, AWS CodeBuild

**Documentation:**
- `DEPLOYMENT.md`: Comprehensive deployment guide
- Updated README.md with CI/CD instructions

### Docker Support

**Date**: 2025-11-21

Full Docker and Docker Compose support:

**Features:**
- Multi-stage Dockerfile for optimized builds
- Docker Compose with PostgreSQL service
- Health checks and dependency management
- Volume persistence for database
- Non-root container user for security

**Files:**
- `Dockerfile`: Multi-stage production build
- `docker-compose.yml`: Complete stack configuration
- `.dockerignore`: Optimized build context

## Future Improvements

Potential areas for further enhancement:

1. **Testing**: Add unit tests and integration tests (CI/CD infrastructure ready)
2. **Configuration**: ✅ Completed - Environment variables implemented
3. **Database**: ✅ Completed - PostgreSQL with repository pattern
4. **Security**: Implement proper origin checking and authentication
5. **Monitoring**: Add metrics and health check endpoints (basic /stats available)
6. **Rate Limiting**: Prevent abuse of WebSocket connections
7. **Graceful Shutdown**: ✅ Completed - Signal handling implemented
8. **Context Usage**: Add context for cancellation and timeouts
9. **Distributed Systems**: Redis for session management across multiple instances
10. **Observability**: Structured logging, distributed tracing, metrics export

## Backward Compatibility

The refactored code maintains full backward compatibility with the original implementation. All client code should continue to work without modification.

**Note**: After database migration, user data from the in-memory cache is not automatically migrated. Users will start with default balance (10000.0) on first login after upgrade.
