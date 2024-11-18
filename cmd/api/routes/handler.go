package routes

import (
	"blocsy/internal/types"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type SolanaTokenFinder interface {
	FindToken(ctx context.Context, address string) (*types.Token, error)
}

type SolanaPairFinder interface {
	FindPair(ctx context.Context, address string) (*types.Pair, *types.QuoteToken, error)
}

type PriceTrackers interface {
	GetUSDPrice(symbol string) float64
}

type SwapsRepo interface{}

type Client struct {
	conn    *websocket.Conn
	wallets map[string]bool
}

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

func (h *Handler) GetHttpHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(120 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Use(APIKeyMiddleware)
		r.Use(RateLimitMiddleware)
	})

	return r
}
