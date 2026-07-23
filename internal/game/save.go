package game

import (
	"encoding/json"
	"fmt"
	"os"
)

type SaveData struct {
	Players       []SavePlayer `json:"players"`
	CurrentPlayer int          `json:"current_player"`
	Market        []SaveMarket `json:"market"`
	Turn          int          `json:"turn"`
	Phase         uint8        `json:"phase"`
	DiceNumbers   []int        `json:"dice_numbers,omitempty"`
	DiceSum       int          `json:"dice_sum,omitempty"`
	DiceCount     int          `json:"dice_count,omitempty"`
}

type SavePlayer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Money     int    `json:"money"`
	Cards     []Card `json:"cards"`
	Landmarks []Card `json:"landmarks"`
}

type SaveMarket struct {
	Card  Card `json:"card"`
	Count int  `json:"count"`
}

func (g *Game) ToSaveData() *SaveData {
	sd := &SaveData{
		CurrentPlayer: g.CurrentPlayer,
		Turn:          g.Turn,
		Phase:         uint8(g.Phase),
		DiceNumbers:   g.DiceResult.Numbers,
		DiceSum:       g.DiceResult.Sum,
		DiceCount:     g.DiceCount,
	}
	for _, p := range g.Players {
		sd.Players = append(sd.Players, SavePlayer{
			ID:        p.ID,
			Name:      p.Name,
			Money:     p.Money,
			Cards:     p.Cards,
			Landmarks: p.Landmarks,
		})
	}
	for _, m := range g.Market {
		if m.Count > 0 {
			sd.Market = append(sd.Market, SaveMarket{Card: m.Card, Count: m.Count})
		}
	}
	return sd
}

func GameFromSaveData(sd *SaveData) *Game {
	players := make([]*Player, len(sd.Players))
	for i, sp := range sd.Players {
		players[i] = &Player{
			ID:        sp.ID,
			Name:      sp.Name,
			Money:     sp.Money,
			Cards:     sp.Cards,
			Landmarks: sp.Landmarks,
		}
	}

	full := initMarket()
	for fi := range full {
		full[fi].Count = 0
		for _, sm := range sd.Market {
			if full[fi].Card.ID == sm.Card.ID {
				full[fi].Count = sm.Count
				break
			}
		}
	}

	phase := PhaseRoll
	if sd.Phase < uint8(PhaseEnd+1) {
		phase = Phase(sd.Phase)
	}

	diceCount := sd.DiceCount
	if diceCount == 0 {
		diceCount = 1
	}

	return &Game{
		Players:       players,
		CurrentPlayer: sd.CurrentPlayer,
		Phase:         phase,
		Market:        full,
		Turn:          sd.Turn,
		DiceResult:    DiceResult{Numbers: sd.DiceNumbers, Sum: sd.DiceSum},
		DiceCount:     diceCount,
	}
}

func (g *Game) Save(path string) error {
	data, err := json.MarshalIndent(g.ToSaveData(), "", "  ")
	if err != nil {
		return fmt.Errorf("save marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadGame(path string) (*Game, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load read: %w", err)
	}

	var sd SaveData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("load unmarshal: %w", err)
	}

	return GameFromSaveData(&sd), nil
}
