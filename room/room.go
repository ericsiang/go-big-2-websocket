package room

import (
	"big2/big2_game"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	heartbeatInterval = 5 * time.Second
	heartbeatTimeout  = 10 * time.Second
	reconnectWindow   = 60 * time.Second
)

type RoomState int

const (
	RoomStateWaiting RoomState = iota
	RoomStateInGame
	RoomStateEnded
)

type Player struct {
	ID             string
	Mu             sync.Mutex
	Conn           *websocket.Conn
	Room           *Room
	Hand           []big2_game.Card
	LastHeartbeat  time.Time
	HeartbeatTimer *time.Timer
	DisconnectTime time.Time
	State          RoomState
	LastActivity        time.Time
}

// 設定心跳機制，確認是否斷線
func (p *Player) StartHeartbeat() {
	// 定時器
	p.HeartbeatTimer = time.NewTimer(heartbeatInterval)
	go func() {
		for {
			<-p.HeartbeatTimer.C
			p.Mu.Lock()
			if time.Since(p.LastHeartbeat) > heartbeatTimeout {
				p.Mu.Unlock()
				p.DisconnectPlayer()
				return
			}
			p.Mu.Unlock()

			err := p.Conn.WriteJSON(Message{Type: "heartbeat"})
			if err != nil {
				slog.Error("Error sending heartbeat to player - ", p.ID, err.Error())
				p.DisconnectPlayer()
				return
			}
			p.HeartbeatTimer.Reset(heartbeatInterval)
		}
	}()
}

func (p *Player) HandleHeartbeatResponse() {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	slog.Info("[HandleHeartbeatResponse]")
	p.LastHeartbeat = time.Now()
}

func (p *Player) DisconnectPlayer() {
	slog.Info("[DisconnectPlayer] ", "player", p.ID)
	if p.Room != nil {
		p.Room.Mu.Lock()
		delete(p.Room.Players, p.ID)
		p.Room.DisconnectedPlayers[p.ID] = p
		p.Room.Mu.Unlock()
	}

	p.DisconnectTime = time.Now()
	p.Conn.Close()
	// 通知其他玩家该玩家已断线
	p.Room.Broadcast(Message{
		Type: "player_disconnected",
		Content: map[string]interface{}{
			"player_id": p.ID,
		},
	})
}

type Big2Game struct {
	Deck        *big2_game.Deck
	CurrentTurn *Player
	LastPlay    []big2_game.Card
	LastPlayer  *Player
	GarbageCard *big2_game.GarbageCard
	Passes      int
}

type Room struct {
	ID                  string
	Players             map[string]*Player
	Mu                  sync.Mutex
	Game                *Big2Game
	DisconnectedPlayers map[string]*Player
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

func GenerateID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func (r *Room) StartGame() {
	big2 := big2_game.Big2Card{}

	// Deal cards
	playerHands, garbageCard := big2.NewDeck(52)
	r.Game.GarbageCard = garbageCard
	i := 0
	for _, player := range r.Players {
		player.Hand = playerHands[i]
		i++
	}

	// 找到擁有方塊 3 的玩家開始
	startingPlayer := r.findStartingPlayer()
	r.Game.CurrentTurn = startingPlayer

	r.Broadcast(Message{Type: "game_start", Content: "The game is starting!"})

	// Send each player their hand
	for _, player := range r.Players {
		player.Conn.WriteJSON(Message{Type: "player_hand", Content: player.Hand})
	}
}

func (r *Room) findStartingPlayer() *Player {
	for _, player := range r.Players {
		for _, card := range player.Hand {
			if card.Suit == big2_game.Block && card.Value == big2_game.Three {
				return player
			}
		}
	}
	return nil
}

func (r *Room) ListPlayers() []string {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	players := make([]string, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player.ID)
	}
	return players
}

func (r *Room) ReconnectPlayer(player *Player) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	for i, p := range r.DisconnectedPlayers {
		if p.ID == player.ID {
			r.Players[i].Conn = player.Conn
			break
		}
	}

	// 发送当前游戏状态给重连的玩家
	player.Conn.WriteJSON(Message{Type: "game_state", Content: map[string]interface{}{
		"current_turn": r.Game.CurrentTurn,
		"last_play":    r.Game.LastPlay,
		"hand":         player.Hand,
	}})

	// 通知其他玩家该玩家已重连
	r.Broadcast(Message{Type: "player_reconnected", Content: map[string]interface{}{
		"player_id": player.ID,
	}})
}

func (r *Room) Broadcast(message Message) {
	if r != nil && len(r.Players) > 0 {
		for _, player := range r.Players {
			err := player.Conn.WriteJSON(message)
			if err != nil {
				slog.Error("Error room broadcasting to player %s: %v", player.ID, err)
			}
		}
	}
}
