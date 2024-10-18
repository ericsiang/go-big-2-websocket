package player

import (
	"big2/room"
	"big2/shared"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	heartbeatInterval = 5 * time.Second
	heartbeatTimeout  = 10 * time.Second
	reconnectWindow   = 60 * time.Second
)

type Player struct {
	id   string
	mu   sync.Mutex
	conn *websocket.Conn
	// Hand           []big2_game.Card
	lastHeartbeat  time.Time
	heartbeatTimer *time.Timer
	disconnectTime time.Time
	room           shared.Room
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

func (p *Player) GetHand() []shared.Card {
	return nil
}
func (p *Player) SetHand([]shared.Card) {

}
func (p *Player) GetState() shared.PlayerState {
	return shared.PlayerStateDisconnected
}
func (p *Player) SetState(shared.PlayerState) {

}
func (p *Player) Disconnect() {
	slog.Info("[DisconnectPlayer] ", "player", p.id)
	// if p.room. != nil {
	// 	p.room.Lock()
	// 	delete(p.room.GetPlayer(), p.id)
	// 	p.room.disconnectedPlayers[p.id] = p
	// 	p.room.Unlock()
	// }

	p.disconnectTime = time.Now()
	p.conn.Close()
	// 通知其他玩家该玩家已断线
	msg := room.NewMessage()
	msg.SetMessage("player_disconnected", map[string]interface{}{
		"player_id": p.id,
	})
	p.room.Broadcast(msg)
}
func (p *Player) Reconnect() error {
	return nil
}

// 設定心跳機制，確認是否斷線
func (p *Player) StartHeartbeat() {
	// 定時器
	p.heartbeatTimer = time.NewTimer(heartbeatInterval)
	go func() {
		for {
			<-p.heartbeatTimer.C
			p.mu.Lock()
			if time.Since(p.lastHeartbeat) > heartbeatTimeout {
				p.mu.Unlock()
				p.DisconnectPlayer()
				return
			}
			p.mu.Unlock()

			err := p.conn.WriteJSON(room.Message{Type: "heartbeat"})
			if err != nil {
				slog.Error("Error sending heartbeat to player - ", p.id, err.Error())
				p.DisconnectPlayer()
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

// 處理斷線 Player
func (p *Player) DisconnectPlayer() {
	slog.Info("[DisconnectPlayer] ", "player", p.id)
	if p.room != nil {
		p.room.Lock()
		delete(p.room.GetPlayer(), p.id)
		p.room.disconnectedPlayers[p.id] = p
		p.room.Unlock()
	}

	p.disconnectTime = time.Now()
	p.conn.Close()
	// 通知其他玩家该玩家已断线
	p.room.Broadcast(room.Message{
		Type: "player_disconnected",
		Content: map[string]interface{}{
			"player_id": p.id,
		},
	})
}
