package room

import (
	"big2/big2_card"
	"big2/handle_errors"
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

func (r *Room) StartGame() {
	err := r.game.Start()
	if err != nil {
		return
	}
	r.Broadcast("game_start", "The game is starting!")
	slog.Info("[StartGame]", "game_start", "The game is starting!")

	// 開始發牌
	newPlayerDeck, _ := r.game.GetHandCards()
	i := 0
	for _, player := range r.GetPlayers() {
		player.SetGameSort(i)
		player.SetHands(newPlayerDeck[i])
		r.game.SetGameSortPlayer(i, player)

		// Send each player their hand
		msg := NewMessage()
		msg.SetMessage("player_hand", player.GetHands())
		player.GetConn().WriteJSON(msg)
		i++
	}

	// 找到擁有方塊 3 的玩家開始
	startingPlayer := r.findStartingPlayer()
	r.game.SetCurrentTurn(startingPlayer.GetGameSort())
	r.Broadcast("first_player_broadcast", map[string]interface{}{
		"player_id": startingPlayer.GetID(),
	})
	msg := NewMessage()
	msg.SetMessage("first_player", "你是先手")
	startingPlayer.GetConn().WriteJSON(msg)
	return
}

func (r *Room) findStartingPlayer() shared.Player {
	for _, player := range r.GetPlayers() {
		for _, card := range player.GetHands() {
			if card.Suit == big2_card.Block && card.Value == big2_card.Three {
				return player
			}
		}
	}
	return nil
}

func NewRoom(roomID string, game shared.Game) *Room {
	return &Room{
		id:                  roomID,
		game:                game,
		players:             make(map[string]shared.Player, 4),
		disconnectedPlayers: make(map[string]shared.Player),
	}
}

func (r *Room) GetID() string {
	return r.id
}

func (r *Room) GetGame() shared.Game {
	return r.game
}

func (r *Room) AddPlayer(player shared.Player) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.players[player.GetID()] = player
	slog.Info("[AddPlayer]", "player.GetID()", player.GetID())
	slog.Info("[AddPlayer]", "player", player)
	slog.Info("[AddPlayer]", "r.players", r.players)

	player.SetRoom(r)
}

func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.players, playerID)
}

func (r *Room) GetPlayer(playerID string) (shared.Player, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, ok := r.players[playerID]
	return player, ok
}

func (r *Room) GetPlayers() map[string]shared.Player {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Info("[GetPlayers]", "r.players", r.players)
	return r.players
}
func (r *Room) GetDisconnectedPlayers() map[string]shared.Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.disconnectedPlayers
}

func (r *Room) SetDisconnectedPlayer(player shared.Player) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.disconnectedPlayers[player.GetID()] = player
}

func (r *Room) RemoveDisconnectedPlayer(player shared.Player) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.disconnectedPlayers, player.GetID())
}

func (r *Room) GetState() shared.RoomState {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.state
}

func (r *Room) SetState(statue shared.RoomState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = statue
}

func (r *Room) ReconnectPlayer(player shared.Player) error {
	slog.Info("[ReconnectPlayer]", "player", player)

	disconnectedPlayer := r.disconnectedPlayers[player.GetID()]
	if len(r.GetPlayers()) < 4 {
		player.SetHands(disconnectedPlayer.GetHands())
		r.players[disconnectedPlayer.GetID()] = player
		delete(r.disconnectedPlayers, disconnectedPlayer.GetID())
		player.StartHeartbeat()
	} else {
		return handle_errors.ErrRoomFull
	}
	slog.Info("[ReconnectPlayer]", "r.players", r.players)
	slog.Info("[ReconnectPlayer]", "r.disconnectedPlayers", r.disconnectedPlayers)

	// 发送当前游戏状态给重连的玩家
	message := NewMessage()
	message.SetMessage("game_state", map[string]interface{}{
		"current_turn": r.game.GetCurrentTurn(),
		"last_play":    r.game.GetLastPlayer(),
		"hand":         player.GetHands(),
	})
	player.GetConn().WriteJSON(message)

	// 通知其他玩家该玩家已重连
	r.Broadcast("player_reconnected", map[string]interface{}{
		"player_id": player.GetID(),
	})
	return nil
}

// 對 room 內的 player broadcast
func (r *Room) Broadcast(msgType string, content interface{}) {
	if r != nil && len(r.players) > 0 {
		slog.Info("[Room Broadcast]", "r.players", r.players)
		message := NewMessage()
		message.SetMessage(msgType, content)
		for _, player := range r.players {
			err := player.GetConn().WriteJSON(message)
			if err != nil {
				slog.Error("[Room Broadcast Error]", "broadcasting to player %s: %v", player.GetID(), "error", err)
			}
		}
	}
}
