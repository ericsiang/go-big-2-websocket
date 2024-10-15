package big2_game

type CardInterface interface {
	AnalyzeCards(cards []Card) (CardType, Card, error)
	CompareCard(cards1, cards2 []Card) (int, error)
}

type CardType int

type Suit int

// 建立 Card 結構體
type Card struct {
	Suit  Suit
	Value int
}
