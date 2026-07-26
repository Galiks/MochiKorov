package game

import (
	"errors"
	"sync"
)

type CardIcon string

const (
	IconWheat   CardIcon = "Пшеница"
	IconRanch   CardIcon = "Ранчо"
	IconBakery  CardIcon = "Пекарня"
	IconCafe    CardIcon = "Кафе"
	IconShop   CardIcon = "Магазин"
	IconForest  CardIcon = "Лес"
	IconRest    CardIcon = "Ресторан"
	IconFactory CardIcon = "Фабрика"
	IconFruit   CardIcon = "Фрукты"
	IconMine    CardIcon = "Шахта"
	IconMajor   CardIcon = "Крупное предприятие"
	IconFlower  CardIcon = "Цветник"
	IconApple   CardIcon = "Яблоня"
	IconTax     CardIcon = "Налог"
	IconTown    CardIcon = "Мэрия"
	IconHarbor  CardIcon = "Бухта"
)

type CardColor string

const (
	ColorRed    CardColor = "Red"
	ColorBlue   CardColor = "Blue"
	ColorGreen  CardColor = "Green"
	ColorPurple CardColor = "Purple"
)

type LandmarkName string

const (
	LandmarkCityHall         LandmarkName = "ГОРОДСКАЯ РАТУША"
	LandmarkHarbor          LandmarkName = "ПОРТ"
	LandmarkTrainStation    LandmarkName = "ЖД ВОКЗАЛ"
	LandmarkShoppingMall    LandmarkName = "ТОРГОВЫЙ ЦЕНТР"
	LandmarkAmusementPark   LandmarkName = "ПАРК РАЗВЛЕЧЕНИЙ"
	LandmarkRadioTower      LandmarkName = "РАДИОВЕЩАТЕЛЬНАЯ БАШНЯ"
	LandmarkMoonTower       LandmarkName = "ЛУННАЯ БАШНЯ"
	LandmarkAirport         LandmarkName = "АЭРОПОРТ"
)

type CardType string

const (
	TypeEstablishment CardType = "establishment"
	TypeLandmark      CardType = "landmark"
)

type ActivationEffect int8

const (
	EffectNone           ActivationEffect = 0
	EffectPerPlayer      ActivationEffect = 1
	EffectFromBank       ActivationEffect = 2
	EffectFromActive     ActivationEffect = 3
	EffectStealOne       ActivationEffect = 4
	EffectPerRanch       ActivationEffect = 5
	EffectPerForest      ActivationEffect = 6
	EffectPerWheat       ActivationEffect = 7
	EffectPerPurple      ActivationEffect = 8
	EffectHalfOthers     ActivationEffect = 9
)

type Card struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Icon         CardIcon         `json:"icon"`
	Color        CardColor        `json:"color"`
	Numbers      []int            `json:"numbers"`
	Price        uint8            `json:"price"`
	EffectType   ActivationEffect `json:"effect_type"`
	EffectValue  int8             `json:"effect_value"`
	Type         CardType         `json:"type"`
	Condition    string           `json:"condition"`
	MinLandmark  uint8            `json:"min_landmark"`
	DefaultStock uint8            `json:"default_stock"`
}

func (c Card) IsLandmark() bool {
	return c.Type == TypeLandmark
}

func (c Card) ActivatesOn(num int) bool {
	for _, n := range c.Numbers {
		if n == num {
			return true
		}
	}
	return false
}

var (
	establishmentRegistry map[string]Card
	landmarkRegistry      map[string]Card
	registryMu            sync.RWMutex
)

func init() {
	establishmentRegistry = make(map[string]Card)
	landmarkRegistry = make(map[string]Card)
	registerDefaultCards()
}

func registerDefaultCards() {
	RegisterEstablishment(Card{ID: "wheat_field", Name: "Пшеничное поле", Icon: IconWheat, Color: ColorBlue, Numbers: []int{1}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "ranch", Name: "Ранчо", Icon: IconRanch, Color: ColorBlue, Numbers: []int{1}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "bakery", Name: "Пекарня", Icon: IconBakery, Color: ColorGreen, Numbers: []int{2}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "convenience_store", Name: "Магазин", Icon: IconShop, Color: ColorRed, Numbers: []int{2}, Price: 2, EffectType: EffectFromActive, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "forest", Name: "Лес", Icon: IconForest, Color: ColorBlue, Numbers: []int{3}, Price: 3, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "cafe", Name: "Кафе", Icon: IconCafe, Color: ColorRed, Numbers: []int{3}, Price: 2, EffectType: EffectFromActive, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "stadium", Name: "Стадион", Icon: IconMajor, Color: ColorPurple, Numbers: []int{4}, Price: 6, EffectType: EffectPerPlayer, EffectValue: 2, DefaultStock: 4})
	RegisterEstablishment(Card{ID: "tv_station", Name: "Телестанция", Icon: IconMajor, Color: ColorPurple, Numbers: []int{4}, Price: 7, EffectType: EffectStealOne, EffectValue: 5, DefaultStock: 4})
	RegisterEstablishment(Card{ID: "cheese_factory", Name: "Сыроварня", Icon: IconFactory, Color: ColorGreen, Numbers: []int{5}, Price: 5, EffectType: EffectPerRanch, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "furniture_factory", Name: "Мебельная фабрика", Icon: IconFactory, Color: ColorGreen, Numbers: []int{5}, Price: 3, EffectType: EffectPerForest, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "fruit_market", Name: "Фруктовый рынок", Icon: IconFruit, Color: ColorGreen, Numbers: []int{6}, Price: 3, EffectType: EffectPerWheat, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "sushi_bar", Name: "Суши-бар", Icon: IconRest, Color: ColorBlue, Numbers: []int{6}, Price: 4, EffectType: EffectFromBank, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "mine", Name: "Шахта", Icon: IconMine, Color: ColorBlue, Numbers: []int{7}, Price: 6, EffectType: EffectFromBank, EffectValue: 5, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "flower_garden", Name: "Цветник", Icon: IconFlower, Color: ColorGreen, Numbers: []int{8}, Price: 4, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "apple_orchard", Name: "Яблоневый сад", Icon: IconApple, Color: ColorBlue, Numbers: []int{9}, Price: 3, EffectType: EffectFromBank, EffectValue: 3, DefaultStock: 6})
	RegisterEstablishment(Card{ID: "nightclub", Name: "Ночной клуб", Icon: IconMajor, Color: ColorPurple, Numbers: []int{10}, Price: 8, EffectType: EffectFromActive, EffectValue: 5, DefaultStock: 4})
	RegisterEstablishment(Card{ID: "tax_office", Name: "Налоговая", Icon: IconTax, Color: ColorBlue, Numbers: []int{11}, Price: 6, EffectType: EffectHalfOthers, EffectValue: 0, DefaultStock: 4})
	RegisterEstablishment(Card{ID: "town_hall", Name: "Мэрия", Icon: IconTown, Color: ColorPurple, Numbers: []int{12}, Price: 6, EffectType: EffectPerPlayer, EffectValue: 1, DefaultStock: 4})
	RegisterEstablishment(Card{ID: "harbor_est", Name: "Бухта (предприятие)", Icon: IconHarbor, Color: ColorBlue, Numbers: []int{12}, Price: 7, EffectType: EffectPerPurple, EffectValue: 2, DefaultStock: 6})

	RegisterLandmark(Card{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "train_station", Name: string(LandmarkTrainStation), Price: 4, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "shopping_mall", Name: string(LandmarkShoppingMall), Price: 10, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "amusement_park", Name: string(LandmarkAmusementPark), Price: 16, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "radio_tower", Name: string(LandmarkRadioTower), Price: 22, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "moon_tower", Name: string(LandmarkMoonTower), Price: 22, Type: TypeLandmark})
	RegisterLandmark(Card{ID: "airport", Name: string(LandmarkAirport), Price: 30, Type: TypeLandmark})
}

func DefaultEstablishments() []Card {
	return []Card{
		{ID: "wheat_field", Name: "Пшеничное поле", Icon: IconWheat, Color: ColorBlue, Numbers: []int{1}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6},
		{ID: "ranch", Name: "Ранчо", Icon: IconRanch, Color: ColorBlue, Numbers: []int{1}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6},
		{ID: "bakery", Name: "Пекарня", Icon: IconBakery, Color: ColorGreen, Numbers: []int{2}, Price: 1, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6},
		{ID: "convenience_store", Name: "Магазин", Icon: IconShop, Color: ColorRed, Numbers: []int{2}, Price: 2, EffectType: EffectFromActive, EffectValue: 3, DefaultStock: 6},
		{ID: "forest", Name: "Лес", Icon: IconForest, Color: ColorBlue, Numbers: []int{3}, Price: 3, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6},
		{ID: "cafe", Name: "Кафе", Icon: IconCafe, Color: ColorRed, Numbers: []int{3}, Price: 2, EffectType: EffectFromActive, EffectValue: 1, DefaultStock: 6},
		{ID: "stadium", Name: "Стадион", Icon: IconMajor, Color: ColorPurple, Numbers: []int{4}, Price: 6, EffectType: EffectPerPlayer, EffectValue: 2, DefaultStock: 4},
		{ID: "tv_station", Name: "Телестанция", Icon: IconMajor, Color: ColorPurple, Numbers: []int{4}, Price: 7, EffectType: EffectStealOne, EffectValue: 5, DefaultStock: 4},
		{ID: "cheese_factory", Name: "Сыроварня", Icon: IconFactory, Color: ColorGreen, Numbers: []int{5}, Price: 5, EffectType: EffectPerRanch, EffectValue: 3, DefaultStock: 6},
		{ID: "furniture_factory", Name: "Мебельная фабрика", Icon: IconFactory, Color: ColorGreen, Numbers: []int{5}, Price: 3, EffectType: EffectPerForest, EffectValue: 3, DefaultStock: 6},
		{ID: "fruit_market", Name: "Фруктовый рынок", Icon: IconFruit, Color: ColorGreen, Numbers: []int{6}, Price: 3, EffectType: EffectPerWheat, EffectValue: 3, DefaultStock: 6},
		{ID: "sushi_bar", Name: "Суши-бар", Icon: IconRest, Color: ColorBlue, Numbers: []int{6}, Price: 4, EffectType: EffectFromBank, EffectValue: 3, DefaultStock: 6},
		{ID: "mine", Name: "Шахта", Icon: IconMine, Color: ColorBlue, Numbers: []int{7}, Price: 6, EffectType: EffectFromBank, EffectValue: 5, DefaultStock: 6},
		{ID: "flower_garden", Name: "Цветник", Icon: IconFlower, Color: ColorGreen, Numbers: []int{8}, Price: 4, EffectType: EffectFromBank, EffectValue: 1, DefaultStock: 6},
		{ID: "apple_orchard", Name: "Яблоневый сад", Icon: IconApple, Color: ColorBlue, Numbers: []int{9}, Price: 3, EffectType: EffectFromBank, EffectValue: 3, DefaultStock: 6},
		{ID: "nightclub", Name: "Ночной клуб", Icon: IconMajor, Color: ColorPurple, Numbers: []int{10}, Price: 8, EffectType: EffectFromActive, EffectValue: 5, DefaultStock: 4},
		{ID: "tax_office", Name: "Налоговая", Icon: IconTax, Color: ColorBlue, Numbers: []int{11}, Price: 6, EffectType: EffectHalfOthers, EffectValue: 0, DefaultStock: 4},
		{ID: "town_hall", Name: "Мэрия", Icon: IconTown, Color: ColorPurple, Numbers: []int{12}, Price: 6, EffectType: EffectPerPlayer, EffectValue: 1, DefaultStock: 4},
		{ID: "harbor_est", Name: "Бухта (предприятие)", Icon: IconHarbor, Color: ColorBlue, Numbers: []int{12}, Price: 7, EffectType: EffectPerPurple, EffectValue: 2, DefaultStock: 6},
	}
}

func DefaultLandmarks() []Card {
	return []Card{
		{ID: "harbor_lm", Name: string(LandmarkHarbor), Price: 2, Type: TypeLandmark},
		{ID: "train_station", Name: string(LandmarkTrainStation), Price: 4, Type: TypeLandmark},
		{ID: "shopping_mall", Name: string(LandmarkShoppingMall), Price: 10, Type: TypeLandmark},
		{ID: "amusement_park", Name: string(LandmarkAmusementPark), Price: 16, Type: TypeLandmark},
		{ID: "radio_tower", Name: string(LandmarkRadioTower), Price: 22, Type: TypeLandmark},
		{ID: "moon_tower", Name: string(LandmarkMoonTower), Price: 22, Type: TypeLandmark},
		{ID: "airport", Name: string(LandmarkAirport), Price: 30, Type: TypeLandmark},
	}
}

type NewCardParams struct {
	Icon         CardIcon
	Color        CardColor
	Numbers      []int
	Price        uint8
	EffectType   ActivationEffect
	EffectValue  int8
	Type         CardType
	Condition    string
	MinLandmark  uint8
	DefaultStock uint8
}

func NewCard(id, name string, params NewCardParams) (Card, error) {
	if id == "" {
		return Card{}, errors.New("card ID cannot be empty")
	}
	if name == "" {
		return Card{}, errors.New("card name cannot be empty")
	}
	if params.Type == TypeEstablishment && len(params.Numbers) == 0 {
		return Card{}, errors.New("establishment must have at least one activation number")
	}

	return Card{
		ID:           id,
		Name:         name,
		Icon:         params.Icon,
		Color:        params.Color,
		Numbers:      params.Numbers,
		Price:        params.Price,
		EffectType:   params.EffectType,
		EffectValue:  params.EffectValue,
		Type:         params.Type,
		Condition:    params.Condition,
		MinLandmark:  params.MinLandmark,
		DefaultStock: params.DefaultStock,
	}, nil
}

func RegisterEstablishment(card Card) error {
	if card.ID == "" {
		return errors.New("establishment ID cannot be empty")
	}
	card.Type = TypeEstablishment
	if card.DefaultStock == 0 {
		card.DefaultStock = 6
	}
	registryMu.Lock()
	establishmentRegistry[card.ID] = card
	registryMu.Unlock()
	return nil
}

func RegisterLandmark(card Card) error {
	if card.ID == "" {
		return errors.New("landmark ID cannot be empty")
	}
	card.Type = TypeLandmark
	registryMu.Lock()
	landmarkRegistry[card.ID] = card
	registryMu.Unlock()
	return nil
}

func RemoveEstablishment(id string) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := establishmentRegistry[id]; !exists {
		return errors.New("establishment not found: " + id)
	}
	delete(establishmentRegistry, id)
	return nil
}

func RemoveLandmark(id string) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := landmarkRegistry[id]; !exists {
		return errors.New("landmark not found: " + id)
	}
	delete(landmarkRegistry, id)
	return nil
}

func GetEstablishment(id string) (Card, bool) {
	registryMu.RLock()
	c, ok := establishmentRegistry[id]
	registryMu.RUnlock()
	return c, ok
}

func GetLandmark(id string) (Card, bool) {
	registryMu.RLock()
	c, ok := landmarkRegistry[id]
	registryMu.RUnlock()
	return c, ok
}

func AllEstablishments() []Card {
	registryMu.RLock()
	result := make([]Card, 0, len(establishmentRegistry))
	for _, c := range establishmentRegistry {
		result = append(result, c)
	}
	registryMu.RUnlock()
	return result
}

func AllLandmarks() []Card {
	registryMu.RLock()
	result := make([]Card, 0, len(landmarkRegistry))
	for _, c := range landmarkRegistry {
		result = append(result, c)
	}
	registryMu.RUnlock()
	return result
}

func RemoveAllCards() {
	registryMu.Lock()
	clear(establishmentRegistry)
	clear(landmarkRegistry)
	registryMu.Unlock()
}
