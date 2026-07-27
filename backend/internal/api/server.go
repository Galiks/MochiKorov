package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"mochi_korov/internal/metrics"
	"mochi_korov/internal/store"
	"mochi_korov/internal/web"
)

type Server struct {
	httpServer *http.Server
	store      store.Store
}

func NewServer(addr string, st store.Store, distDir string) *Server {
	metrics.Register()

	mux := http.NewServeMux()
	h := &handler{store: st}

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /api/establishments", h.handleListEstablishments)
	mux.HandleFunc("POST /api/establishments", h.handleCreateEstablishment)
	mux.HandleFunc("GET /api/establishments/{id}", h.handleGetEstablishment)
	mux.HandleFunc("PUT /api/establishments/{id}", h.handleUpdateEstablishment)
	mux.HandleFunc("DELETE /api/establishments/{id}", h.handleDeleteEstablishment)

	mux.HandleFunc("GET /api/landmarks", h.handleListLandmarks)
	mux.HandleFunc("POST /api/landmarks", h.handleCreateLandmark)
	mux.HandleFunc("GET /api/landmarks/{id}", h.handleGetLandmark)
	mux.HandleFunc("PUT /api/landmarks/{id}", h.handleUpdateLandmark)
	mux.HandleFunc("DELETE /api/landmarks/{id}", h.handleDeleteLandmark)

	mux.HandleFunc("GET /api/sessions", h.handleListSessions)
	mux.HandleFunc("POST /api/sessions", h.handleCreateSession)
	mux.HandleFunc("GET /api/sessions/{id}", h.handleGetSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", h.handleDeleteSession)
	mux.HandleFunc("PUT /api/sessions/{id}/game", h.handleSaveSessionGame)
	mux.HandleFunc("GET /api/sessions/{id}/game", h.handleLoadSessionGame)

	mux.HandleFunc("GET /api/card-sets", h.handleListCardSets)
	mux.HandleFunc("POST /api/card-sets", h.handleCreateCardSet)
	mux.HandleFunc("GET /api/card-sets/{id}", h.handleGetCardSet)
	mux.HandleFunc("DELETE /api/card-sets/{id}", h.handleDeleteCardSet)
	mux.HandleFunc("PUT /api/card-sets/{id}/cards", h.handleSaveCardSetCards)
	mux.HandleFunc("GET /api/card-sets/{id}/cards", h.handleLoadCardSetCards)
	mux.HandleFunc("POST /api/card-sets/seed", h.handleSeedCardSet)

	web.RegisterGameRoutes(mux, st)

	if info, err := os.Stat(distDir); err == nil && info.IsDir() {
		fileServer := http.FileServer(http.Dir(distDir))
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.ServeFile(w, r, distDir+"/index.html")
				return
			}
			path := distDir + r.URL.Path
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, distDir+"/index.html")
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("MochiKorov API server. Frontend not found. Build it with: cd frontend && npm run build"))
		})
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: withCORS(mux),
		},
		store: st,
	}
}

func (s *Server) Start() error {
	log.Printf("API server starting on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpServer.Shutdown(ctx)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Player-Token")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
