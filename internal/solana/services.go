package solana

import (
	"defi-intel/cmd/api/websocket"
	"defi-intel/internal/types"
	solClient "github.com/blocto/solana-go-sdk/client"
	"github.com/streadway/amqp"
	"net/http"
	"sync"
	"sync/atomic"
)

type SolanaService struct {
	cli *solClient.Client
}

type TokenFinder struct {
	cache  TockenCache
	solCli *SolanaService
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
	solCli *SolanaService
	repo   TokensAndPairsRepo
	pRepo  SwapsRepo

	Wg        sync.WaitGroup
	TxChan    chan types.SolanaBlockTx
	Websocket *websocket.WebSocketServer
}

type BackfillService struct {
	solCli       *SolanaService
	pRepo        SwapsRepo
	queueHandler *SolanaQueueHandler
}

type SolanaQueueHandler struct {
	txHandler *TxHandler
	conn      *amqp.Connection
	ch        *amqp.Channel
	mu        sync.Mutex
	pRepo     SwapsRepo
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
