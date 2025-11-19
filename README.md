# anti-martingale-backend

Anti-Martingale Mini game backend service, developed in Golang

## 遊戲簡介

輕鬆小品，單人遊戲，一局不到 1 分鐘，遊戲規則簡單好懂，收益也是純純的乘法非常好算，跟朋友一起挑戰爆擊賺大錢，刺激有趣。

## Project Structure

This project follows Go best practices with a clean architecture:

```
anti-martingale-backend/
├── cmd/server/         # Application entry point
├── internal/           # Private application code
│   ├── config/        # Configuration and constants
│   ├── database/      # Database connection and repositories
│   ├── model/         # Data models
│   ├── game/          # Core game logic
│   ├── handler/       # HTTP/WebSocket handlers
│   └── util/          # Utility functions
├── migrations/        # Database migrations
├── bin/               # Build output
├── Dockerfile         # Docker configuration
└── docker-compose.yml # Docker Compose configuration
```

For detailed refactoring documentation, see [README_REFACTORING.md](README_REFACTORING.md).

## Requirements

- Go 1.23+ (for local development)
- PostgreSQL 16+ (for local development)
- Docker & Docker Compose (recommended)

## How To

### 使用 Docker Compose 啟動 (推薦)

這是最簡單的方式，會自動啟動 PostgreSQL 資料庫和應用程式：

```bash
# 啟動服務（首次會自動建置）
docker-compose up

# 背景執行
docker-compose up -d

# 查看日誌
docker-compose logs -f app

# 停止服務
docker-compose down

# 停止並刪除資料
docker-compose down -v
```

### 本地開發

**1. 啟動 PostgreSQL 資料庫**

使用 Docker：
```bash
docker run -d \
  --name antimartingale-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=antimartingale \
  -p 5432:5432 \
  postgres:16-alpine
```

或使用系統安裝的 PostgreSQL。

**2. 設定環境變數**

複製環境變數範例檔：
```bash
cp .env.example .env
```

根據需要修改 `.env` 檔案中的設定。

**3. 執行應用程式**

**使用 Makefile:**
```bash
# 建置
make build

# 執行
make run

# 清理
make clean
```

**使用 Go 指令:**
```bash
# 安裝依賴
go mod download

# 建置
go build -o bin/server ./cmd/server

# 執行
go run ./cmd/server

# 或執行已建置的檔案
./bin/server
```

## 環境變數

| 變數名稱 | 預設值 | 說明 |
|---------|--------|------|
| DB_HOST | localhost | 資料庫主機 |
| DB_PORT | 5432 | 資料庫埠號 |
| DB_USER | postgres | 資料庫使用者 |
| DB_PASSWORD | postgres | 資料庫密碼 |
| DB_NAME | antimartingale | 資料庫名稱 |
| DB_SSLMODE | disable | SSL 模式 |
| SERVER_PORT | :8080 | 伺服器埠號 |

### API 說明

- `/game`: websocket 接口，提供前端遊戲連線
- `/stats`: 取得遊戲統計資料

## 遊戲玩法

遊戲採回合制，每回合分成三個階段：「下注 → 出場 → 沒收」，不斷循環進行。

收益形式採常見的「籌碼 \* 倍數」，玩家下注的金額為「籌碼」，遊戲期間會有浮動的「倍數」。

### 下注階段

這是回合前的準備階段，玩家可以下注任意金額，也就是選擇這回合個人的籌碼大小。

此階段大約 10 秒鐘。

### 出場階段

進入出場階段時，停止玩家下注。

隨著時間推進，倍率會從 x1.0 開始慢慢往上升（比如每 0.1 秒上升 0.01，所以 10 秒後的倍率是 x2.0 ），沒有上限。

此階段結束前，玩家都可以選擇「出場」並將當前倍率鎖定為個人倍率。

故回合結束時預期收益為「下注籌碼 \* 個人倍率」。

未下注與已出場的玩家只能等待下一個回合的下注階段。

### 沒收階段

這個階段一定會發生，但發生在出場階段的隨機時間點（或者被我們操弄的時間點）。

進入沒收階段時倍率會歸零，並強制玩家鎖定倍率出場，獲得「下注籌碼 \* 當前倍率 = 0」的收益（換句話說，就是莊家沒收所有未出場玩家的籌碼）。

### 情境舉例

#### 回合#1

- 下注階段：我首先下注 100 usdt
- 出場階段：倍率開始上升，達到倍率 x1.5 時我選擇出場，個人倍率為 x1.5
- 沒收階段：沒想到回合倍率最終達到 x3.0，但因為我 x1.5 就選擇出場了所以只拿到 100 \* 1.5 = 150 usdt

#### 回合#2

- 下注階段：我再下注 150 usdt
- 出場階段：倍率開始上升，由於上輪 x3.0 我沒吃到，所以我想等到至少 x 2.0 再出場
- 沒收階段：沒想到回合倍率最終僅至 x1.1，而我還沒出場，所以拿回 100 \* 0 = 0 usdt，剛剛賺的都沒了啦靠北，想申請未成年人退款了嗚嗚還我錢
