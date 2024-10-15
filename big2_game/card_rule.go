package big2_game

import (
	"fmt"
	"sort"
)

// 牌型 同花順> 鐵支> 葫蘆 > 順子> 三條> 對子> 單張
// 數字大小 2>A>K>Q>J>10>9>8>7>6>5>4>3

// 先比牌型，再比點數，最後花色

// 23456 > 10JQKA > 910JQK > 8910JQ > 78910J > 678910 > 56789 > 45678 > 34567 > A2345
// 1、23456為最大順，以2的花色作為比大小的依據
// 2、A2345為最小順，以5的花色作為比大小的依據
// 3、無JQKA2、QKA23、KA234這種順

// 花色
const (
	Plum   Suit = iota //梅花
	Block              //方塊
	Heart              //紅心
	Spades             //黑桃
)

// 牌型 同花順> 鐵支> 葫蘆 > 順子> 三條> 對子> 單張

const (
	None          CardType = iota
	Single                 //單張
	Pair                   //對子
	ThreeOfAKind           //三條
	Straight               //順子
	FullHouse              //葫蘆(3+2)
	FourOfAKind            //鐵支(4+1)
	StraightFlush          //同花順
)

const (
	Ace = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	J
	Q
	K
)

var (
	CardTypeStringSlice = []string{"None", "單張", "對子", "三條", "順子", "葫蘆", "鐵支", "同花順"}
	Suits               = []string{"梅花", "方塊", "紅心", "黑桃"}
)

type ErrInvalidCardCount struct {
	Msg string
}

func (e ErrInvalidCardCount) Error() string {
	return e.Msg
}

type ErrCompareFail struct {
	Msg string
}

func (e ErrCompareFail) Error() string {
	return e.Msg
}

// 建立 Big2Card 結構體，包含 Card
type Big2Card struct {
	Card
}

type GarbageCard struct {
	AlreadyUseCard []Card
}

// 產生牌組
func (c Big2Card) NewDeck(totalCount int) ([][]Card, *GarbageCard) {
	deck := &Deck{}
	alreadyUseCard := []Card{}
	garbageCard := &GarbageCard{AlreadyUseCard: alreadyUseCard}
	generateDeck := deck.GenerateDeck(totalCount)
	big2Deck := c.addBig2Deck(generateDeck)
	// fmt.Println("generateDeck:", big2Deck)
	deck.ShuffleDeck(big2Deck)
	newPlayerDeck := deck.DealDeck(4, big2Deck)
	return newPlayerDeck, garbageCard
}

// 產生Big2牌組
func (c Big2Card) addBig2Deck(generateDeck []Card) []Card {
	suits := []Suit{Plum, Block, Heart, Spades}
	values := []int{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, J, Q, K}
	for i, _ := range generateDeck {
		generateDeck[i] = Card{Suit: suits[i%4], Value: values[i%13]}
	}

	return generateDeck
}

// 判斷牌型
func (c Big2Card) AnalyzeCards(cards []Card) (handType CardType, highCard Card, err error) {
	if len(cards) > 5 {
		err := ErrInvalidCardCount{Msg: "牌型數量不正確"}
		return 0, Card{}, err
	}

	if c.isStraightFlush(cards) { // 判斷是否為同花順
		return StraightFlush, c.getStraightHighCard(cards), nil
	}
	if c.isFourOfAKind(cards) { // 判斷是否為鐵支
		return FourOfAKind, c.getSameCountHighCard(cards, 4), nil
	}
	if c.isFourOfAKind(cards) { // 判斷是否為鐵支
		return FourOfAKind, c.getSameCountHighCard(cards, 4), nil
	}
	if c.isFullHouse(cards) { // 判斷是否為葫蘆
		return FullHouse, c.getSameCountHighCard(cards, 3), nil
	}
	if c.isStraight(cards) { // 判斷是否為順子
		return Straight, c.getStraightHighCard(cards), nil
	}
	if c.isThreeOfAKind(cards) { // 判斷是否為三條
		return ThreeOfAKind, c.getSameCountHighCard(cards, 3), nil
	}
	if c.isPair(cards) { // 判斷是否為對子
		return Pair, c.getSameCountHighCard(cards, 2), nil
	}
	//為單張
	return Single, c.getSingleHighCard(cards), nil
}

// 比較牌組
func (c Big2Card) CompareCard(cards1, cards2 []Card) (int, error) {
	// 1. 判斷牌型
	handType1, highCard1, err := c.AnalyzeCards(cards1)
	if err != nil {
		fmt.Println("cards1:", err)
		return 0, err
	}
	handType2, highCard2, err := c.AnalyzeCards(cards2)
	if err != nil {
		fmt.Println("cards2:", err)
		return 0, err
	}

	fmt.Printf("handType1 牌型: %s, 最高牌: %s %d\n", CardTypeStringSlice[handType1], Suits[highCard1.Suit], highCard1.Value)
	fmt.Printf("handType2 牌型: %s, 最高牌: %s %d\n", CardTypeStringSlice[handType2], Suits[highCard2.Suit], highCard2.Value)

	// 2. 比較牌型
	if handType1 > handType2 {
		return 1, nil // cards1 勝
	}

	if handType1 < handType2 {
		return 2, nil // cards2 勝
	}

	// 牌型相同，比較點數，Two最大，Ace次之，其他依序遞減，同點比較花色
	// 將 Two 跟 Ace 調整點數
	highCardValue1 := c.getAdjustedCardValue(highCard1.Value)
	highCardValue2 := c.getAdjustedCardValue(highCard2.Value)
	// 比較點數
	if highCardValue1 > highCardValue2 {
		return 1, nil // cards1 勝
	}
	if highCardValue1 < highCardValue2 {
		return 2, nil // cards2 勝
	}

	// 同點比較花色
	if highCard1.Suit > highCard2.Suit {
		return 1, nil // cards1 勝
	} else if highCard1.Suit < highCard2.Suit {
		return 2, nil // cards2 勝
	}

	return 0, ErrCompareFail{Msg: "比較牌型失敗"}
}

// 取得調整後的點數
func (c Big2Card) getAdjustedCardValue(value int) int {
	if value == Two {
		return 15
	} else if value == Ace {
		return 14
	} else {
		return value
	}
}

// 判斷是否為同花順
func (c Big2Card) isStraightFlush(cards []Card) bool {
	return c.isStraight(cards) && c.isFlush(cards)
}

// 判斷是否為順子 (不包含 JQKA2, QKA23, KA234)
func (c Big2Card) isStraight(cards []Card) bool {
	if len(cards) != 5 {
		return false // 順子必須有五張牌
	}

	// 先排序
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Value < cards[j].Value
	})

	// 檢查是否有 A
	if cards[0].Value == Ace {
		// // A2345
		// if cards[1].Value == Two && cards[2].Value == Three && cards[3].Value == Four && cards[4].Value == Five {
		// 	return true
		// }

		// A10JQK
		if cards[1].Value == Ten && cards[2].Value == J && cards[3].Value == Q && cards[4].Value == K {
			return true
		}

		// 不包含 JQKA2, QKA23, KA234(A2345	除外皆不算)
		if cards[4].Value != Five {
			return false
		}
	}

	// 檢查是否連續
	for i := 1; i < len(cards); i++ {
		if cards[i].Value != cards[i-1].Value+1 {
			return false
		}
	}

	return true
}

// 判斷是否為同花
func (c Big2Card) isFlush(cards []Card) bool {
	suit := cards[0].Suit
	for _, card := range cards {
		if card.Suit != suit {
			return false
		}
	}
	return true
}

// 取得順子的 HighCard
func (c Big2Card) getStraightHighCard(cards []Card) Card {
	// 如果是 23456，則最高點數為 2
	if cards[0].Value == Two && cards[4].Value == Six {
		return cards[0]
	} else if cards[0].Value == Ace && cards[1].Value == Two && cards[4].Value == Five { // 如果是 A2345，則最高點數為 5
		return cards[4]
	} else if cards[0].Value == Ace && cards[1].Value == Ten && cards[4].Value == K { // 如果是 A10JQK，則最高點數為 A
		return cards[0]
	} else {
		// 其他返回最高的點數
		return cards[len(cards)-1]
	}
}

// 是否為鐵支
func (c Big2Card) isFourOfAKind(cards []Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}
	hashFour := false
	hasOne := false
	for _, count := range valueCounts {
		if count == 4 {
			hashFour = true
		} else if count == 1 {
			hasOne = true
		}
	}
	return hashFour && hasOne
}

// 取得鐵支、三條、對子的 HighCard
func (c Big2Card) getSameCountHighCard(cards []Card, checkCount int) Card {
	highValueCards := c.getHighValueCards(cards, checkCount)
	highCard := c.getSameValueHighSuit(highValueCards)
	return highCard
}

// 是否為葫蘆
func (c Big2Card) isFullHouse(cards []Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	hasThree := false
	hasTwo := false
	for _, count := range valueCounts {
		if count == 3 {
			hasThree = true
		} else if count == 2 {
			hasTwo = true
		}
	}

	return hasThree && hasTwo
}

// 是否為三條
func (c Big2Card) isThreeOfAKind(cards []Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}
	for _, count := range valueCounts {
		if count == 3 {
			return true
		}
	}
	return false
}

// 取得最高點數的相同牌組
func (c Big2Card) getHighValueCards(cards []Card, checkCount int) []Card {
	valueCounts := make(map[int][]Card)
	highValue := 0
	for _, card := range cards {
		valueCounts[card.Value] = append(valueCounts[card.Value], card)
		if len(valueCounts[card.Value]) == checkCount {
			if card.Value == Two { //2 最大
				highValue = Two
				return valueCounts[highValue]
			} else if card.Value == Ace { //接著是1
				highValue = Ace
			}

			if card.Value > highValue && highValue != Ace {
				highValue = card.Value
			}

		}
	}

	return valueCounts[highValue]
}

// 同點取得最高花色
func (c Big2Card) getSameValueHighSuit(cards []Card) Card {
	highCard := Card{}
	for _, card := range cards {
		if card.Suit > highCard.Suit {
			highCard.Suit = card.Suit
			highCard.Value = card.Value
		}
	}

	return highCard
}

// 是否為對子
func (c Big2Card) isPair(cards []Card) bool {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}
	for _, count := range valueCounts {
		if count == 2 {
			return true
		}
	}
	return false
}

// 取得單張的 HighCard
func (c Big2Card) getSingleHighCard(cards []Card) Card {
	highCard := Card{}
	for _, card := range cards {
		if card.Value == Two { //2 最大
			return card
		} else if card.Value == Ace { //接著是1
			highCard = card
		}

		if card.Value > highCard.Value && highCard.Value != Ace {
			highCard = card
		}
	}

	return highCard
}
