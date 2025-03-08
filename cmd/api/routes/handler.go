package routes

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/v1", func(r chi.Router) {
		r.Use(APIKeyMiddleware)
		r.Use(RateLimitMiddleware)

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
