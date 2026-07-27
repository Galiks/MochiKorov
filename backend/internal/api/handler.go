package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"mochi_korov/internal/game"
	"mochi_korov/internal/store"
)

type handler struct {
	store store.Store
}

type cardRequest struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Icon         game.CardIcon           `json:"icon"`
	Color        game.CardColor          `json:"color"`
	Numbers      []int                   `json:"numbers"`
	Price        uint8                   `json:"price"`
	EffectType   game.ActivationEffect   `json:"effect_type"`
	EffectValue  int8                    `json:"effect_value"`
	Type         game.CardType           `json:"type"`
	Condition    string                  `json:"condition"`
	MinLandmark  uint8                   `json:"min_landmark"`
	DefaultStock uint8                   `json:"default_stock"`
}

type cardResponse struct {
	Card    game.Card `json:"card,omitempty"`
	Message string    `json:"message,omitempty"`
	Error   string    `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func requestToParams(r cardRequest) game.NewCardParams {
	return game.NewCardParams{
		Icon:         r.Icon,
		Color:        r.Color,
		Numbers:      r.Numbers,
		Price:        r.Price,
		EffectType:   r.EffectType,
		EffectValue:  r.EffectValue,
		Type:         r.Type,
		Condition:    r.Condition,
		MinLandmark:  r.MinLandmark,
		DefaultStock: r.DefaultStock,
	}
}

func (h *handler) handleListEstablishments(w http.ResponseWriter, r *http.Request) {
	cards := game.AllEstablishments()
	writeJSON(w, http.StatusOK, map[string]interface{}{"cards": cards})
}

func (h *handler) handleCreateEstablishment(w http.ResponseWriter, r *http.Request) {
	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	card, err := game.NewCard(req.ID, req.Name, requestToParams(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := game.RegisterEstablishment(card); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cardResponse{Card: card, Message: "created"})
}

func (h *handler) handleGetEstablishment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	card, ok := game.GetEstablishment(id)
	if !ok {
		writeError(w, http.StatusNotFound, "establishment not found: "+id)
		return
	}
	writeJSON(w, http.StatusOK, cardResponse{Card: card})
}

func (h *handler) handleUpdateEstablishment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := game.GetEstablishment(id); !ok {
		writeError(w, http.StatusNotFound, "establishment not found: "+id)
		return
	}

	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	card, err := game.NewCard(id, req.Name, requestToParams(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	game.RemoveEstablishment(id)
	game.RegisterEstablishment(card)

	writeJSON(w, http.StatusOK, cardResponse{Card: card, Message: "updated"})
}

func (h *handler) handleDeleteEstablishment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := game.RemoveEstablishment(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cardResponse{Message: "deleted"})
}

func (h *handler) handleListLandmarks(w http.ResponseWriter, r *http.Request) {
	cards := game.AllLandmarks()
	writeJSON(w, http.StatusOK, map[string]interface{}{"cards": cards})
}

func (h *handler) handleCreateLandmark(w http.ResponseWriter, r *http.Request) {
	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	req.Type = game.TypeLandmark

	card, err := game.NewCard(req.ID, req.Name, requestToParams(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := game.RegisterLandmark(card); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cardResponse{Card: card, Message: "created"})
}

func (h *handler) handleGetLandmark(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	card, ok := game.GetLandmark(id)
	if !ok {
		writeError(w, http.StatusNotFound, "landmark not found: "+id)
		return
	}
	writeJSON(w, http.StatusOK, cardResponse{Card: card})
}

func (h *handler) handleUpdateLandmark(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := game.GetLandmark(id); !ok {
		writeError(w, http.StatusNotFound, "landmark not found: "+id)
		return
	}

	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	req.Type = game.TypeLandmark

	card, err := game.NewCard(id, req.Name, requestToParams(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	game.RemoveLandmark(id)
	game.RegisterLandmark(card)

	writeJSON(w, http.StatusOK, cardResponse{Card: card, Message: "updated"})
}

func (h *handler) handleDeleteLandmark(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := game.RemoveLandmark(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, cardResponse{Message: "deleted"})
}

func (h *handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.store.ListSessions()
	if err != nil {
		log.Printf("list sessions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []store.Session{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

func (h *handler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		MaxPlayers int    `json:"max_players"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	if req.Name == "" {
		req.Name = req.ID
	}
	if req.MaxPlayers < 2 || req.MaxPlayers > 5 {
		req.MaxPlayers = 2
	}

	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	creatorToken := hex.EncodeToString(tokenBytes)

	sess, err := h.store.CreateSession(req.ID, req.Name, req.MaxPlayers, creatorToken)
	if err != nil {
		log.Printf("create session: %v", err)
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"session": sess,
		"token":   creatorToken,
	})
}

func (h *handler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := h.store.GetSession(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (h *handler) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.DeleteSession(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *handler) handleSaveSessionGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var sd game.SaveData
	if err := json.NewDecoder(r.Body).Decode(&sd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := h.store.SaveGameData(id, &sd); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func (h *handler) handleLoadSessionGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, err := h.store.LoadGameData(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if data == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"game_data": nil})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"game_data": data})
}

func (h *handler) handleListCardSets(w http.ResponseWriter, r *http.Request) {
	sets, err := h.store.ListCardSets()
	if err != nil {
		log.Printf("list card sets: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list card sets")
		return
	}
	if sets == nil {
		sets = []store.CardSet{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"card_sets": sets})
}

func (h *handler) handleCreateCardSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	if req.Name == "" {
		req.Name = req.ID
	}

	cs, err := h.store.CreateCardSet(req.ID, req.Name)
	if err != nil {
		log.Printf("create card set: %v", err)
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, cs)
}

func (h *handler) handleGetCardSet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cs, err := h.store.GetCardSet(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func (h *handler) handleDeleteCardSet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.DeleteCardSet(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *handler) handleSaveCardSetCards(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Cards []game.Card `json:"cards"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := h.store.SaveCardSetCards(id, req.Cards); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "saved"})
}

func (h *handler) handleLoadCardSetCards(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cards, err := h.store.LoadCardSetCards(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if cards == nil {
		cards = []game.Card{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"cards": cards})
}

func (h *handler) handleSeedCardSet(w http.ResponseWriter, r *http.Request) {
	// Check if "base" set already exists
	existing, err := h.store.ListCardSets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, cs := range existing {
		if cs.ID == "base" {
			writeJSON(w, http.StatusOK, map[string]string{"message": "already seeded"})
			return
		}
	}

	if err := h.store.SeedDefaults(); err != nil {
		log.Printf("seed defaults: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to seed defaults")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "seeded"})
}
