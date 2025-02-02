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
	players         map[string]*Player
	connections     map[*websocket.Conn]string
	mutex           sync.Mutex
	phaseTimer      *time.Timer
	confiscateTimer *time.Timer
	phaseEndTime    time.Time
	statistics      Stats
}

type Player struct {
	UserID      string
	Nickname    string
	BetAmount   float64
	LockedMulti float64
	IsActive    bool
	Connection  *websocket.Conn
}

type Stats struct {
	rounds    int
	multiAcc  float64
	betAcc    float64
	payoutAcc float64
	maxMulti  float64
}

var (
	adjectives = [...]string{"毛茸茸的", "兇猛的", "危險的", "有毒的", "溫馴的", "敏捷的", "聰明的", "具有攻擊性的", "微小的", "家養的", "野生的", "草食性的", "肉食性的", "可愛的", "具有攻擊性的", "敏捷的", "美麗的", "專橫的", "坦率的", "肉食性的", "聰明的", "冷酷的", "冷血的", "色彩繽紛的", "令人想擁抱的", "好奇的", "可愛的", "危險的", "致命的", "家養的", "支配的", "精力充沛的", "快速的", "好鬥的", "兇猛的", "猛烈的", "蓬鬆的", "友善的", "毛茸茸的", "模糊的", "暴躁的", "多毛的", "沉重的", "草食性的", "嫉妒的", "巨大的", "懶惰的", "吵鬧的", "討人喜歡的", "有愛心的", "惡意的", "母性的", "刻薄的", "凌亂的", "夜行性的", "吵鬧的", "愛管閒事的", "挑剔的", "愛玩的", "有毒的", "迅速的", "粗糙的", "無禮的", "有鱗的", "矮小的", "害羞的", "黏滑的", "緩慢的", "小的", "聰明的", "有異味的", "柔軟的", "有刺的", "臭的", "強壯的", "固執的", "順從的", "高的", "溫馴的", "頑強的", "有領地意識的", "微小的", "惡毒的", "溫暖的", "野生的"}
	animals    = [...]string{"土豚", "信天翁", "短吻鱷", "羊駝", "螞蟻", "食蟻獸", "羚羊", "猿", "犰狳", "驢", "狒狒", "獾", "梭魚", "蝙蝠", "熊", "海狸", "蜜蜂", "野牛", "野豬", "水牛", "蝴蝶", "駱駝", "水豚", "馴鹿", "食火雞", "貓", "毛毛蟲", "牛", "羚羊", "獵豹", "雞", "黑猩猩", "龍貓", "紅嘴山鴉", "蛤蜊", "眼鏡蛇", "蟑螂", "鱈魚", "鸕鶿", "郊狼", "螃蟹", "鶴", "鱷魚", "烏鴉", "杓鷸", "鹿", "恐龍", "狗", "狗魚", "海豚", "三趾鴴", "鴿子", "蜻蜓", "鴨子", "儒艮", "黑腹濱鷸", "老鷹", "針鼴", "鰻魚", "大羚羊", "大象", "麋鹿", "鴯鶓", "隼", "雪貂", "雀鳥", "魚", "紅鶴", "蒼蠅", "狐狸", "青蛙", "印度野牛", "瞪羚", "沙鼠", "長頸鹿", "蚋", "角馬", "山羊", "金翅雀", "金魚", "鵝", "大猩猩", "蒼鷹", "蚱蜢", "松雞", "原駝", "海鷗", "倉鼠", "野兔", "鷹", "刺蝟", "蒼鷺", "鯡魚", "河馬", "大黃蜂", "馬", "人類", "蜂鳥", "鬣狗", "山羊", "朱鷺", "胡狼", "美洲豹", "松鴉", "水母", "袋鼠", "翠鳥", "無尾熊", "笑翠鳥", "高棉牛", "捻角羚", "鳳頭麥雞", "雲雀", "狐猴", "豹", "獅子", "駱馬", "龍蝦", "蝗蟲", "懶猴", "蝨子", "琴鳥", "喜鵲", "綠頭鴨", "海牛", "山魈", "螳螂", "貂", "狐獴", "水貂", "鼴鼠", "貓鼬", "猴子", "麋鹿", "蚊子", "老鼠", "騾", "獨角鯨", "蠑螈", "夜鶯", "章魚", "霍加皮", "負鼠", "劍羚", "鴕鳥", "水獺", "貓頭鷹", "牡蠣", "黑豹", "鸚鵡", "鷓鴣", "孔雀", "鵜鶘", "企鵝", "野雞", "豬", "鴿子", "小馬", "豪豬", "海豚", "鵪鶉", "奎利亞", " 魁札爾鳥", "兔子", "浣熊", "秧雞", "綿羊", "老鼠", "烏鴉", "紅鹿", "小熊貓", "馴鹿", "犀牛", "白嘴鴉", "蠑螈", "鮭魚", "沙元", "鷸", "沙丁魚", "蠍子", "海馬", "海豹", "鯊魚", "羊", "鼩鼱", "臭鼬", "蝸牛", "蛇", "麻雀", "蜘蛛", "琵鷺", "魷魚", "松鼠", "八哥", "魟魚", "臭蟲", "鸛", "燕子", "天鵝", "賁", "眼鏡猴", "白蟻", "老虎", "蟾蜍", "鱒魚", "火雞", "烏龜", "毒蛇", "禿鷹", "小袋鼠", "海象", "黃蜂", "黃鼠狼", "鯨魚", "野貓", "狼", "金鋼狼", "袋熊", "啄木鳥", "蠕蟲", "鷦鷯", "犛牛", "斑馬"}
)

func generateNickname() string {
	return adjectives[rand.Intn(len(adjectives))] + animals[rand.Intn(len(animals))]
}

var (
	userBalances  sync.Map
	userNicknames sync.Map
	gameInstance  = NewGame()
)

func NewGame() *Game {
	return &Game{
		players:     make(map[string]*Player),
		connections: make(map[*websocket.Conn]string),
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

	g.Broadcast("phase", map[string]interface{}{"phase": "betting", "countdown": g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(), "multiplier": g.multiplier, "multi": 0.0})

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

	gameDuration := calculateGameEndTime()
	g.phaseEndTime = tagTime.Add(gameDuration)

	g.Broadcast("phase", map[string]interface{}{"phase": "cashout", "countdown": time.Now().UnixMilli() - tagTime.UnixMilli(), "multiplier": g.multiplier})

	// multiplier update
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			g.mutex.Lock()
			if g.phase == CashoutPhase && time.Now().Before(g.phaseEndTime) {
				g.multiplier += 0.01 //calculateMultiplierGrowth(g.multiplier)
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
	g.Broadcast("phase", map[string]interface{}{"phase": "confiscate", "countdown": g.phaseEndTime.UnixMilli() - time.Now().UnixMilli(), "multiplier": g.multiplier})
	g.statistics.multiAcc += g.multiplier
	g.statistics.maxMulti = math.Max(g.statistics.maxMulti, g.multiplier)

	// settlement
	for _, player := range g.players {
		if player.IsActive {
			profit := player.BetAmount * player.LockedMulti
			balance := getBalance(player.UserID)
			userBalances.Store(player.UserID, balance+profit)

			playerID, playerIsConnected := g.connections[player.Connection]
			if playerIsConnected && player.UserID == playerID {
				g.SendResult(player.Connection, profit)
				g.SendBalance(player.Connection, balance+profit)
				g.SendBetAmount(player.Connection, 0)
			}

			g.statistics.betAcc += player.BetAmount
			g.statistics.payoutAcc += profit
		}
		player.IsActive = false
	}

	ticker := time.NewTicker(time.Second)
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
			player, exists := g.players[g.connections[conn]]
			if exists && !player.IsActive {
				delete(g.players, g.connections[conn])
				fmt.Println("player status removed")
			}
			delete(g.connections, conn)
			fmt.Println("player disconnected")
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

	serverUserID, playerIsOnline := g.connections[conn]
	player, playerExists := g.players[serverUserID]

	switch data["action"] {
	case "login":
		fmt.Println("login")
		if playerIsOnline && playerExists { // refresh connection
			fmt.Println("refresh connection")
			delete(g.connections, conn)
			g.connections[conn] = player.UserID
			g.players[player.UserID].Connection = conn

			g.SendLoginConfirmed(conn, player.UserID, player.Nickname)

			balance, _ := userBalances.LoadOrStore(player.UserID, 10000.0)
			g.SendBalance(conn, balance.(float64))
			return
		}

		clientUserID, ok := data["id"].(string)
		if !ok {
			return
		}

		userID := clientUserID
		if clientUserID == "" { // never login before
			userID = uuid.NewString()
		} else if _, exists := g.players[userID]; exists { // reconnect
			fmt.Println("reconnect")
			g.connections[conn] = userID // update connection
			g.players[userID].Connection = conn

			g.SendLoginConfirmed(conn, userID, g.players[userID].Nickname)

			balance, _ := userBalances.LoadOrStore(userID, 10000.0)
			g.SendBalance(conn, balance.(float64))

			if g.phase == BettingPhase && g.players[userID].IsActive { // handle already bet
				g.SendBetAmount(conn, g.players[userID].BetAmount)
			}

			if g.phase == CashoutPhase && g.players[userID].LockedMulti != 0 { // handle already locked
				g.SendLockMulti(conn, g.players[userID].LockedMulti)
			}

			return
		}

		fmt.Println("new login")

		nickname, _ := userNicknames.LoadOrStore(userID, generateNickname())

		// create player
		g.players[userID] = &Player{
			UserID:     userID,
			Nickname:   nickname.(string),
			IsActive:   false,
			Connection: conn,
		}
		g.connections[conn] = userID

		g.SendLoginConfirmed(conn, userID, nickname.(string))

		balance, _ := userBalances.LoadOrStore(userID, 10000.0)
		g.SendBalance(conn, balance.(float64))

	case "bet":
		if g.phase != BettingPhase || !playerExists {
			return
		}

		amount, ok := data["amount"].(float64)
		balance := getBalance(player.UserID)
		if !player.IsActive {
			// conditions to ignore initialize
			if !ok || balance < amount || amount < 1.0 {
				return
			}
			// initialize
			player.BetAmount = 0.0
			player.IsActive = true
		}

		userBalances.Store(player.UserID, balance-amount)
		player.BetAmount += amount
		g.SendBalance(conn, balance-amount)
		g.SendBetAmount(conn, player.BetAmount)

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	conn.WriteJSON(map[string]interface{}{"event": "balance", "data": map[string]float64{"value": balance}})
}

func (g *Game) SendResult(conn *websocket.Conn, profit float64) {
	conn.WriteJSON(map[string]interface{}{"event": "result", "data": map[string]float64{"profit": profit}})
}

func (g *Game) SendLockMulti(conn *websocket.Conn, multi float64) {
	conn.WriteJSON(map[string]interface{}{"event": "lock_multi", "data": map[string]float64{"multi": multi}})
}

func (g *Game) SendBetAmount(conn *websocket.Conn, betAmount float64) {
	conn.WriteJSON(map[string]interface{}{"event": "bet_amount", "data": map[string]float64{"value": betAmount}})
}

func (g *Game) SendLoginConfirmed(conn *websocket.Conn, userId string, nickname string) {
	conn.WriteJSON(map[string]interface{}{"event": "login_confirmed", "data": map[string]string{"id": userId, "name": nickname}})
}

func (g *Game) Broadcast(event string, data map[string]interface{}) {
	msg, _ := json.Marshal(map[string]interface{}{"event": event, "data": data})

	for _, player := range g.players {
		conn := player.Connection
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

// earning optimization
func calculateMultiplierGrowth(currentMulti float64) float64 {
	// 倍率越高，增長速度越快
	baseGrowth := 0.01
	accelerator := 1.0 + (currentMulti-1.0)*0.1
	return baseGrowth * accelerator
}
func calculateGameEndTimeCheat() time.Duration {
	// 基礎遊戲時間範圍
	const minDuration = 0.0          // 最短 0 秒
	const maxDuration = 15.0         // 最長 15 秒
	const targetHouseEdgeRate = 0.05 // 目標賠率 5%
	const fuzzyRate = 0.25           // 加入 25% 隨機變化

	// 根據歷史賠率動態調整
	houseEdge := gameInstance.statistics.betAcc - gameInstance.statistics.payoutAcc
	targetEdge := gameInstance.statistics.betAcc * targetHouseEdgeRate

	// 如果賠率太高，傾向於更快結束遊戲
	adjustment := math.Max(0, (targetEdge-houseEdge)/targetEdge)

	duration := maxDuration - (maxDuration-minDuration)*adjustment
	randomFactor := (1.0 - fuzzyRate) + rand.Float64()*2*fuzzyRate

	return time.Duration(duration * randomFactor * float64(time.Second))
}
func calculateGameEndTime() time.Duration {
	if rand.Float64() < 0.7 {
		// 70% 使用均勻分佈（0~10 秒）
		return time.Duration(10*rand.Float64()) * time.Second
	} else {
		// 30% 使用指數分佈（均值 5 秒，限制在 0~50 秒）
		exp := rand.ExpFloat64() * 5.0
		clamped := math.Max(0.0, math.Min(exp, 50.0))
		return time.Duration(clamped) * time.Second
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
