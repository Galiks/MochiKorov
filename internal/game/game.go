package game

import (
	"fmt"
	"math/rand"
	"sort"
)

type Phase uint8

const (
	PhaseRoll Phase = iota
	PhaseIncome
	PhaseBuy
	PhaseEnd
)

type Game struct {
	Players       []*Player
	CurrentPlayer int
	Phase         Phase
	Market        []MarketCard
	Turn          int
	DiceResult    DiceResult
	DiceCount     int
	winner        *Player
}

type MarketCard struct {
	Card  Card `json:"card"`
	Count int  `json:"count"`
}

func NewGame(playerNames []string) *Game {
	players := make([]*Player, len(playerNames))
	for i, name := range playerNames {
		players[i] = NewPlayer(i, name)
	}

	g := &Game{
		Players:       players,
		CurrentPlayer: 0,
		Phase:         PhaseRoll,
		Market:        initMarket(),
		DiceCount:     1,
	}

	for _, p := range g.Players {
		if p.ID == 0 {
			continue
		}
		giveFreeCard(g.Market, "wheat_field", p)
	}

	return g
}

func giveFreeCard(market []MarketCard, id string, p *Player) {
	for i := range market {
		if market[i].Card.ID == id && market[i].Count > 0 {
			market[i].Count--
			p.Cards = append(p.Cards, market[i].Card)
			return
		}
	}
}

func initMarket() []MarketCard {
	cards := AllEstablishments()
	result := make([]MarketCard, 0, len(cards))
	for _, c := range cards {
		stock := c.DefaultStock
		if stock == 0 {
			stock = 6
		}
		result = append(result, MarketCard{Card: c, Count: int(stock)})
	}
	return result
}

func (g *Game) Current() *Player {
	if g.CurrentPlayer < 0 || g.CurrentPlayer >= len(g.Players) {
		return nil
	}
	return g.Players[g.CurrentPlayer]
}

func (g *Game) Roll() (DiceResult, error) {
	return g.RollWith(0)
}

func (g *Game) RollWith(preferred int) (DiceResult, error) {
	player := g.Current()
	count := 1
	if preferred == 2 && player.CanRollTwoDice() {
		count = 2
	} else if preferred == 0 && player.CanRollTwoDice() {
		count = 2
	}
	result, err := RollDice(count)
	if err != nil {
		return DiceResult{}, err
	}
	g.DiceCount = count
	g.DiceResult = result
	g.Phase = PhaseIncome
	return result, nil
}

func (g *Game) Reroll() (DiceResult, error) {
	count := g.DiceCount
	if count == 0 {
		count = 1
	}
	return g.RollWith(count)
}

func (g *Game) RerollDice(indices []int) (DiceResult, error) {
	player := g.Current()
	if !player.CanReroll() {
		return DiceResult{}, fmt.Errorf("player %s cannot reroll", player.Name)
	}
	if len(indices) == 0 {
		return g.RollWith(g.DiceCount)
	}
	for _, idx := range indices {
		if idx >= 0 && idx < len(g.DiceResult.Numbers) {
			g.DiceResult.Numbers[idx] = rand.Intn(6) + 1
		}
	}
	g.DiceResult.Sum = 0
	for _, n := range g.DiceResult.Numbers {
		g.DiceResult.Sum += n
	}
	return g.DiceResult, nil
}

func (g *Game) ActivateCards() []string {
	log := make([]string, 0)
	rollSum := g.DiceResult.Sum
	active := g.Current()

	for _, player := range g.Players {
		hasMall := player.HasShoppingMall()

		for _, card := range player.Cards {
			if !card.ActivatesOn(rollSum) {
				continue
			}
			if card.MinLandmark > 0 && uint8(player.CountLandmarks()) < card.MinLandmark {
				continue
			}
			if card.Condition != "" && !player.OwnsLandmarkNamed(card.Condition) {
				continue
			}

			// Цветовая активация
			if card.Color == ColorGreen && player.ID != active.ID {
				continue
			}
			if card.Color == ColorRed && player.ID == active.ID {
				continue
			}
			// Фиолетовые — только в ход владельца
			if card.Color == ColorPurple && player.ID != active.ID {
				continue
			}

			switch card.EffectType {
			case EffectFromBank:
				income := int(card.EffectValue)
				income = mallBonus(income, hasMall, card)
				player.AddMoney(income)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s", player.Name, income, card.Name))

			case EffectFromActive:
				income := int(card.EffectValue)
				total := active.RemoveMoney(income)
				if total < income {
					remaining := income - total
					n := len(g.Players)
					for offset := 1; offset < n && remaining > 0; offset++ {
						idx := (active.ID - offset + n) % n
						other := g.Players[idx]
						if other.ID == active.ID || other.ID == player.ID {
							continue
						}
						taken := other.RemoveMoney(remaining)
						total += taken
						remaining -= taken
					}
				}
				player.AddMoney(total)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s", player.Name, total, card.Name))

			case EffectPerPlayer:
				income := int(card.EffectValue)
				total := 0
				for _, other := range g.Players {
					if other.ID == player.ID {
						continue
					}
					total += other.RemoveMoney(income)
				}
				player.AddMoney(total)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (каждый платит по %d)", player.Name, total, card.Name, income))

			case EffectStealOne:
				if len(g.Players) <= 1 {
					continue
				}
				targets := make([]*Player, 0)
				for _, other := range g.Players {
					if other.ID != player.ID {
						targets = append(targets, other)
					}
				}
				target := targets[rand.Intn(len(targets))]
				income := int(card.EffectValue)
				actual := target.RemoveMoney(income)
				player.AddMoney(actual)
				log = append(log, fmt.Sprintf("%s крадёт %d монет(ы) у %s через %s", player.Name, actual, target.Name, card.Name))

			case EffectPerRanch:
				count := player.CountCardsByID("ranch")
				income := int(card.EffectValue) * count
				income = mallBonus(income, hasMall, card)
				player.AddMoney(income)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (x%d ранчо)", player.Name, income, card.Name, count))

			case EffectPerForest:
				count := player.CountCardsByID("forest")
				income := int(card.EffectValue) * count
				income = mallBonus(income, hasMall, card)
				player.AddMoney(income)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (x%d леса)", player.Name, income, card.Name, count))

			case EffectPerWheat:
				count := player.CountCardsByID("wheat_field")
				income := int(card.EffectValue) * count
				income = mallBonus(income, hasMall, card)
				player.AddMoney(income)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (x%d пшеницы)", player.Name, income, card.Name, count))

			case EffectPerPurple:
				count := 0
				for _, pc := range player.Cards {
					if pc.Color == ColorPurple {
						count++
					}
				}
				income := int(card.EffectValue) * count
				income = mallBonus(income, hasMall, card)
				player.AddMoney(income)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (x%d фиол.)", player.Name, income, card.Name, count))

			case EffectHalfOthers:
				total := 0
				for _, other := range g.Players {
					if other.ID == player.ID {
						continue
					}
					half := other.Money / 2
					total += other.RemoveMoney(half)
				}
				total = mallBonus(total, hasMall, card)
				player.AddMoney(total)
				log = append(log, fmt.Sprintf("%s получает %d монет(ы) от %s (половина)", player.Name, total, card.Name))
			}
		}
	}

	g.Phase = PhaseBuy
	return log
}

func mallBonus(base int, hasMall bool, card Card) int {
	if hasMall && (card.Color == ColorGreen || card.Color == ColorBlue) {
		return base + 1
	}
	return base
}

func (g *Game) BuyCardFromMarket(marketIndex int) error {
	player := g.Current()
	if marketIndex < 0 || marketIndex >= len(g.Market) {
		return fmt.Errorf("invalid market index")
	}
	item := &g.Market[marketIndex]
	if item.Count <= 0 {
		return fmt.Errorf("%s is out of stock", item.Card.Name)
	}
	if !player.CanAfford(item.Card.Price) {
		return fmt.Errorf("not enough money to buy %s", item.Card.Name)
	}
	item.Count--
	player.BuyCard(item.Card)
	g.Phase = PhaseEnd
	return nil
}

func allLandmarkDefs() []Card {
	return AllLandmarks()
}

func (g *Game) BuyLandmark(landmarkID string) error {
	player := g.Current()
	for _, lm := range allLandmarkDefs() {
		if lm.ID != landmarkID {
			continue
		}
		if player.OwnsLandmarkNamed(lm.Name) {
			return fmt.Errorf("%s already owns %s", player.Name, lm.Name)
		}
		if !player.CanAfford(lm.Price) {
			return fmt.Errorf("not enough money to buy %s", lm.Name)
		}
		player.BuyLandmark(lm)
		g.Phase = PhaseEnd
		return nil
	}
	return fmt.Errorf("unknown landmark: %s", landmarkID)
}

func (g *Game) AvailableLandmarks() []Card {
	result := make([]Card, 0)
	player := g.Current()
	for _, lm := range allLandmarkDefs() {
		if !player.OwnsLandmarkNamed(lm.Name) {
			result = append(result, lm)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Price < result[j].Price
	})
	return result
}

func (g *Game) AvailableMarket() []MarketCard {
	result := make([]MarketCard, 0)
	for _, item := range g.Market {
		if item.Count > 0 {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		colorOrder := map[CardColor]int{
			ColorBlue: 0, ColorGreen: 1, ColorRed: 2, ColorPurple: 3,
		}
		ci, cj := result[i].Card, result[j].Card
		if colorOrder[ci.Color] != colorOrder[cj.Color] {
			return colorOrder[ci.Color] < colorOrder[cj.Color]
		}
		if ci.Price != cj.Price {
		return ci.Price < cj.Price
	}
	return ci.Name < cj.Name
	})
	return result
}

func (g *Game) EndTurn() {
	g.Phase = PhaseRoll
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
	g.Turn++
}

func (g *Game) CheckWin() *Player {
	target := len(allLandmarkDefs())
	for _, p := range g.Players {
		count := 0
		for _, lm := range p.Landmarks {
			if lm.Price > 0 {
				count++
			}
		}
		if count >= target {
			g.winner = p
			return p
		}
	}
	return nil
}
