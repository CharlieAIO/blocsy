package routes

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	pricer      PriceTrackers
	tokenFinder SolanaTokenFinder
	pairFinder  SolanaPairFinder
	swapsRepo   SwapsRepo
	nodes       []Node
}

func NewHandler(pricer PriceTrackers, tokenFinder SolanaTokenFinder, pairFinder SolanaPairFinder, swapsRepo SwapsRepo, nodes []Node) *Handler {

	return &Handler{
		pricer:      pricer,
		tokenFinder: tokenFinder,
		pairFinder:  pairFinder,
		swapsRepo:   swapsRepo,
		nodes:       nodes,
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

	fileServer := http.FileServer(http.Dir("/app/docs"))
	r.Get("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/swagger", fileServer).ServeHTTP(w, r)
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
	))

	r.Route("/v1", func(r chi.Router) {
		r.Use(APIKeyMiddleware)
		r.Use(RateLimitMiddleware)

		r.Get("/search", h.SearchQueryHandler)

		r.Get("/pair/{pair}", h.PairLookupHandler)
		r.Get("/token/{token}", h.TokenLookupHandler)

		r.Get("/check-bundled/{token}", h.CheckBundledHandler)
		r.Get("/find-swap/{token}/{amount}/{timestamp}", h.FindSwapHandler)
		r.Get("/top-traders/{token}", h.TopTradersHandler)

		r.Get("/pnl/{wallet}", h.AggregatedPnlHandler)
		r.Get("/token-pnl/{wallet}", h.TokenPnlHandler)
		r.Get("/activity/{wallet}", h.WalletActivityHandler)
	})

	return r
}
