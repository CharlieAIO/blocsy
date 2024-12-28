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

	processor *TokenProcessor
}
type TokenProcessor struct {
	queue chan string
	seen  sync.Map
	wg    sync.WaitGroup
}

type PairsService struct {
	cache       PairsCache
	tokenFinder *TokenFinder
	solSvc      *SolanaService
	repo        PairsRepo

	processor *PairProcessor
}
type PairProcessor struct {
	queue chan PairProcessorQueue
	seen  sync.Map
	wg    sync.WaitGroup
}
type PairProcessorQueue struct {
	address string
	token   *string
}

type TxHandler struct {
	sh    *SwapHandler
	repo  TokensAndPairsRepo
	pRepo SwapsRepo

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

type SwapHandler struct {
	tf SolanaTokenFinder
	pf SolanaPairFinder
}

type Node struct {
	name    string
	url     string
	cli     *http.Client
	counter atomic.Int64
	timings atomic.Int64
}
