package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"mochi_korov/internal/game"
)

type FileStore struct {
	mu       sync.RWMutex
	dataDir  string
	sessions map[string]*Session
	cardSets map[string]*CardSet
}

func NewFileStore(dataDir string) *FileStore {
	s := &FileStore{
		dataDir:  dataDir,
		sessions: make(map[string]*Session),
		cardSets: make(map[string]*CardSet),
	}
	s.loadAll()
	return s
}

func (s *FileStore) Close() error {
	return nil
}

func (s *FileStore) sessionPath(id string) string {
	return filepath.Join(s.dataDir, "sessions", id+".json")
}

func (s *FileStore) cardSetPath(id string) string {
	return filepath.Join(s.dataDir, "card_sets", id+".json")
}

func (s *FileStore) ensureDirs() error {
	for _, dir := range []string{"sessions", "card_sets"} {
		if err := os.MkdirAll(filepath.Join(s.dataDir, dir), 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStore) loadAll() {
	s.ensureDirs()

	sessionsDir := filepath.Join(s.dataDir, "sessions")
	if entries, err := os.ReadDir(sessionsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
				continue
			}
			id := e.Name()[:len(e.Name())-5]
			if data, err := os.ReadFile(filepath.Join(sessionsDir, e.Name())); err == nil {
				var sess Session
				if json.Unmarshal(data, &sess) == nil {
					s.sessions[id] = &sess
				}
			}
		}
	}

	setsDir := filepath.Join(s.dataDir, "card_sets")
	if entries, err := os.ReadDir(setsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
				continue
			}
			id := e.Name()[:len(e.Name())-5]
			if data, err := os.ReadFile(filepath.Join(setsDir, e.Name())); err == nil {
				var cs CardSet
				if json.Unmarshal(data, &cs) == nil {
					s.cardSets[id] = &cs
				}
			}
		}
	}
}

func (s *FileStore) saveSession(sess *Session) error {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.sessionPath(sess.ID), data, 0644)
}

func (s *FileStore) saveCardSet(cs *CardSet) error {
	data, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.cardSetPath(cs.ID), data, 0644)
}

func (s *FileStore) ListSessions() ([]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		result = append(result, *sess)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result, nil
}

func (s *FileStore) CreateSession(id, name string, maxPlayers int, creatorToken string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; exists {
		return nil, fmt.Errorf("session already exists: %s", id)
	}

	now := time.Now()
	sess := &Session{
		ID:         id,
		Name:       name,
		CreatedAt:  now,
		UpdatedAt:  now,
		MaxPlayers: maxPlayers,
		LobbyPlayers: []LobbyPlayer{
			{Name: name, Token: creatorToken},
		},
	}

	if err := s.ensureDirs(); err != nil {
		return nil, err
	}
	if err := s.saveSession(sess); err != nil {
		return nil, err
	}
	s.sessions[id] = sess
	return sess, nil
}

func (s *FileStore) GetSession(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess, nil
}

func (s *FileStore) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(s.sessions, id)
	os.Remove(s.sessionPath(id))
	return nil
}

func (s *FileStore) SaveGameData(id string, data *game.SaveData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	sess.GameData = data
	sess.UpdatedAt = time.Now()
	return s.saveSession(sess)
}

func (s *FileStore) LoadGameData(id string) (*game.SaveData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sess.GameData, nil
}

func (s *FileStore) SetSessionCompleted(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	sess.Completed = true
	sess.UpdatedAt = time.Now()
	return s.saveSession(sess)
}

func (s *FileStore) SaveLobbyPlayers(id string, players []LobbyPlayer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, exists := s.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	sess.LobbyPlayers = players
	sess.UpdatedAt = time.Now()
	return s.saveSession(sess)
}

func (s *FileStore) ListCardSets() ([]CardSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]CardSet, 0, len(s.cardSets))
	for _, cs := range s.cardSets {
		result = append(result, *cs)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result, nil
}

func (s *FileStore) CreateCardSet(id, name string) (*CardSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cardSets[id]; exists {
		return nil, fmt.Errorf("card_set already exists: %s", id)
	}

	now := time.Now()
	cs := &CardSet{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.ensureDirs(); err != nil {
		return nil, err
	}
	if err := s.saveCardSet(cs); err != nil {
		return nil, err
	}
	s.cardSets[id] = cs
	return cs, nil
}

func (s *FileStore) GetCardSet(id string) (*CardSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, exists := s.cardSets[id]
	if !exists {
		return nil, fmt.Errorf("card_set not found: %s", id)
	}
	return cs, nil
}

func (s *FileStore) DeleteCardSet(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cardSets[id]; !exists {
		return fmt.Errorf("card_set not found: %s", id)
	}

	delete(s.cardSets, id)
	os.Remove(s.cardSetPath(id))
	return nil
}

func (s *FileStore) SaveCardSetCards(id string, cards []game.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cs, exists := s.cardSets[id]
	if !exists {
		return fmt.Errorf("card_set not found: %s", id)
	}

	cs.Cards = cards
	cs.UpdatedAt = time.Now()
	return s.saveCardSet(cs)
}

func (s *FileStore) LoadCardSetCards(id string) ([]game.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cs, exists := s.cardSets[id]
	if !exists {
		return nil, fmt.Errorf("card_set not found: %s", id)
	}
	return cs.Cards, nil
}

func (s *FileStore) SeedDefaults() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cardSets["base"]; exists {
		return nil
	}

	if err := s.ensureDirs(); err != nil {
		return err
	}

	now := time.Now()
	cards := game.DefaultEstablishments()
	cards = append(cards, game.DefaultLandmarks()...)
	cs := &CardSet{
		ID:        "base",
		Name:      "Базовый набор",
		CreatedAt: now,
		UpdatedAt: now,
		Cards:     cards,
	}
	if err := s.saveCardSet(cs); err != nil {
		return err
	}
	s.cardSets["base"] = cs
	return nil
}
