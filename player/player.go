package player

import (
	"big2/big2_card"
	"big2/room"
	"big2/shared"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	heartbeatInterval = 5 * time.Second
	heartbeatTimeout  = 300 * time.Second
	reconnectWindow   = 60 * time.Second
)

type Player struct {
	id             string
	mu             sync.Mutex
	conn           *websocket.Conn
	hands          []big2_card.Card
	lastHeartbeat  time.Time
	heartbeatTimer *time.Timer
	disconnectTime time.Time
	room           shared.Room
	state          shared.PlayerState
	gameSort       int
}

func NewEmptyPlayer() *Player {
	return &Player{}
}

func NewPlayer(id string, conn *websocket.Conn) *Player {
	return &Player{
		id:            id,
		conn:          conn,
		lastHeartbeat: time.Now(),
	}
}

func (p *Player) GetID() string {
	return p.id
}

func (p *Player) GetConn() *websocket.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.conn
}

func (p *Player) SetConn(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conn = conn
}
func (p *Player) GetHands() []big2_card.Card {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.hands
}
func (p *Player) SetHands(cards []big2_card.Card) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.hands = cards
}

func (p *Player) RemoveHands(index int) []big2_card.Card {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append(p.hands[:index], p.hands[index+1:]...)
}

func (p *Player) GetState() shared.PlayerState {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.state
}
func (p *Player) SetState(state shared.PlayerState) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = state
}

func (p *Player) GetRoom() shared.Room {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.room
}

func (p *Player) SetRoom(room shared.Room) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.room = room
}
func (p *Player) GetGameSort() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.gameSort
}
func (p *Player) SetGameSort(sort int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.gameSort = sort
}

func (p *Player) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()

	slog.Info("[Disconnect] ", "player", p.id)
	p.disconnectTime = time.Now()
	p.conn.Close()
	if p.room != nil {
		p.room.RemovePlayer(p.GetID())
		p.room.SetDisconnectedPlayer(p)
		// 通知其他玩家该玩家已断线
		p.room.Broadcast("player_disconnected", map[string]interface{}{
			"player_id": p.id,
		})
		slog.Info("[Disconnect] ", "DisconnectedPlayer", p.room.GetDisconnectedPlayers())
	}
}

// 設定心跳機制，確認是否斷線
func (p *Player) StartHeartbeat() {
	// 定時器
	p.heartbeatTimer = time.NewTimer(heartbeatInterval)
	go func() {
		for {
			<-p.heartbeatTimer.C
			if time.Since(p.lastHeartbeat) > heartbeatTimeout {
				p.Disconnect()
				return
			}

			err := p.conn.WriteJSON(room.Message{Type: "heartbeat"})
			if err != nil {
				slog.Error("[StartHeartbeat]", "player", p.id, "error", err.Error())
				p.Disconnect()
				return
			}
			p.heartbeatTimer.Reset(heartbeatInterval)
		}
	}()
}

// client heartbeat response
func (p *Player) HandleHeartbeatResponse() {
	p.mu.Lock()
	defer p.mu.Unlock()

	slog.Info("[HandleHeartbeatResponse]")
	p.lastHeartbeat = time.Now()
}
