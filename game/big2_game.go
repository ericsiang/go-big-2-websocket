package game

import (
	"big2/big2_card"
	"big2/handle_errors"
	"big2/shared"
	"log/slog"
)

type Big2Game struct {
	Deck         *big2_card.Deck
	currentTurn  shared.Player
	lastPlayCard []big2_card.Card
	lastPlayer   shared.Player
	garbageCard  []map[string]big2_card.Card
	passes       int
	state        shared.GameState
	big2Card     *big2_card.Big2Card
}

func NewBig2Game() *Big2Game {
	big2Card := big2_card.NewBig2Card()
	return &Big2Game{
		state:    shared.GameStateWaiting,
		big2Card: &big2Card,
	}
}

func (g *Big2Game) Start() error {
	if g.state != shared.GameStateWaiting {
		return handle_errors.ErrInvalidGameState
	}
	g.state = shared.GameStatePlaying

	return nil
}

func (g *Big2Game) GetHandCards() ([][]big2_card.Card, *big2_card.GarbageCard) {
	newPlayerDeck, garbageCard := g.big2Card.NewDeck(52)
	return newPlayerDeck, garbageCard
}

func (g *Big2Game) CheckCard(action string, cards []big2_card.Card) error {
	if action == "first" {
		checked := g.big2Card.CheckFirstCard(cards)
		if !checked {
			slog.Warn("[CheckCard]","err",handle_errors.ErrFirstCardNotAllow.Error())
			return handle_errors.ErrFirstCardNotAllow
		}
	}
	_, _, err := g.big2Card.AnalyzeCards(cards)
	return err
}

func (g *Big2Game) PlayCards(shared.Player, []big2_card.Card) error {
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

func (g *Big2Game) SetCurrentTurn(player shared.Player) {
	g.currentTurn = player
}

func (g *Big2Game) GetLastPlayer() shared.Player {
	return g.lastPlayer
}

func (g *Big2Game) SetLastPlayer(player shared.Player) {
	g.lastPlayer = player
}
func (g *Big2Game) GetLastPlayerCard() []big2_card.Card {
	return g.lastPlayCard
}
func (g *Big2Game) SetLastPlayerCard(cards []big2_card.Card) {
	g.lastPlayCard = cards
}
