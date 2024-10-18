package room

import (
	"big2/player"
	"big2/shared"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type Room struct {
	id                  string
	mu                  sync.Mutex
	game                shared.Game
	players             map[string]shared.Player
	disconnectedPlayers map[string]shared.Player
	state               shared.RoomState
	lastActivity        time.Time
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) SetMessage(msgType string, msgContent interface{}) {
	m.Type = msgType
	m.Content = msgContent
}

func GenerateID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

// func (r *Room) StartGame() {
// 	playerHands, garbageCard := r.Game.Start()
// 	r.Game.GarbageCard = garbageCard
// 	i := 0
// 	for _, player := range r.Players {
// 		player.Hand = playerHands[i]
// 		i++
// 	}

// 	// 找到擁有方塊 3 的玩家開始
// 	startingPlayer := r.findStartingPlayer()
// 	r.CurrentTurn = startingPlayer

// 	r.Broadcast(Message{Type: "game_start", Content: "The game is starting!"})

// 	// Send each player their hand
// 	for _, player := range r.Players {
// 		player.Conn.WriteJSON(Message{Type: "player_hand", Content: player.Hand})
// 	}
// }

// func (r *Room) findStartingPlayer() *player.Player {
// 	for _, player := range r.Players {
// 		for _, card := range player.Hand {
// 			if card.Suit == big2_game.Block && card.Value == big2_game.Three {
// 				return player
// 			}
// 		}
// 	}
// 	return nil
// }

func NewRoom(roomID string, game shared.Game) *Room {
	return &Room{
		id:                  roomID,
		game:                game,
		players:             make(map[string]shared.Player, 4),
		disconnectedPlayers: make(map[string]shared.Player),
	}
}

func (r *Room) MuLock() {
	r.mu.Lock()
}
func (r *Room) MuUnlock() {
	r.mu.Unlock()
}

func (r *Room) AddPlayer(shared.Player) error {
	return nil
}

func (r *Room) RemovePlayer(shared.Player) error {
	return nil
}

func (r *Room) GetPlayer(string) (shared.Player, error) {
	return nil, nil
}
func (r *Room) GetPlayers() []shared.Player {
	return nil
}
func (r *Room) GetDisconnectedPlayers() []shared.Player {
	return nil
}

func (r *Room) GetState() shared.RoomState {
	return shared.RoomStateWaiting
}
func (r *Room) SetState(shared.RoomState) {

}

// 列出 room 內的 Player
func (r *Room) ListPlayers() []string {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	players := make([]string, 0, len(r.Players))
	for _, player := range r.Players {
		players = append(players, player.GetID())
	}
	return players
}

func (r *Room) ReconnectPlayer(player *player.Player) {
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
		"current_turn": r.CurrentTurn,
		"last_play":    r.Game.LastPlay,
		"hand":         player.Hand,
	}})

	// 通知其他玩家该玩家已重连
	r.Broadcast(Message{Type: "player_reconnected", Content: map[string]interface{}{
		"player_id": player.ID,
	}})
}

// 對 room 內的 player broadcast
func (r *Room) Broadcast(message Message) {
	if r != nil && len(r.players) > 0 {
		for _, player := range r.players {
			err := player.GetConn().WriteJSON(message)
			if err != nil {
				slog.Error("Error room broadcasting to player %s: %v", player.ID, err)
			}
		}
	}
}
