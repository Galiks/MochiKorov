package game

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	p := NewPlayer(0, "Тест", true)
	if p.ID != 0 {
		t.Fatalf("expected ID 0, got %d", p.ID)
	}
	if p.Name != "Тест" {
		t.Fatalf("expected name Тест, got %s", p.Name)
	}
	if p.Money != 3 {
		t.Fatalf("expected 3 starting money, got %d", p.Money)
	}
	if len(p.Landmarks) != 1 || p.Landmarks[0].ID != "city_hall" {
		t.Fatal("player should start with city hall landmark")
	}
	if len(p.Cards) != 0 {
		t.Fatal("player should start with no cards")
	}
}

func TestAddMoney(t *testing.T) {
	p := NewPlayer(0, "Test", true)
	p.AddMoney(5)
	if p.Money != 8 {
		t.Fatalf("expected 8 (3+5), got %d", p.Money)
	}
}

func TestRemoveMoneyEnough(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 10
	taken := p.RemoveMoney(4)
	if taken != 4 {
		t.Fatalf("expected 4 taken, got %d", taken)
	}
	if p.Money != 6 {
		t.Fatalf("expected 6 remaining, got %d", p.Money)
	}
}

func TestRemoveMoneyNotEnough(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 3
	taken := p.RemoveMoney(10)
	if taken != 3 {
		t.Fatalf("expected 3 taken (all), got %d", taken)
	}
	if p.Money != 0 {
		t.Fatalf("expected 0 remaining, got %d", p.Money)
	}
}

func TestRemoveMoneyZero(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 0
	taken := p.RemoveMoney(5)
	if taken != 0 {
		t.Fatalf("expected 0 taken, got %d", taken)
	}
	if p.Money != 0 {
		t.Fatalf("expected 0 remaining, got %d", p.Money)
	}
}

func TestCanAfford(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 5

	if !p.CanAfford(5) {
		t.Fatal("should afford exactly 5")
	}
	if !p.CanAfford(3) {
		t.Fatal("should afford 3 when having 5")
	}
	if p.CanAfford(6) {
		t.Fatal("should not afford 6 when having 5")
	}
}

func TestCanAffordZero(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 0

	if p.CanAfford(1) {
		t.Fatal("should not afford anything with 0 money")
	}
}

func TestBuyCard(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 10
	card := Card{ID: "bakery", Name: "Пекарня", Price: 1}

	p.BuyCard(card)

	if p.Money != 9 {
		t.Fatalf("expected 9 money (10-1), got %d", p.Money)
	}
	if len(p.Cards) != 1 || p.Cards[0].ID != "bakery" {
		t.Fatal("player should have bakery card")
	}
}

func TestPlayerBuyLandmark(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 20
	lm := Card{ID: "shopping_mall", Name: string(LandmarkShoppingMall), Price: 10, Type: TypeLandmark}

	p.BuyLandmark(lm)

	if p.Money != 10 {
		t.Fatalf("expected 10 money (20-10), got %d", p.Money)
	}
	found := false
	for _, l := range p.Landmarks {
		if l.ID == "shopping_mall" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("player should have shopping mall landmark")
	}
}

func TestHasShoppingMall(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	if p.HasShoppingMall() {
		t.Fatal("should not have shopping mall initially")
	}

	p.Landmarks = append(p.Landmarks,
		Card{ID: "shopping_mall", Name: string(LandmarkShoppingMall), Price: 10, Type: TypeLandmark},
	)
	if !p.HasShoppingMall() {
		t.Fatal("should have shopping mall")
	}
}

func TestCountLandmarks(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	if p.CountLandmarks() != 0 {
		t.Fatalf("expected 0 paid landmarks (city hall is free), got %d", p.CountLandmarks())
	}

	p.Landmarks = append(p.Landmarks,
		Card{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark},
	)
	if p.CountLandmarks() != 1 {
		t.Fatalf("expected 1 landmark, got %d", p.CountLandmarks())
	}
}

func TestCountCardsByID(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Cards = []Card{
		{ID: "wheat_field"},
		{ID: "bakery"},
		{ID: "wheat_field"},
	}

	if p.CountCardsByID("wheat_field") != 2 {
		t.Fatalf("expected 2 wheat_field, got %d", p.CountCardsByID("wheat_field"))
	}
	if p.CountCardsByID("bakery") != 1 {
		t.Fatalf("expected 1 bakery, got %d", p.CountCardsByID("bakery"))
	}
	if p.CountCardsByID("forest") != 0 {
		t.Fatalf("expected 0 forest, got %d", p.CountCardsByID("forest"))
	}
}

func TestOwnsLandmarkNamed(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	if p.OwnsLandmarkNamed(string(LandmarkHarbor)) {
		t.Fatal("should not own harbor initially")
	}

	p.Landmarks = append(p.Landmarks,
		Card{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark},
	)
	if !p.OwnsLandmarkNamed(string(LandmarkHarbor)) {
		t.Fatal("should own harbor after purchase")
	}
}

func TestCanRollTwoDiceTrainStation(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Landmarks = append(p.Landmarks,
		Card{ID: "train_station", Name: string(LandmarkTrainStation), Price: 4, Type: TypeLandmark},
	)
	if !p.CanRollTwoDice() {
		t.Fatal("train station should allow 2 dice")
	}
}

func TestCanRerollRadioTower(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Landmarks = append(p.Landmarks,
		Card{ID: "radio_tower", Name: string(LandmarkRadioTower), Price: 22, Type: TypeLandmark},
	)
	if !p.CanReroll() {
		t.Fatal("radio tower should allow reroll")
	}
}

func TestCanRerollAirport(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Landmarks = append(p.Landmarks,
		Card{ID: "airport", Name: string(LandmarkAirport), Price: 30, Type: TypeLandmark},
	)
	if !p.CanReroll() {
		t.Fatal("airport should allow reroll")
	}
}
