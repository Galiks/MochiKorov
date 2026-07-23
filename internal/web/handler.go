package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"mochi_korov/internal/game"
	"mochi_korov/internal/metrics"
	"mochi_korov/internal/store"
)

type handler struct {
	store    store.Store
	lokiURL  string
}

func lokiURL() string {
	u := os.Getenv("LOKI_URL")
	if u == "" {
		u = "http://localhost:3100"
	}
	return u
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

type lokiPayload struct {
	Streams []lokiStream `json:"streams"`
}

func (h *handler) pushLog(sessionID string, turn, playerID int, actionType string, details map[string]interface{}) {
	go func() {
		line, _ := json.Marshal(details)
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)

		payload := lokiPayload{
			Streams: []lokiStream{{
				Stream: map[string]string{
					"app":     "mochikorov",
					"session": sessionID,
					"action":  actionType,
					"player":  strconv.Itoa(playerID),
					"turn":    strconv.Itoa(turn),
				},
				Values: [][2]string{{ts, string(line)}},
			}},
		}

		body, _ := json.Marshal(payload)
		http.Post(h.lokiURL+"/loki/api/v1/push", "application/json", bytes.NewReader(body))
	}()
}

func (h *handler) recordPlayerMoney(sessionID string, players []*game.Player) {
	for _, p := range players {
		metrics.PlayerMoney.WithLabelValues(
			sessionID, strconv.Itoa(p.ID), p.Name,
		).Set(float64(p.Money))
	}
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiError{Error: msg})
}

func (h *handler) loadGame(id string) (*game.Game, error) {
	data, err := h.store.LoadGameData(id)
	if err != nil {
		return nil, fmt.Errorf("load game: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("no game data in session %s", id)
	}
	return game.GameFromSaveData(data), nil
}

func (h *handler) saveGame(id string, g *game.Game) error {
	return h.store.SaveGameData(id, g.ToSaveData())
}

func (h *handler) stateResponse(id string, g *game.Game, log []string) *game.GameStateResponse {
	if log == nil {
		log = []string{}
	}
	g.CheckWin()
	h.recordPlayerMoney(id, g.Players)
	return g.ToStateResponse(id, log)
}

func (h *handler) loadCardSet(setName string) {
	cards, err := h.store.LoadCardSetCards(setName)
	if err != nil {
		log.Printf("card set %q not found: %v", setName, err)
		return
	}
	game.RemoveAllCards()
	for _, c := range cards {
		if c.IsLandmark() {
			game.RegisterLandmark(c)
		} else {
			game.RegisterEstablishment(c)
		}
	}
}

func (h *handler) handleStart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g := game.NewGame([]string{"Игрок", "Бот 1", "Бот 2", "Бот 3"})

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.GamesStarted.Inc()
	metrics.ActiveGames.Inc()
	h.pushLog(id, 0, 0, "game_start", map[string]interface{}{
		"players": []string{"Игрок", "Бот 1", "Бот 2", "Бот 3"},
	})

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, []string{"Новая игра создана"}))
}

func (h *handler) handleState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, nil))
}

func (h *handler) handleRoll(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if g.CurrentPlayer != 0 {
		writeError(w, http.StatusBadRequest, "not your turn")
		return
	}
	if g.Phase != game.PhaseRoll {
		writeError(w, http.StatusBadRequest, "cannot roll now")
		return
	}

	var rollReq struct {
		DiceCount int `json:"dice_count"`
	}
	json.NewDecoder(r.Body).Decode(&rollReq)

	result, err := g.RollWith(rollReq.DiceCount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log := []string{fmt.Sprintf("%s бросает кубики: %s", g.Current().Name, game.FormatDice(result))}

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.RollsTotal.Inc()
	h.pushLog(id, g.Turn, g.CurrentPlayer, "roll", map[string]interface{}{
		"dice": result.Numbers, "sum": result.Sum, "player_name": g.Current().Name,
	})

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, log))
}

func (h *handler) handleCollect(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if g.CurrentPlayer != 0 {
		writeError(w, http.StatusBadRequest, "not your turn")
		return
	}
	if g.Phase != game.PhaseIncome {
		writeError(w, http.StatusBadRequest, "cannot collect income now")
		return
	}

	actLog := g.ActivateCards()
	log := make([]string, 0)
	log = append(log, actLog...)

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.TurnsTotal.Inc()
	for _, entry := range actLog {
		h.pushLog(id, g.Turn, g.CurrentPlayer, "income", map[string]interface{}{
			"message": entry, "player_name": g.Current().Name,
		})
	}

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, log))
}

func (h *handler) handleReroll(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if g.CurrentPlayer != 0 {
		writeError(w, http.StatusBadRequest, "not your turn")
		return
	}
	if !g.Current().CanReroll() {
		writeError(w, http.StatusBadRequest, "cannot reroll")
		return
	}
	if g.Phase != game.PhaseIncome {
		writeError(w, http.StatusBadRequest, "cannot reroll now")
		return
	}

	var rerollReq struct {
		Indices []int `json:"indices"`
	}
	json.NewDecoder(r.Body).Decode(&rerollReq)

	result, err := g.RerollDice(rerollReq.Indices)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logLines := []string{fmt.Sprintf("%s перебрасывает: %s", g.Current().Name, game.FormatDice(result))}

	actLog := g.ActivateCards()
	logLines = append(logLines, actLog...)

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.RerollsTotal.Inc()
	metrics.TurnsTotal.Inc()
	h.pushLog(id, g.Turn, g.CurrentPlayer, "reroll", map[string]interface{}{
		"dice": result.Numbers, "sum": result.Sum, "player_name": g.Current().Name,
	})
	for _, entry := range actLog {
		h.pushLog(id, g.Turn, g.CurrentPlayer, "income", map[string]interface{}{
			"message": entry, "player_name": g.Current().Name,
		})
	}

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, logLines))
}

func (h *handler) handleBuy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Type   string `json:"type"`
		CardID string `json:"card_id"`
		Index  int    `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if g.CurrentPlayer != 0 {
		writeError(w, http.StatusBadRequest, "not your turn")
		return
	}
	if g.Phase != game.PhaseBuy {
		writeError(w, http.StatusBadRequest, "cannot buy now")
		return
	}

	var log []string

	switch req.Type {
	case "market":
		for mi := range g.Market {
			if g.Market[mi].Card.ID == req.CardID && g.Market[mi].Count > 0 {
				if err := g.BuyCardFromMarket(mi); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
				card := g.Market[mi].Card
				log = []string{fmt.Sprintf("%s покупает %s", g.Current().Name, card.Name)}
				metrics.CardsBoughtTotal.WithLabelValues(card.ID, card.Name).Inc()
				h.pushLog(id, g.Turn, g.CurrentPlayer, "buy_card", map[string]interface{}{
					"card_name": card.Name, "card_id": card.ID,
					"cost": card.Price, "player_name": g.Current().Name,
				})
				break
			}
		}

	case "landmark":
		landmarks := g.AvailableLandmarks()
		if req.Index < 0 || req.Index >= len(landmarks) {
			writeError(w, http.StatusBadRequest, "invalid landmark index")
			return
		}
		lm := landmarks[req.Index]
		if err := g.BuyLandmark(lm.ID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log = []string{fmt.Sprintf("%s покупает достопримечательность %s", g.Current().Name, lm.Name)}
		metrics.LandmarksBoughtTotal.WithLabelValues(lm.ID, lm.Name).Inc()
		h.pushLog(id, g.Turn, g.CurrentPlayer, "buy_landmark", map[string]interface{}{
			"landmark_name": lm.Name, "landmark_id": lm.ID,
			"cost": lm.Price, "player_name": g.Current().Name,
		})

	case "skip":
		g.Phase = game.PhaseEnd
		log = []string{fmt.Sprintf("%s пропускает ход", g.Current().Name)}
		h.pushLog(id, g.Turn, g.CurrentPlayer, "skip", map[string]interface{}{
			"player_name": g.Current().Name,
		})

	default:
		writeError(w, http.StatusBadRequest, "invalid type: market, landmark, or skip")
		return
	}

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, log))
}

func (h *handler) handleEndTurn(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g, err := h.loadGame(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if g.CurrentPlayer != 0 {
		writeError(w, http.StatusBadRequest, "not your turn")
		return
	}
	if g.Phase != game.PhaseBuy && g.Phase != game.PhaseEnd {
		writeError(w, http.StatusBadRequest, "cannot end turn now")
		return
	}

	log := []string{}

	g.EndTurn()
	h.pushLog(id, g.Turn, g.CurrentPlayer, "end_turn", map[string]interface{}{
		"player_name": g.Current().Name,
	})

	for g.CurrentPlayer != 0 {
		if g.CheckWin() != nil {
			break
		}

		playerName := g.Current().Name
		turnLog := g.DoAITurn()
		log = append(log, turnLog...)

		h.pushLog(id, g.Turn, g.CurrentPlayer, "ai_turn", map[string]interface{}{
			"player_name": playerName,
		})

		if g.CheckWin() != nil {
			break
		}
	}

	if winner := g.CheckWin(); winner != nil {
		log = append(log, fmt.Sprintf("ПОБЕДИТЕЛЬ: %s!", winner.Name))
		metrics.GamesCompleted.Inc()
		metrics.ActiveGames.Dec()
		h.pushLog(id, g.Turn, winner.ID, "win", map[string]interface{}{
			"player_id": winner.ID, "player_name": winner.Name,
		})
		h.store.SetSessionCompleted(id)
	}

	log = append(log, fmt.Sprintf("Ход перешёл к %s", g.Current().Name))

	if err := h.saveGame(id, g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, h.stateResponse(id, g, log))
}
