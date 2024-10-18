package big2_game

import "big2/shared"

type Big2Game struct {
	Deck         *Deck
	currentTurn  shared.Player
	LastPlayCard []shared.Card
	LastPlayer   shared.Player
	GarbageCard  *GarbageCard
	Passes       int
	state        shared.GameState
}

func NewBig2Game() *Big2Game {
	return &Big2Game{
		state: shared.GameStateWaiting,
	}
}

func (g *Big2Game) Start() {
	g.state = shared.GameStatePlaying
	// players := shared.Room.GetPlayers()
	// deck := big2.NewDeck(52)

	// Deal cards
	return
}

func (g *Big2Game) PlayCards(shared.Player, []shared.Card) error {
	return nil
}

func (g *Big2Game) Pass(player shared.Player) error {
	return nil
}

func (g *Big2Game) GetState() shared.GameState {
	return g.state
}

func (g *Big2Game) GetCurrentTurn() shared.Player {
	return g.currentTurn
}
