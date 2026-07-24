package game

type Player struct {
	ID        int
	Name      string
	Money     int
	Cards     []Card
	Landmarks []Card
}

func NewPlayer(id int, name string) *Player {
	return &Player{
		ID:    id,
		Name:  name,
		Money: 3,
		Landmarks: []Card{
			{ID: "city_hall", Name: string(LandmarkCityHall), Price: 0, Type: TypeLandmark},
		},
	}
}

func (p *Player) HasLandmark(name LandmarkName) bool {
	for _, lm := range p.Landmarks {
		if lm.Name == string(name) {
			return true
		}
	}
	return false
}

func (p *Player) HasShoppingMall() bool {
	return p.HasLandmark(LandmarkShoppingMall)
}

func (p *Player) CanRollTwoDice() bool {
	return p.HasLandmark(LandmarkHarbor) || p.HasLandmark(LandmarkTrainStation)
}

func (p *Player) CanReroll() bool {
	return p.HasLandmark(LandmarkAmusementPark) || p.HasLandmark(LandmarkRadioTower) || p.HasLandmark(LandmarkAirport)
}

func (p *Player) CountLandmarks() int {
	count := 0
	for _, lm := range p.Landmarks {
		if lm.Price > 0 {
			count++
		}
	}
	return count
}

func (p *Player) CountCardsByIcon(icon CardIcon) int {
	count := 0
	for _, c := range p.Cards {
		if c.Icon == icon {
			count++
		}
	}
	return count
}

func (p *Player) CountCardsByID(id string) int {
	count := 0
	for _, c := range p.Cards {
		if c.ID == id {
			count++
		}
	}
	return count
}

func (p *Player) CanAfford(price uint8) bool {
	return p.Money >= int(price)
}

func (p *Player) BuyCard(card Card) {
	p.Money -= int(card.Price)
	p.Cards = append(p.Cards, card)
}

func (p *Player) BuyLandmark(card Card) {
	p.Money -= int(card.Price)
	p.Landmarks = append(p.Landmarks, card)
}

func (p *Player) AddMoney(amount int) {
	p.Money += amount
}

func (p *Player) RemoveMoney(amount int) int {
	if amount > p.Money {
		actual := p.Money
		p.Money = 0
		return actual
	}
	p.Money -= amount
	return amount
}

func (p *Player) OwnsLandmarkNamed(name string) bool {
	for _, lm := range p.Landmarks {
		if lm.Name == name {
			return true
		}
	}
	return false
}
