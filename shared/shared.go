package shared

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

type Card interface {
	GetRank() int
	GetSuit() int
	String() string
}

type Player interface {
	GetID() string
	GetHand() []Card
	SetHand([]Card)
	GetState() PlayerState
	SetState(PlayerState)
	Disconnect()
	Reconnect() error
}

type Room interface {
	GetID() string
	Lock()
	Unlock()
	AddPlayer(Player) error
	RemovePlayer(string) error
	GetPlayer(string) (Player, error)
	GetPlayers() []Player
	GetDisconnectedPlayers() []Player
	GetState() RoomState
	SetState(RoomState)
	// GetLastActivity() time.Time
	// UpdateLastActivity()

	Broadcast(Message)
}

type Message interface {
	SetMessage(string, interface{})
}

type Game interface {
	Start()
	PlayCards(Player, []Card) error
	Pass(Player) error
	GetState() GameState
	GetCurrentTurn() Player
	// GetWinner() Player
}
