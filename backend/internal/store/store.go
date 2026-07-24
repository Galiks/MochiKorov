package store

import (
	"time"

	"mochi_korov/internal/game"
)

type Session struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Completed bool            `json:"completed"`
	GameData  *game.SaveData  `json:"game_data,omitempty"`
}

type CardSet struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Cards     []game.Card `json:"cards,omitempty"`
}

type Store interface {
	Close() error

	ListSessions() ([]Session, error)
	CreateSession(id, name string) (*Session, error)
	GetSession(id string) (*Session, error)
	DeleteSession(id string) error
	SaveGameData(id string, data *game.SaveData) error
	LoadGameData(id string) (*game.SaveData, error)
	SetSessionCompleted(id string) error

	ListCardSets() ([]CardSet, error)
	CreateCardSet(id, name string) (*CardSet, error)
	GetCardSet(id string) (*CardSet, error)
	DeleteCardSet(id string) error
	SaveCardSetCards(id string, cards []game.Card) error
	LoadCardSetCards(id string) ([]game.Card, error)

	SeedDefaults() error
}
