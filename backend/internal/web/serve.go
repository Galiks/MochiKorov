package web

import (
	"net/http"

	"mochi_korov/internal/store"
)

func RegisterGameRoutes(mux *http.ServeMux, st store.Store) {
	h := &handler{store: st, lokiURL: lokiURL()}

	mux.HandleFunc("POST /api/game/{id}/start", h.handleStart)
	mux.HandleFunc("GET /api/game/{id}/state", h.handleState)
	mux.HandleFunc("POST /api/game/{id}/roll", h.handleRoll)
	mux.HandleFunc("POST /api/game/{id}/collect", h.handleCollect)
	mux.HandleFunc("POST /api/game/{id}/reroll", h.handleReroll)
	mux.HandleFunc("POST /api/game/{id}/buy", h.handleBuy)
	mux.HandleFunc("POST /api/game/{id}/end-turn", h.handleEndTurn)
}
