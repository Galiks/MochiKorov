package models

import "fmt"

type PlayerTurn struct {
	TotalResult       uint8
	FirstResultCubes  CubeResult
	RerollCubeIndexes []uint8
}

type Player struct {
	Turn  PlayerTurn
	ID    uint8
	Name  string
	Cards HandCard
	Money uint8
}

func NewPlayer(id uint8, name string) *Player {
	return &Player{
		ID:   id,
		Name: name,
		Cards: HandCard{
			Card{
				ID:   "",
				Name: string(CityHallAttractionName),
			},
		},
		Money: 3,
	}
}

// Первый бросок кубиков игрока
func (u *Player) RollCubes(cubCount uint8) (uint8, error) {
	cubes, cubeSum, err := u.roll(cubCount)
	if err != nil {
		return 0, err
	}
	u.Turn.FirstResultCubes = cubes
	u.Turn.TotalResult = cubeSum
	return cubeSum, nil
}

// игрок может перебросить 1 или несколько кубиков.
func (u *Player) RerollCubes(indexes []uint8) (uint8, error) {
	// if !u.isCanReroll() {
	// 	return 0, fmt.Errorf("user %s cannot reroll cubes", u.Name)
	// }
	var rerollRelust uint8 = 0
	for i := range indexes {
		if indexes[i] >= uint8(len(u.Turn.FirstResultCubes.Numbers)) {
			return 0, fmt.Errorf("invalid index for reroll: %d", indexes[i])
		}
		rerollRelust += u.Turn.FirstResultCubes.Numbers[indexes[i]]
	}
	_, cubeSum, err := u.roll(uint8(len(indexes)))
	if err != nil {
		return 0, err
	}
	u.Turn.TotalResult = u.Turn.TotalResult - rerollRelust + cubeSum
	u.Turn.RerollCubeIndexes = indexes
	return cubeSum, nil
}

func (u *Player) AddTwoToCubeResult() (uint8, error) {
	if !u.isCanAddNumberToCube() {
		return 0, fmt.Errorf("user %s cannot add number to cube", u.Name)
	}
	u.Turn.TotalResult += 2
	return u.Turn.TotalResult, nil
}

func (u *Player) roll(cubCount uint8) (CubeResult, uint8, error) {
	cubes, err := GetCubeNumber(cubCount)
	if err != nil {
		return CubeResult{}, 0, err
	}
	cubeSum := uint8(0)
	for _, num := range cubes.Numbers {
		cubeSum += num
	}
	return cubes, cubeSum, nil
}

func (u *Player) isCanReroll() bool {
	return u.Cards.IsHasCardByNameAndType(string(RadioTowerAttractionName), AttractionsEng)
}

func (u *Player) isCanAddNumberToCube(cubeResult uint8) bool {
	return u.Cards.IsHasCardByNameAndType(string(HarborAttractionName), AttractionsEng)
}

func (u *Player) GetMoney() uint8 {
	return u.Money
}

func (u *Player) GetCards() []Card {
	return u.Cards
}

func (u *Player) AddCard(card Card) {
	u.Cards = append(u.Cards, card)
}
