package game

import (
	"big2/big2_card"
	"big2/handle_errors"
	"big2/shared"
	"log/slog"
	"sync"
)

type Big2Game struct {
	Deck           *big2_card.Deck
	mu             sync.Mutex
	currentTurn    int
	lastPlayCard   []big2_card.Card
	lastPlayer     shared.Player
	garbageCard    []map[string][]big2_card.Card
	passes         int
	state          shared.GameState
	big2Card       *big2_card.Big2Card
	gameSortPlayer []shared.Player
}

func NewBig2Game() *Big2Game {
	big2Card := big2_card.NewBig2Card()
	return &Big2Game{
		state:          shared.GameStateWaiting,
		big2Card:       &big2Card,
		currentTurn:    -1,
		gameSortPlayer: make([]shared.Player, 0, 4),
	}
}

func (g *Big2Game) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

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
			slog.Warn("[CheckCard]", "err", handle_errors.ErrFirstCardNotAllow.Error())
			return handle_errors.ErrFirstCardNotAllow
		}
	}
	_, _, err := g.big2Card.AnalyzeCards(cards)
	return err
}

func (g *Big2Game) PlayCards(player shared.Player, cards []big2_card.Card) error {
	handCards := player.GetHands()
	for _, card := range cards {
		for i, handCard := range handCards {
			if card.Suit == handCard.Suit && card.Value == handCard.Value {
				// 從手牌移除要出的牌
				player.RemoveHands(i)
				break
			}
		}
	}
	// 要出的牌加入到 garbageCard
	value := make(map[string][]big2_card.Card)
	value[player.GetID()] = cards
	g.garbageCard = append(g.garbageCard, value)
	return nil
}

func (g *Big2Game) Pass(player shared.Player) error {
	return nil
}

func (g *Big2Game) GetState() shared.GameState {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.state
}

func (g *Big2Game) GetCurrentTurn() int {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.currentTurn
}

func (g *Big2Game) SetCurrentTurn(currentTurn int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.currentTurn = currentTurn
}

func (g *Big2Game) GetLastPlayer() shared.Player {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.lastPlayer
}

func (g *Big2Game) SetLastPlayer(player shared.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.lastPlayer = player
}
func (g *Big2Game) GetLastPlayerCard() []big2_card.Card {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.lastPlayCard
}
func (g *Big2Game) SetLastPlayerCard(cards []big2_card.Card) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.lastPlayCard = cards
}
func (g *Big2Game) GetGameSortPlayer() []shared.Player {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.gameSortPlayer
}
func (g *Big2Game) SetGameSortPlayer(key int, value shared.Player) {
	g.gameSortPlayer[key] = value
}
func (g *Big2Game) GetNextPlayer() shared.Player {
	nextCurrentTurn := g.GetCurrentTurn() + 1
	g.SetCurrentTurn(nextCurrentTurn)
	return g.gameSortPlayer[nextCurrentTurn%4]
}
