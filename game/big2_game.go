package game

import (
	"big2/big2_card"
	"big2/handle_errors"
	"big2/room"
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

type playCards struct {
	Message string           `json:"message"`
	Cards   []big2_card.Card `json:"cards"`
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

func (g *Big2Game) CheckAndPlayFirstCard(cards []big2_card.Card, player shared.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentPlayer := g.gameSortPlayer[g.currentTurn%4]
	if currentPlayer.GetID() != player.GetID() {
		return handle_errors.ErrNotPlayerTurn
	}

	checked := g.big2Card.CheckFirstCard(cards)
	if !checked {
		slog.Warn("[CheckCard]", "err", handle_errors.ErrFirstCardNotAllow.Error())
		return handle_errors.ErrFirstCardNotAllow
	}

	// 檢查是否為手牌
	handCards := player.GetHands()
	handCardCountCheck := len(cards)
	count := 0
	for _, card := range cards {
		for _, handCard := range handCards {
			if card.Suit == handCard.Suit && card.Value == handCard.Value {
				count++
			}
		}
	}
	if handCardCountCheck != count {
		return handle_errors.ErrHandsCardNotFound
	}

	_, _, err := g.big2Card.AnalyzeCards(cards)
	if err != nil {
		return err
	}
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
	outhandCards := make(map[string][]big2_card.Card)
	outhandCards[player.GetID()] = cards
	g.garbageCard = append(g.garbageCard, outhandCards)
	g.lastPlayCard = cards
	g.lastPlayer = player
	g.currentTurn = g.currentTurn + 1
	g.passes = 0
	slog.Info("[CheckAndPlayFirstCard]", "currentTurn", g.currentTurn)
	content := playCards{player.GetID() + "出牌了！", cards}
	player.GetRoom().Broadcast("play_cards_broadcast", content)
	return nil
}

func (g *Big2Game) CompareCards(card1, card2 []big2_card.Card) error {
	// 1 => card1 win , 2 => card2 win
	win, err := g.big2Card.CompareCard(card1, card2)
	if err != nil {
		return err
	}
	if win != 2 {
		return handle_errors.ErrCardNotBigger
	}
	return nil
}

func (g *Big2Game) PlayCards(action string, player shared.Player, cards []big2_card.Card) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentPlayer := g.gameSortPlayer[g.currentTurn%4]
	if currentPlayer.GetID() != player.GetID() {
		return handle_errors.ErrNotPlayerTurn
	}

	if g.state == shared.GameStateWaiting {
		return handle_errors.ErrInvalidGameState
	}

	// 檢查是否為手牌
	handCards := player.GetHands()
	handCardCountCheck := len(cards)
	count := 0
	for _, card := range cards {
		for _, handCard := range handCards {
			if card.Suit == handCard.Suit && card.Value == handCard.Value {
				count++
			}
		}
	}
	if handCardCountCheck != count {
		return handle_errors.ErrHandsCardNotFound
	}

	// 當是先手或其他人都 pass 時 ，只檢查牌型
	if action == "first" || g.passes == 3 {
		_, _, err := g.big2Card.AnalyzeCards(cards)
		if err != nil {
			return err
		}
	} else {
		// 比對牌型
		err := g.CompareCards(g.lastPlayCard, cards)
		if err != nil {
			return handle_errors.ErrCardNotBigger
		}
	}

	for _, card := range cards {
		for i, handCard := range handCards {
			if card.Suit == handCard.Suit && card.Value == handCard.Value {
				// 從手牌移除要出的牌
				player.RemoveHands(i)
				break
			}
		}
	}
	slog.Info("[PlayCards]", "card", player.GetHands())
	// 要出的牌加入到 garbageCard
	outhandCards := make(map[string][]big2_card.Card)
	outhandCards[player.GetID()] = cards
	g.garbageCard = append(g.garbageCard, outhandCards)
	g.lastPlayCard = cards
	g.lastPlayer = player
	// 有出牌要歸 0
	g.passes = 0
	// 手牌沒了，該玩家獲勝
	if handCards := player.GetHands(); len(handCards) == 0 {
		g.state = shared.GameStateEnded
		player.GetRoom().Broadcast("play_cards_broadcast", player.GetID()+" 獲勝了！")
		return nil
	}
	g.currentTurn = g.currentTurn + 1
	slog.Info("[PlayCards]", "currentTurn", g.currentTurn)
	content := playCards{player.GetID() + "出牌了！", cards}
	player.GetRoom().Broadcast("play_cards_broadcast", content)
	return nil
}

func (g *Big2Game) Pass(player shared.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	slog.Info("[Pass]", "player", player.GetID())

	if g.state != shared.GameStatePlaying {
		return handle_errors.ErrInvalidGameState
	}
	g.lastPlayer = player
	g.passes = g.passes + 1
	g.currentTurn = g.currentTurn + 1
	player.GetRoom().Broadcast("game_action_pass_broadcast", player.GetID()+" 沒牌出 Pass 了！")
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
func (g *Big2Game) SetGameSortPlayer(key int, player shared.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	slog.Info("[SetGameSortPlayer]", "g.gameSortPlayer len", len(g.gameSortPlayer))
	g.gameSortPlayer = append(g.gameSortPlayer, player)
}
func (g *Big2Game) NextPlayer(player shared.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	slog.Info("[NextPlayer]", "currentTurn", g.currentTurn)
	// 手牌沒了，該玩家獲勝
	if handCards := player.GetHands(); len(handCards) == 0 {
		g.state = shared.GameStateEnded
		player.GetRoom().Broadcast("play_cards_broadcast", player.GetID()+" 獲勝了！")
		return
	}

	nextPlayer := g.gameSortPlayer[g.currentTurn%4]
	content := "輪到( " + nextPlayer.GetID() + " ) 出牌了！"
	nextPlayer.GetRoom().Broadcast("game_action_broadcast", content)

	message := room.NewMessage()
	message.SetMessage("game_action_next_play", "到你出牌了！")
	nextPlayer.GetConn().WriteJSON(message)
}
