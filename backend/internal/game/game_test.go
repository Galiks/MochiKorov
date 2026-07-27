package game

import (
	"testing"
)

func newTestGame() *Game {
	return NewGame([]PlayerDef{
		{Name: "Игрок", IsHuman: true},
		{Name: "Бот 1", IsHuman: false},
		{Name: "Бот 2", IsHuman: false},
	})
}

func setDice(g *Game, nums ...int) {
	g.DiceResult = DiceResult{Numbers: nums}
	g.DiceResult.Sum = 0
	for _, n := range nums {
		g.DiceResult.Sum += n
	}
	g.DiceCount = len(nums)
}

func TestNewGame(t *testing.T) {
	g := newTestGame()
	if len(g.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(g.Players))
	}
	if g.CurrentPlayer != 0 {
		t.Fatalf("expected player 0 to start, got %d", g.CurrentPlayer)
	}
	if g.Phase != PhaseRoll {
		t.Fatalf("expected PhaseRoll, got %v", g.Phase)
	}
	if g.Current().Money != 3 {
		t.Fatalf("expected player 0 to have 3 money, got %d", g.Current().Money)
	}
}

func TestNewGameBotsStartWithWheat(t *testing.T) {
	g := newTestGame()
	for _, p := range g.Players {
		if p.IsHuman {
			if len(p.Cards) != 0 {
				t.Fatalf("human should start with 0 cards, got %d", len(p.Cards))
			}
		} else {
			if len(p.Cards) != 1 || p.Cards[0].ID != "wheat_field" {
				t.Fatalf("bot %s should start with 1 wheat_field, got %v", p.Name, p.Cards)
			}
		}
	}
}

func TestMarketNonEmpty(t *testing.T) {
	g := newTestGame()
	if len(g.Market) == 0 {
		t.Fatal("market should not be empty")
	}
	for _, m := range g.Market {
		if m.Count <= 0 {
			t.Fatalf("market card %s should have positive stock", m.Card.ID)
		}
	}
}

func TestRollWithOneDie(t *testing.T) {
	g := newTestGame()
	result, err := g.RollWith(1)
	if err != nil {
		t.Fatalf("RollWith(1) failed: %v", err)
	}
	if len(result.Numbers) != 1 {
		t.Fatalf("expected 1 die, got %d", len(result.Numbers))
	}
	if result.Numbers[0] < 1 || result.Numbers[0] > 6 {
		t.Fatalf("die out of range: %d", result.Numbers[0])
	}
	if result.Sum != result.Numbers[0] {
		t.Fatalf("sum mismatch: %d vs %d", result.Sum, result.Numbers[0])
	}
	if g.Phase != PhaseIncome {
		t.Fatalf("expected PhaseIncome after roll, got %v", g.Phase)
	}
}

func TestRollWithTwoDiceFallsBackToOneWithoutLandmark(t *testing.T) {
	g := newTestGame()
	result, err := g.RollWith(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Numbers) != 1 {
		t.Fatalf("expected 1 die (fallback without landmark), got %d", len(result.Numbers))
	}
}

func TestRollWithTwoDiceWithLandmark(t *testing.T) {
	g := newTestGame()
	g.Current().Landmarks = append(g.Current().Landmarks,
		Card{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark},
	)
	result, err := g.RollWith(2)
	if err != nil {
		t.Fatalf("RollWith(2) with harbor failed: %v", err)
	}
	if len(result.Numbers) != 2 {
		t.Fatalf("expected 2 dice, got %d", len(result.Numbers))
	}
}

func TestInvalidDiceCount(t *testing.T) {
	g := newTestGame()
	_, err := g.RollWith(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivateCardsFromBank(t *testing.T) {
	g := newTestGame()
	g.Current().Cards = append(g.Current().Cards,
		Card{ID: "wheat_field", Name: "Пшеничное поле", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
	)
	setDice(g, 1)
	g.Phase = PhaseIncome

	moneyBefore := g.Current().Money
	log := g.ActivateCards()
	if len(log) == 0 {
		t.Fatal("expected activation log")
	}
	if g.Current().Money != moneyBefore+1 {
		t.Fatalf("expected +1 money, got %d -> %d", moneyBefore, g.Current().Money)
	}
	if g.Phase != PhaseBuy {
		t.Fatalf("expected PhaseBuy after activation, got %v", g.Phase)
	}
}

func TestActivateCardsGreenOnlyForOwner(t *testing.T) {
	g := newTestGame()
	human := g.Players[0]
	bot := g.Players[1]
	human.Cards = append(human.Cards,
		Card{ID: "bakery", Name: "Пекарня", Color: ColorGreen, Numbers: []int{2}, EffectType: EffectFromBank, EffectValue: 1},
	)
	bot.Cards = append(bot.Cards,
		Card{ID: "bakery", Name: "Пекарня", Color: ColorGreen, Numbers: []int{2}, EffectType: EffectFromBank, EffectValue: 1},
	)
	setDice(g, 2)
	g.Phase = PhaseIncome
	moneyHuman := human.Money
	moneyBot := bot.Money

	_ = g.ActivateCards()

	if human.Money != moneyHuman+1 {
		t.Fatalf("active player should get green card income, got %d -> %d", moneyHuman, human.Money)
	}
	if bot.Money != moneyBot {
		t.Fatalf("non-active player should NOT get green card income, got %d -> %d", moneyBot, bot.Money)
	}
}

func TestActivateCardsRedOnlyForOthers(t *testing.T) {
	g := newTestGame()
	human := g.Players[0]
	bot := g.Players[1]
	human.Cards = append(human.Cards,
		Card{ID: "convenience_store", Name: "Магазин", Color: ColorRed, Numbers: []int{2}, EffectType: EffectFromActive, EffectValue: 3},
	)
	bot.Cards = append(bot.Cards,
		Card{ID: "convenience_store", Name: "Магазин", Color: ColorRed, Numbers: []int{2}, EffectType: EffectFromActive, EffectValue: 3},
	)
	human.Money = 10
	bot.Money = 10
	setDice(g, 2)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	// Human's red card should NOT activate (active player's red cards don't trigger)
	// Bot's red card activates, takes 3 from the active player (human)
	if human.Money != 7 {
		t.Fatalf("active player should lose 3 to bot's red card, got %d (started at 10)", human.Money)
	}
	if bot.Money != 13 {
		t.Fatalf("bot should gain 3 from active player, got %d (started at 10)", bot.Money)
	}
}

func TestActivateCardsPerPlayer(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	g.Players[1].Money = 10
	g.Players[2].Money = 10
	player.Cards = append(player.Cards,
		Card{ID: "stadium", Name: "Стадион", Color: ColorPurple, Numbers: []int{4}, EffectType: EffectPerPlayer, EffectValue: 2},
	)
	setDice(g, 4)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	expected := 3 + 2*2
	if player.Money != expected {
		t.Fatalf("expected %d money, got %d (started at 3)", expected, player.Money)
	}
}

func TestActivateCardsPerRanch(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Cards = append(player.Cards,
		Card{ID: "ranch", Name: "Ранчо", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "ranch", Name: "Ранчо", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "cheese_factory", Name: "Сыроварня", Color: ColorGreen, Numbers: []int{5}, EffectType: EffectPerRanch, EffectValue: 3},
	)
	setDice(g, 5)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	if player.Money != 3+2*3 {
		t.Fatalf("expected %d money (3 start + 2*3 from 2 ranches), got %d", 3+2*3, player.Money)
	}
}

func TestActivateCardsPerForest(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Cards = append(player.Cards,
		Card{ID: "forest", Name: "Лес", Color: ColorBlue, Numbers: []int{3}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "forest", Name: "Лес", Color: ColorBlue, Numbers: []int{3}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "furniture_factory", Name: "Мебельная фабрика", Color: ColorGreen, Numbers: []int{5}, EffectType: EffectPerForest, EffectValue: 3},
	)
	setDice(g, 5)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	if player.Money != 3+2*3 {
		t.Fatalf("expected %d money, got %d", 3+2*3, player.Money)
	}
}

func TestActivateCardsPerWheat(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Cards = append(player.Cards,
		Card{ID: "wheat_field", Name: "Пшеничное поле", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "wheat_field", Name: "Пшеничное поле", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
		Card{ID: "fruit_market", Name: "Фруктовый рынок", Color: ColorGreen, Numbers: []int{6}, EffectType: EffectPerWheat, EffectValue: 3},
	)
	setDice(g, 6)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	if player.Money != 3+2*3 {
		t.Fatalf("expected %d money, got %d", 3+2*3, player.Money)
	}
}

func TestActivateCardsShoppingMallBonus(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Landmarks = append(player.Landmarks,
		Card{ID: "shopping_mall", Name: string(LandmarkShoppingMall), Price: 10, Type: TypeLandmark},
	)
	player.Cards = append(player.Cards,
		Card{ID: "wheat_field", Name: "Пшеничное поле", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
	)
	setDice(g, 1)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	if player.Money != 3+1+1 {
		t.Fatalf("expected %d money (3 start + 1 base + 1 mall), got %d", 3+1+1, player.Money)
	}
}

func TestActivateCardsNoActivationOnWrongNumber(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Cards = append(player.Cards,
		Card{ID: "wheat_field", Name: "Пшеничное поле", Color: ColorBlue, Numbers: []int{1}, EffectType: EffectFromBank, EffectValue: 1},
	)
	setDice(g, 2)
	g.Phase = PhaseIncome

	_ = g.ActivateCards()

	if player.Money != 3 {
		t.Fatalf("expected no change (wrong dice), got %d", player.Money)
	}
}

func TestBuyCardFromMarket(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Money = 10
	g.Phase = PhaseBuy

	idx := -1
	for i, m := range g.Market {
		if m.Card.ID == "bakery" && m.Count > 0 {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("bakery not found in market")
	}

	err := g.BuyCardFromMarket(idx)
	if err != nil {
		t.Fatalf("BuyCardFromMarket failed: %v", err)
	}

	if g.Market[idx].Count != 5 {
		t.Fatalf("expected stock to decrease by 1 (6->5), got %d", g.Market[idx].Count)
	}
	found := false
	for _, c := range player.Cards {
		if c.ID == "bakery" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("player should have bakery card")
	}
}

func TestBuyCardNotEnoughMoney(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Money = 0
	g.Phase = PhaseBuy

	for i, m := range g.Market {
		if m.Card.ID == "bakery" && m.Count > 0 {
			err := g.BuyCardFromMarket(i)
			if err == nil {
				t.Fatal("expected error for insufficient funds")
			}
			return
		}
	}
}

func TestBuyCardOutOfStock(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Money = 100
	g.Phase = PhaseBuy

	for i := range g.Market {
		g.Market[i].Count = 0
	}

	err := g.BuyCardFromMarket(0)
	if err == nil {
		t.Fatal("expected error for out of stock")
	}
}

func TestAvailableLandmarks(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	avail := g.AvailableLandmarks()

	if len(avail) == 0 {
		t.Fatal("expected some landmarks to be available")
	}

	player.BuyLandmark(avail[0])
	avail2 := g.AvailableLandmarks()
	for _, lm := range avail2 {
		if lm.ID == avail[0].ID {
			t.Fatal("bought landmark should not appear in available")
		}
	}
}

func TestBuyLandmark(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Money = 100
	g.Phase = PhaseBuy

	avail := g.AvailableLandmarks()
	if len(avail) == 0 {
		t.Fatal("no landmarks available")
	}

	err := g.BuyLandmark(avail[0].ID)
	if err != nil {
		t.Fatalf("BuyLandmark failed: %v", err)
	}

	if !player.OwnsLandmarkNamed(avail[0].Name) {
		t.Fatal("player should own purchased landmark")
	}
}

func TestBuyLandmarkDuplicate(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	player.Money = 100
	g.Phase = PhaseBuy

	avail := g.AvailableLandmarks()
	err := g.BuyLandmark(avail[0].ID)
	if err != nil {
		t.Fatalf("first BuyLandmark failed: %v", err)
	}

	err = g.BuyLandmark(avail[0].ID)
	if err == nil {
		t.Fatal("expected error when buying duplicate landmark")
	}
}

func TestEndTurn(t *testing.T) {
	g := newTestGame()
	g.CurrentPlayer = 0
	g.Phase = PhaseBuy
	turnBefore := g.Turn

	g.EndTurn()

	if g.Phase != PhaseRoll {
		t.Fatalf("expected PhaseRoll after EndTurn, got %v", g.Phase)
	}
	if g.CurrentPlayer != 1 {
		t.Fatalf("expected next player (1), got %d", g.CurrentPlayer)
	}
	if g.Turn != turnBefore+1 {
		t.Fatalf("expected turn %d, got %d", turnBefore+1, g.Turn)
	}
}

func TestEndTurnWrapAround(t *testing.T) {
	g := newTestGame()
	g.CurrentPlayer = 2
	g.Phase = PhaseBuy

	g.EndTurn()

	if g.CurrentPlayer != 0 {
		t.Fatalf("expected wrap to player 0, got %d", g.CurrentPlayer)
	}
}

func TestCheckWin(t *testing.T) {
	g := newTestGame()
	player := g.Current()
	allLm := AllLandmarks()
	for _, lm := range allLm {
		player.Landmarks = append(player.Landmarks, lm)
	}

	winner := g.CheckWin()
	if winner == nil {
		t.Fatal("expected a winner")
	}
	if winner.ID != player.ID {
		t.Fatalf("expected player %d to win, got %d", player.ID, winner.ID)
	}
}

func TestCheckWinNotYet(t *testing.T) {
	g := newTestGame()
	winner := g.CheckWin()
	if winner != nil {
		t.Fatal("expected no winner at game start")
	}
}

func TestRemoveMoneyPartial(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 3

	taken := p.RemoveMoney(5)
	if taken != 3 {
		t.Fatalf("expected to take 3 (all money), got %d", taken)
	}
	if p.Money != 0 {
		t.Fatalf("expected 0 money, got %d", p.Money)
	}
}

func TestRemoveMoneyFull(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	p.Money = 10

	taken := p.RemoveMoney(5)
	if taken != 5 {
		t.Fatalf("expected to take 5, got %d", taken)
	}
	if p.Money != 5 {
		t.Fatalf("expected 5 money, got %d", p.Money)
	}
}

func TestCanRerollWithLandmark(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	if p.CanReroll() {
		t.Fatal("should not be able to reroll without landmark")
	}

	p.Landmarks = append(p.Landmarks,
		Card{ID: "amusement_park", Name: string(LandmarkAmusementPark), Price: 16, Type: TypeLandmark},
	)
	if !p.CanReroll() {
		t.Fatal("should be able to reroll with amusement park")
	}
}

func TestCanRollTwoDiceWithLandmark(t *testing.T) {
	p := 	NewPlayer(0, "Test", true)
	if p.CanRollTwoDice() {
		t.Fatal("should not be able to roll 2 dice without landmark")
	}

	p.Landmarks = append(p.Landmarks,
		Card{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark},
	)
	if !p.CanRollTwoDice() {
		t.Fatal("should be able to roll 2 dice with harbor")
	}
}

func TestToSaveDataRoundTrip(t *testing.T) {
	g := newTestGame()
	g.Current().Money = 15
	g.Current().Cards = append(g.Current().Cards,
		Card{ID: "bakery", Name: "Пекарня", Color: ColorGreen, Numbers: []int{2}, EffectType: EffectFromBank, EffectValue: 1},
	)
	setDice(g, 3, 4)
	g.Phase = PhaseBuy

	sd := g.ToSaveData()
	restored := GameFromSaveData(sd)

	if restored.CurrentPlayer != g.CurrentPlayer {
		t.Fatalf("CurrentPlayer mismatch: %d vs %d", restored.CurrentPlayer, g.CurrentPlayer)
	}
	if restored.Turn != g.Turn {
		t.Fatalf("Turn mismatch: %d vs %d", restored.Turn, g.Turn)
	}
	if restored.Phase != g.Phase {
		t.Fatalf("Phase mismatch: %v vs %v", restored.Phase, g.Phase)
	}
	if restored.Current().Money != g.Current().Money {
		t.Fatalf("Money mismatch: %d vs %d", restored.Current().Money, g.Current().Money)
	}
	if len(restored.Current().Cards) != len(g.Current().Cards) {
		t.Fatalf("Cards count mismatch: %d vs %d", len(restored.Current().Cards), len(g.Current().Cards))
	}
	if restored.DiceResult.Sum != g.DiceResult.Sum {
		t.Fatalf("DiceSum mismatch: %d vs %d", restored.DiceResult.Sum, g.DiceResult.Sum)
	}
}

func TestFormatDice(t *testing.T) {
	result := DiceResult{Numbers: []int{3, 5}, Sum: 8}
	s := FormatDice(result)
	expected := "3, 5 (сумма: 8)"
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}

func TestSingleDieFormat(t *testing.T) {
	result := DiceResult{Numbers: []int{4}, Sum: 4}
	s := FormatDice(result)
	expected := "4 (сумма: 4)"
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}
