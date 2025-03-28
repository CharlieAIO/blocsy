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

	processor         *TokenProcessor
	mintBurnProcessor *MintBurnProcessor
}
type TokenProcessor struct {
	queue chan string
	seen  sync.Map
	wg    sync.WaitGroup
}

type MintBurnProcessor struct {
	queue       chan MintBurnProcessorQueue
	seen        sync.Map
	activeLocks sync.Map
	wg          sync.WaitGroup
}

type MintBurnProcessorQueue struct {
	address string
	amount  string
	Type    string
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
	queueHandler *QueueHandler
	nodeUrls     []*Node
}

type QueueHandler struct {
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

type BlockListener struct {
	grpcAddress  string
	queueHandler *QueueHandler
}
