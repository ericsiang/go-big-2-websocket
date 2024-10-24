package shared

import (
	"big2/big2_card"

	"github.com/gorilla/websocket"
)

type PlayerState int

const (
	PlayerStateNotReady PlayerState = iota
	PlayerStateReady
	PlayerStatePlaying
	PlayerStateDisconnected
)

type RoomState int

const (
	RoomStateWaiting RoomState = iota
	RoomStateInGame
	RoomStateEnded
)

type GameState int

const (
	GameStateWaiting GameState = iota
	GameStatePlaying
	GameStateEnded
)

type Game interface {
	Start() error
	GetHandCards() ([][]big2_card.Card, *big2_card.GarbageCard)
	SetCurrentTurn(Player)
	CheckCard(string, []big2_card.Card) error
	PlayCards(Player, []big2_card.Card) error
	Pass(Player) error
	GetState() GameState
	GetCurrentTurn() Player
	GetLastPlayer() Player
	SetLastPlayer(Player)
	GetLastPlayerCard() []big2_card.Card
	SetLastPlayerCard([]big2_card.Card)
	// GetWinner() Player
}

type Player interface {
	GetID() string
	GetConn() *websocket.Conn
	SetConn(*websocket.Conn)
	GetHands() []big2_card.Card
	SetHands([]big2_card.Card)
	GetState() PlayerState
	SetState(PlayerState)
	GetRoom() Room
	SetRoom(Room)
	GetGameSort() int
	SetGameSort(int)
	Disconnect()
	StartHeartbeat()
}

type Room interface {
	GetID() string
	GetGame() Game
	StartGame()
	AddPlayer(Player)
	RemovePlayer(string)
	GetPlayer(string) (Player, bool)
	GetPlayers() map[string]Player
	GetDisconnectedPlayers() map[string]Player
	SetDisconnectedPlayer(Player)
	RemoveDisconnectedPlayer(Player)
	GetState() RoomState
	SetState(RoomState)
	// GetLastActivity() time.Time
	// UpdateLastActivity()
	ReconnectPlayer(Player) error
	Broadcast(string, interface{})
}

type Message interface {
	SetMessage(string, interface{})
}
