package routes

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	pricer      PriceTrackers
	tokenFinder SolanaTokenFinder
	pairFinder  SolanaPairFinder
	swapsRepo   SwapsRepo
}

func NewHandler(pricer PriceTrackers, tokenFinder SolanaTokenFinder, pairFinder SolanaPairFinder, swapsRepo SwapsRepo) *Handler {
	return &Handler{
		pricer:      pricer,
		tokenFinder: tokenFinder,
		pairFinder:  pairFinder,
		swapsRepo:   swapsRepo,
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) GetHttpHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(120 * time.Second))

	r.Get("/health", HealthCheckHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Use(APIKeyMiddleware)
		r.Use(RateLimitMiddleware)

		r.Get("/pair/{pair}", h.PairLookupHandler)

		r.Get("/pnl-aggregated/{wallet}", h.AggregatedPnlHandler)
		r.Get("/pnl/{wallet}", h.PnlHandler)
	})

	return r
}
