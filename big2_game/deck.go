package big2_game

import (
	"math/rand"
)

type Deck struct {
	Cards []interface{}
}

func (d *Deck) NewDeck(totalCount int) [][]Card {
	// 產生牌組
	generateDeck := d.GenerateDeck(totalCount)
	// 洗牌
	d.ShuffleDeck(generateDeck)
	// 發牌
	return d.DealDeck(4, generateDeck)
}

// 產生空牌組
func (d *Deck) GenerateDeck(totalCount int) []Card {
	deck := make([]Card, 0)
	for i := 0; i < totalCount; i++ {
		deck = append(deck, Card{})
	}

	return deck
}

// 洗牌
func (d *Deck) ShuffleDeck(deck []Card) {
	// for i := len(deck) - 1; i > 0; i-- {
	// 	j := rand.Intn(i + 1)
	// 	deck[i], deck[j] = deck[j], deck[i]
	// }

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// 發牌
func (d *Deck) DealDeck(numPlayers int, deck []Card) [][]Card {
	// 計算每個玩家可以獲得的牌數
	cardsPerPlayer := len(deck) / numPlayers

	// 儲存每個玩家的手牌
	playerHands := make([][]Card, 4)

	// 發牌
	for i := 0; i < numPlayers; i++ {
		// 直接從牌組中取出指定数量的牌，作為玩家的手牌
		playerHands[i] = deck[i*cardsPerPlayer : (i+1)*cardsPerPlayer]
	}
	return playerHands
}
