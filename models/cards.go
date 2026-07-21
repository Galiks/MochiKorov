package models

import "fmt"

type Card struct {
	ID          string
	Name        string
	Description string
	Icon        CardIcon
	Type        CardTypeEng
	Conditions  string // набор
	CubeNumber  CubeNumber
	Price       uint8 // цена карточки
	MoneyOnCard uint8 // Венчурный фонд может иметь деньги на карточке, которые можно забрать у всех игроков, если достопримечательность активирована.
	Money       int8  // карточка даёт или отнимает деньги.
	IsEnable    bool
}

type HandCard []Card

func (h HandCard) IsHasCardByName(cardName string) bool {
	for _, card := range h {
		if card.Name == cardName {
			return true
		}
	}
	return false
}

func (h HandCard) IsHasCardByNameAndType(cardName string, cardType CardTypeEng) bool {
	for _, card := range h {
		if card.Name == cardName && card.Type == cardType {
			return true
		}
	}
	return false
}

type CardIcon string

const (
	ShopIcon               CardIcon = "Shop"
	BoatIcon               CardIcon = "Boat"
	RanchIcon              CardIcon = "Ranch"
	CafeIcon               CardIcon = "Cafe"
	FactoryIcon            CardIcon = "Factory"
	FruitIcon              CardIcon = "Fruit"
	IndustryIcon           CardIcon = "Industry"
	MajorEstablishmentIcon CardIcon = "Major Establishment"
	CaseIcon               CardIcon = "Case"
	WheatSpikeIcon         CardIcon = "Wheat Spike"
)

type CardTypeEng string

const (
	RedEng         CardTypeEng = "Red"
	BlueEng        CardTypeEng = "Blue"
	PurpleEng      CardTypeEng = "Purple"
	GreenEng       CardTypeEng = "Green"
	AttractionsEng CardTypeEng = "Attractions"
)

type CardTypeRus string

const (
	RedRus         CardTypeRus = "Красная"
	BlueRus        CardTypeRus = "Синяя"
	PurpleRus      CardTypeRus = "Фиолетовая"
	GreenRus       CardTypeRus = "Зеленая"
	AttractionsRus CardTypeRus = "Достопримечательности" // TODO: У них есть свои особенность, по типу переброс кубика.
)

type TranslateCard map[CardTypeEng]CardTypeRus

type AttractionName string

const (
	CityHallAttractionName      AttractionName = "ГОРОДСКАЯ РАТУША"       // 0 монета. Есть изначально
	HarborAttractionName        AttractionName = "ПОРТ"                   // 2 монеты
	TrainStationAttractionName  AttractionName = "ЖД ВОКЗАЛ"              // 4 монеты
	ShoppingMallAttractionName  AttractionName = "ТОРГОВЫЙ ЦЕНТР"         // 10 монет
	AmusementParkAttractionName AttractionName = "ПАРК РАЗВЛЕЧЕНИЙ"       // 16 монет
	RadioTowerAttractionName    AttractionName = "РАДИОВЕЩАТЕЛЬНАЯ БАШНЯ" // 22 монеты
	MoonTowerAttractionName     AttractionName = "ЛУННАЯ БАШНЯ"           // 22 монет
	AirportAttractionName       AttractionName = "АЭРОПОРТ"               // 30 монет
)

type CubeNumber uint8

func NewCubeNumber(n CubeNumber) (CubeNumber, error) {
	if n == 0 || n > 14 {
		return 0, fmt.Errorf("Invalid cube number. Must be between 1 and 14, got %d", n)
	}
	return CubeNumber(n), nil
}

type Condition string // условия работы карточки

const (
	NeedAttraction             Condition = "NeedAttraction"      // Нужна достопримечательность. Например, ПОРТ
	EqualOrMoreAttractionCount Condition = "AttractionCount"     // Нужно определённое кол-во достопримечательностей от какого-то числа. Например, от 3 достопримечательностей для частного бара.
	LessAttractionCount        Condition = "LessAttractionCount" // Нужно определённое кол-во достопримечательностей меньше какого-то числа. Например, меньше 2 достопримечательностей для кукурузного поля.
	AlreadyHasCardType         Condition = "AlreadyHasCardType"  // Уже имеет определённый тип карты
)

type CardCondition struct {
	Name   Condition
	Number uint8 // число достопримечательностей, которое нужно иметь для активации карты
}

func (c *CardCondition) checkConditionName(name string) bool {
	switch name {
	case string(NeedAttraction):
		return true
	case string(EqualOrMoreAttractionCount):
		return true
	case string(LessAttractionCount):
		return true
	case string(AlreadyHasCardType):
		return true
	default:
		return false
	}
}

// type Condition struct {
// 	Name           string
// 	NeedAttraction bool
// }

// type Condition_NeedAttraction struct {
// 	AttractionName string
// }

func checkAttractionName(name AttractionName) bool {
	switch name {
	case "ПОРТ":
		return true
	default:
		return false
	}
}

// type CardInterface interface {
// 	GetID() string
// 	GetName() string
// 	GetDescription() string
// 	GetType() CardTypeEng
// 	GetCubeNumber() CubeNumber
// 	GetPrice() uint
// 	GetConditions() []Condition
// }
