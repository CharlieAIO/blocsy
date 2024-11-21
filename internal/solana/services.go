package solana

import (
	"blocsy/cmd/api/websocket"
	"blocsy/internal/types"
	"context"
	solClient "github.com/blocto/solana-go-sdk/client"
	"github.com/streadway/amqp"
	"net/http"
	"sync"
	"sync/atomic"
)

type SolanaService struct {
	client *solClient.Client
}

type TokenFinder struct {
	cache  TockenCache
	solSvc *SolanaService
	repo   TokensRepo
}

type PairsService struct {
	cache       PairsCache
	tokenFinder *TokenFinder
	solSvc      *SolanaService
	repo        PairsRepo
}

type TxHandler struct {
	sh     *SwapHandler
	solSvc *SolanaService
	repo   TokensAndPairsRepo
	pRepo  SwapsRepo

	Wg        sync.WaitGroup
	TxChan    chan types.SolanaBlockTx
	Websocket *websocket.WebSocketServer
}

type BackfillService struct {
	solSvc       *SolanaService
	pRepo        SwapsRepo
	queueHandler *SolanaQueueHandler
	nodeUrls     []*Node
}

type SolanaQueueHandler struct {
	txHandler *TxHandler
	conn      *amqp.Connection
	ch        *amqp.Channel
	mu        sync.Mutex
	pRepo     SwapsRepo

	ctx context.Context

	rabbitChan chan amqp.Delivery
	workerWg   sync.WaitGroup
	workers    int
	workerPool map[int]context.CancelFunc
}

type Node struct {
	name    string
	url     string
	cli     *http.Client
	counter atomic.Int64
	timings atomic.Int64
}

type SwapHandler struct {
	tf SolanaTokenFinder
	pf SolanaPairFinder
}
