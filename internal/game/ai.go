package game

import "fmt"

type aiPhase int

const (
	aiEarly aiPhase = iota
	aiMid
	aiLate
)

func (g *Game) DoAITurn() []string {
	log := make([]string, 0)

	result, err := g.Roll()
	if err == nil {
		log = append(log, g.Current().Name+" бросает кубики: "+FormatDice(result))
	}

	if err == nil && g.Current().CanReroll() && result.Sum <= 4 {
		result, err = g.Reroll()
		if err == nil {
			log = append(log, g.Current().Name+" перебрасывает: "+FormatDice(result))
		}
	}

	actLog := g.ActivateCards()
	log = append(log, actLog...)

	player := g.Current()
	bought := false

	phase := aiCurrentPhase(player)

	if lm := g.aiPickLandmark(phase, player); lm != "" {
		err := g.BuyLandmark(lm)
		if err == nil {
			log = append(log, fmt.Sprintf("%s покупает достопримечательность: %s", player.Name, lmName(lm)))
			bought = true
		}
	}

	if !bought {
		if cardIdx := g.aiPickCard(phase, player); cardIdx >= 0 {
			err := g.BuyCardFromMarket(cardIdx)
			if err == nil {
				log = append(log, fmt.Sprintf("%s покупает карту: %s", player.Name, g.Market[cardIdx].Card.Name))
				bought = true
			}
		}
	}

	if !bought {
		log = append(log, fmt.Sprintf("%s ничего не покупает", player.Name))
		g.Phase = PhaseEnd
	}

	g.EndTurn()
	return log
}

func aiCurrentPhase(p *Player) aiPhase {
	switch {
	case p.CountLandmarks() < 3:
		return aiEarly
	case p.CountLandmarks() < 5:
		return aiMid
	default:
		return aiLate
	}
}

func (g *Game) aiPickLandmark(phase aiPhase, p *Player) string {
	avail := g.AvailableLandmarks()

	switch phase {
	case aiEarly:
		for _, lm := range avail {
			if p.CanAfford(lm.Price) && (lm.ID == "harbor_lm" || lm.ID == "train_station") {
				return lm.ID
			}
		}
	case aiMid:
		greenBlue := 0
		for _, c := range p.Cards {
			if c.Color == ColorGreen || c.Color == ColorBlue {
				greenBlue++
			}
		}
		if greenBlue >= 3 {
			for _, lm := range avail {
				if p.CanAfford(lm.Price) && lm.ID == "shopping_mall" {
					return lm.ID
				}
			}
		}
		for _, lm := range avail {
			if p.CanAfford(lm.Price) && lm.ID == "amusement_park" {
				return lm.ID
			}
		}
		best := ""
		bestPrice := -1
		for _, lm := range avail {
			if p.CanAfford(lm.Price) && int(lm.Price) > bestPrice {
				bestPrice = int(lm.Price)
				best = lm.ID
			}
		}
		return best

	case aiLate:
		best := ""
		bestPrice := -1
		for _, lm := range avail {
			if p.CanAfford(lm.Price) && int(lm.Price) > bestPrice {
				bestPrice = int(lm.Price)
				best = lm.ID
			}
		}
		return best
	}
	return ""
}

func (g *Game) aiPickCard(phase aiPhase, p *Player) int {
	market := g.AvailableMarket()

	switch phase {
	case aiEarly:
		for _, id := range []string{"wheat_field", "ranch", "bakery"} {
			if idx := g.marketIdx(id); idx >= 0 && p.CanAfford(g.Market[idx].Card.Price) {
				return idx
			}
		}
		best := -1
		bestPrice := -1
		for _, item := range market {
			if p.CanAfford(item.Card.Price) && (bestPrice == -1 || int(item.Card.Price) < bestPrice) {
				bestPrice = int(item.Card.Price)
				best = marketIdxByList(g.Market, item.Card.ID)
			}
		}
		return best

	case aiMid:
		if p.CountCardsByID("ranch") >= 2 {
			if idx := g.marketIdx("cheese_factory"); idx >= 0 && p.CanAfford(g.Market[idx].Card.Price) {
				return idx
			}
		}
		if p.CountCardsByID("forest") >= 2 {
			if idx := g.marketIdx("furniture_factory"); idx >= 0 && p.CanAfford(g.Market[idx].Card.Price) {
				return idx
			}
		}
		if p.CountCardsByID("wheat_field") >= 2 {
			if idx := g.marketIdx("fruit_market"); idx >= 0 && p.CanAfford(g.Market[idx].Card.Price) {
				return idx
			}
		}
		for _, id := range []string{"mine", "stadium", "tv_station", "sushi_bar"} {
			if idx := g.marketIdx(id); idx >= 0 && p.CanAfford(g.Market[idx].Card.Price) {
				return idx
			}
		}
		best := -1
		bestPrice := -1
		for _, item := range market {
			if p.CanAfford(item.Card.Price) && int(item.Card.Price) > bestPrice {
				bestPrice = int(item.Card.Price)
				best = marketIdxByList(g.Market, item.Card.ID)
			}
		}
		return best

	case aiLate:
		best := -1
		bestPrice := -1
		for _, item := range market {
			if p.CanAfford(item.Card.Price) && int(item.Card.Price) > bestPrice {
				bestPrice = int(item.Card.Price)
				best = marketIdxByList(g.Market, item.Card.ID)
			}
		}
		return best
	}
	return -1
}

func (g *Game) marketIdx(id string) int {
	for i := range g.Market {
		if g.Market[i].Card.ID == id && g.Market[i].Count > 0 {
			return i
		}
	}
	return -1
}

func marketIdxByList(market []MarketCard, id string) int {
	for i := range market {
		if market[i].Card.ID == id && market[i].Count > 0 {
			return i
		}
	}
	return -1
}

func lmName(id string) string {
	for _, lm := range allLandmarkDefs() {
		if lm.ID == id {
			return lm.Name
		}
	}
	return id
}

func FormatDice(result DiceResult) string {
	s := ""
	for i, n := range result.Numbers {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%d", n)
	}
	return s + fmt.Sprintf(" (сумма: %d)", result.Sum)
}
