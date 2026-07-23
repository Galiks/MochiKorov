package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mochi_korov/internal/game"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	s := &PostgresStore{pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStore) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			game_data JSONB,
			completed BOOLEAN NOT NULL DEFAULT FALSE
		)`,
		`ALTER TABLE sessions ADD COLUMN IF NOT EXISTS completed BOOLEAN NOT NULL DEFAULT FALSE`,
		`CREATE TABLE IF NOT EXISTS card_sets (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			cards JSONB NOT NULL DEFAULT '[]'::jsonb
		)`,
	}

	for _, q := range queries {
		if _, err := s.pool.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) SeedDefaults() error {
	ctx := context.Background()

	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM card_sets`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	_, err = s.pool.Exec(ctx,
		`INSERT INTO card_sets (id, name, created_at, updated_at, cards) VALUES ($1, $2, $3, $4, $5)`,
		"base", "Базовый набор", now, now, "[]",
	)
	if err != nil {
		return fmt.Errorf("seed card_set: %w", err)
	}

	cards := game.DefaultEstablishments()
	cards = append(cards, game.DefaultLandmarks()...)
	return s.SaveCardSetCards("base", cards)
}

func scanSession(row pgx.Row) (*Session, error) {
	var s Session
	var gameData []byte
	err := row.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt, &gameData, &s.Completed)
	if err != nil {
		return nil, err
	}
	if gameData != nil {
		var sd game.SaveData
		if err := json.Unmarshal(gameData, &sd); err == nil {
			s.GameData = &sd
		}
	}
	return &s, nil
}

func (s *PostgresStore) ListSessions() ([]Session, error) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, created_at, updated_at, game_data, completed FROM sessions ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *sess)
	}
	return result, nil
}

func (s *PostgresStore) CreateSession(id, name string) (*Session, error) {
	ctx := context.Background()
	now := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		id, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	row := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at, game_data, completed FROM sessions WHERE id = $1`, id,
	)
	return scanSession(row)
}

func (s *PostgresStore) GetSession(id string) (*Session, error) {
	ctx := context.Background()
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at, game_data, completed FROM sessions WHERE id = $1`, id,
	)
	sess, err := scanSession(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, err
	}
	return sess, nil
}

func (s *PostgresStore) DeleteSession(id string) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) SetSessionCompleted(id string) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET completed = TRUE, updated_at = NOW() WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) SaveGameData(id string, data *game.SaveData) error {
	ctx := context.Background()
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET game_data = $1, updated_at = NOW() WHERE id = $2`,
		raw, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) LoadGameData(id string) (*game.SaveData, error) {
	ctx := context.Background()
	var raw []byte
	err := s.pool.QueryRow(ctx,
		`SELECT game_data FROM sessions WHERE id = $1`, id,
	).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	var sd game.SaveData
	if err := json.Unmarshal(raw, &sd); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &sd, nil
}

func scanCardSet(row pgx.Row) (*CardSet, error) {
	var cs CardSet
	var cardsJSON []byte
	err := row.Scan(&cs.ID, &cs.Name, &cs.CreatedAt, &cs.UpdatedAt, &cardsJSON)
	if err != nil {
		return nil, err
	}
	if cardsJSON != nil {
		var cards []game.Card
		if err := json.Unmarshal(cardsJSON, &cards); err == nil {
			cs.Cards = cards
		}
	}
	return &cs, nil
}

func (s *PostgresStore) ListCardSets() ([]CardSet, error) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, created_at, updated_at, cards FROM card_sets ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CardSet
	for rows.Next() {
		cs, err := scanCardSet(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *cs)
	}
	return result, nil
}

func (s *PostgresStore) CreateCardSet(id, name string) (*CardSet, error) {
	ctx := context.Background()
	now := time.Now()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO card_sets (id, name, created_at, updated_at, cards) VALUES ($1, $2, $3, $4, '[]'::jsonb)`,
		id, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create card_set: %w", err)
	}

	row := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at, cards FROM card_sets WHERE id = $1`, id,
	)
	return scanCardSet(row)
}

func (s *PostgresStore) GetCardSet(id string) (*CardSet, error) {
	ctx := context.Background()
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at, cards FROM card_sets WHERE id = $1`, id,
	)
	cs, err := scanCardSet(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("card_set not found: %s", id)
		}
		return nil, err
	}
	return cs, nil
}

func (s *PostgresStore) DeleteCardSet(id string) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx, `DELETE FROM card_sets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("card_set not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) SaveCardSetCards(id string, cards []game.Card) error {
	ctx := context.Background()
	raw, err := json.Marshal(cards)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE card_sets SET cards = $1, updated_at = NOW() WHERE id = $2`,
		raw, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("card_set not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) LoadCardSetCards(id string) ([]game.Card, error) {
	ctx := context.Background()
	var raw []byte
	err := s.pool.QueryRow(ctx,
		`SELECT cards FROM card_sets WHERE id = $1`, id,
	).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("card_set not found: %s", id)
		}
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	var cards []game.Card
	if err := json.Unmarshal(raw, &cards); err != nil {
		return nil, fmt.Errorf("unmarshal cards: %w", err)
	}
	return cards, nil
}
