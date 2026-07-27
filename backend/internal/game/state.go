package game

type PlayerState struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Money            int    `json:"money"`
	Cards            []Card `json:"cards"`
	Landmarks        []Card `json:"landmarks"`
	IsCurrent        bool   `json:"is_current"`
	IsHuman          bool   `json:"is_human"`
	LandmarkCount    int    `json:"landmark_count"`
	CanRollTwoDice   bool   `json:"can_roll_two_dice"`
	CanReroll        bool   `json:"can_reroll"`
	ShoppingMall     bool   `json:"shopping_mall"`
}

type GameStateResponse struct {
	SessionID          string         `json:"session_id"`
	Turn               int            `json:"turn"`
	Phase              string         `json:"phase"`
	CurrentPlayer      PlayerState    `json:"current_player"`
	Players            []PlayerState  `json:"players"`
	Market             []MarketCard   `json:"market"`
	AvailableLandmarks []Card         `json:"available_landmarks"`
	Dice               *DiceResult    `json:"dice,omitempty"`
	Log                []string       `json:"log,omitempty"`
	Winner             *PlayerState   `json:"winner,omitempty"`
	CanRoll            bool           `json:"can_roll"`
	CanReroll          bool           `json:"can_reroll"`
	CanBuy             bool           `json:"can_buy"`
	GameOver           bool           `json:"game_over"`
	TotalLandmarks     int            `json:"total_landmarks"`
	YourID             int            `json:"your_id"`
}

func toPlayerState(p *Player, currentID int) PlayerState {
	return PlayerState{
		ID:             p.ID,
		Name:           p.Name,
		Money:          p.Money,
		Cards:          p.Cards,
		Landmarks:      p.Landmarks,
		IsCurrent:      p.ID == currentID,
		IsHuman:        p.IsHuman,
		LandmarkCount:  p.CountLandmarks(),
		CanRollTwoDice: p.CanRollTwoDice(),
		CanReroll:      p.CanReroll(),
		ShoppingMall:   p.HasShoppingMall(),
	}
}

func (g *Game) ToStateResponse(sessionID string, log []string, yourID int) *GameStateResponse {
	players := make([]PlayerState, len(g.Players))
	for i, p := range g.Players {
		players[i] = toPlayerState(p, g.CurrentPlayer)
	}

	current := toPlayerState(g.Current(), g.CurrentPlayer)

	var winner *PlayerState
	if g.winner != nil {
		ws := toPlayerState(g.winner, g.CurrentPlayer)
		winner = &ws
	}

	phase := "roll"
	switch g.Phase {
	case PhaseRoll:
		phase = "roll"
	case PhaseIncome:
		phase = "income"
	case PhaseBuy:
		phase = "buy"
	case PhaseEnd:
		phase = "end"
	}

	isHuman := g.Current().IsHuman
	canRoll := g.Phase == PhaseRoll && isHuman
	canReroll := g.Phase == PhaseIncome && isHuman && g.Current().CanReroll() && g.DiceResult.Sum > 0
	canBuy := g.Phase == PhaseBuy && isHuman && winner == nil

	dice := &g.DiceResult
	if len(g.DiceResult.Numbers) == 0 {
		dice = nil
	}

	return &GameStateResponse{
		SessionID:          sessionID,
		Turn:               g.Turn,
		Phase:              phase,
		CurrentPlayer:      current,
		Players:            players,
		Market:             g.AvailableMarket(),
		AvailableLandmarks: g.AvailableLandmarks(),
		Dice:               dice,
		Log:                log,
		Winner:             winner,
		CanRoll:            canRoll,
		CanReroll:          canReroll,
		CanBuy:             canBuy,
		GameOver:           winner != nil,
		TotalLandmarks:     len(AllLandmarks()),
		YourID:             yourID,
	}
}
