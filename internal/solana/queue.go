package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	queueName = "solana-tx"
	txWorkers = 10
)

func NewSolanaQueueHandler(txHandler *TxHandler, pRepo SwapsRepo) *SolanaQueueHandler {
	qh := &SolanaQueueHandler{
		txHandler: txHandler,
		pRepo:     pRepo,
		workers:   500,
	}
	qh.connectToRabbitMQ()

	return qh
}

func (qh *SolanaQueueHandler) connectToRabbitMQ() {
	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		os.Getenv("RABBIT_MQ_USER"),
		os.Getenv("RABBIT_MQ_PASS"),
		os.Getenv("RABBIT_MQ_HOST"),
		os.Getenv("RABBIT_MQ_PORT"),
		os.Getenv("RABBIT_MQ_VHOST"),
	)

	var err error
	qh.conn, err = amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	qh.ch, err = qh.conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
}

func (qh *SolanaQueueHandler) Close() {
	if qh.ch != nil {
		if err := qh.ch.Close(); err != nil {
			log.Printf("Failed to close channel: %v", err)
		}
	}
	if qh.conn != nil {
		if err := qh.conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}
}

func (qh *SolanaQueueHandler) monitorQueue(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queue, err := qh.ch.QueueInspect(queueName)
			if err != nil {
				log.Printf("Failed to inspect queue: %v", err)
				continue
			}

			qh.adjustWorkers(queue.Messages)
		}
	}
}

func (qh *SolanaQueueHandler) adjustWorkers(queueSize int) {
	var newNumWorkers int

	switch {
	case queueSize > 1500:
		newNumWorkers = 1000
	case queueSize > 1000:
		newNumWorkers = 950
	case queueSize > 750:
		newNumWorkers = 900
	case queueSize > 500:
		newNumWorkers = 800
	default:
		newNumWorkers = 750

	}

	qh.mu.Lock()
	currentWorkers := len(qh.workerPool)
	qh.mu.Unlock()

	if newNumWorkers > currentWorkers {
		for i := currentWorkers; i < newNumWorkers; i++ {
			qh.startWorker(i, qh.ctx)
		}
		log.Printf("Increased workers from %d to %d", currentWorkers, newNumWorkers)
	} else if newNumWorkers < currentWorkers {
		for i := currentWorkers - 1; i >= newNumWorkers; i-- {
			qh.stopWorker(i)
		}
		log.Printf("Decreased workers from %d to %d", currentWorkers, newNumWorkers)
	}
}

func (qh *SolanaQueueHandler) reconnect() bool {
	qh.Close()
	backoff := time.Second
	for i := 0; i < 5; i++ {
		qh.connectToRabbitMQ()
		if qh.conn != nil && qh.ch != nil {
			log.Println("Successfully reconnected to RabbitMQ")
			return true
		}
		log.Printf("Reconnection attempt %d failed", i+1)
		time.Sleep(backoff)
		backoff *= 2
	}
	log.Println("Failed to reconnect to RabbitMQ after multiple attempts")
	return false
}

func (qh *SolanaQueueHandler) setPrefetch(value int) {
	err := qh.ch.Qos(
		value,
		0,
		false,
	)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
	}
}

func (qh *SolanaQueueHandler) ListenToSolanaQueue(ctx context.Context) {
	blocked := qh.conn.NotifyBlocked(make(chan amqp.Blocking))
	go func() {
		for b := range blocked {
			if b.Active {
				log.Printf("Connection blocked: %s", b.Reason)
			} else {
				log.Println("Connection unblocked")
			}
		}
	}()

	qh.ctx = ctx
	qh.rabbitChan = make(chan amqp.Delivery, qh.workers)
	qh.workerPool = make(map[int]context.CancelFunc)

	for i := 0; i < qh.workers; i++ {
		qh.startWorker(i, ctx)
	}

	qh.setPrefetch(3000)

	q, err := qh.ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := qh.ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			select {
			case <-ctx.Done():
				return
			default:
				qh.rabbitChan <- d
			}
		}
		close(qh.rabbitChan)
	}()
	//go qh.monitorQueue(ctx)

	qh.workerWg.Wait()

	runtime.GC()
}

func (qh *SolanaQueueHandler) startWorker(workerID int, parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	qh.mu.Lock()
	qh.workerPool[workerID] = cancel
	qh.mu.Unlock()

	qh.workerWg.Add(1)
	go func() {
		defer qh.workerWg.Done()
		qh.solanaWorker(ctx)
	}()
	log.Printf("Worker %d started", workerID)
}

func (qh *SolanaQueueHandler) stopWorker(workerID int) {
	qh.mu.Lock()
	defer qh.mu.Unlock()

	if cancel, exists := qh.workerPool[workerID]; exists {
		cancel()
		delete(qh.workerPool, workerID)
		log.Printf("Worker %d stopped", workerID)
	}
}

func (qh *SolanaQueueHandler) AddToSolanaQueue(blockData types.BlockData) {
	qh.mu.Lock()
	defer qh.mu.Unlock()

	if qh.conn == nil || qh.conn.IsClosed() || qh.ch == nil {
		log.Println("Connection or channel is closed, attempting to reconnect...")
		if !qh.reconnect() {
			log.Println("Reconnection failed. Exiting AddToSolanaQueue")
			return
		}
	}

	q, err := qh.ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return
	}

	blockBytes, err := blockData.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal SolanaBlockTx to JSON: %v", err)
		return
	}

	err = qh.ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        blockBytes,
		})
	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		return
	}
}

func (qh *SolanaQueueHandler) solanaWorker(ctx context.Context) {
	defer qh.workerWg.Done()
	for {
		select {
		case x, ok := <-qh.rabbitChan:
			if !ok {
				log.Println("Channel closed")
				return
			}

			blockData := types.BlockData{}
			err := blockData.UnmarshalJSON(x.Body)
			if err != nil {
				log.Printf("Failed to unmarshal tx: %v", err)
				err := x.Nack(false, false)
				if err != nil {
					log.Printf("Failed to nack message: %v", err)
					continue
				}
				continue
			}

			if qh.conn.IsClosed() || qh.ch == nil {
				log.Println("Connection is closed or channel is nil, exiting worker...")

				if !qh.reconnect() {
					log.Println("Failed to reconnect, exiting worker...")
					return
				}
			}

			txChan := make(chan types.SolanaTx, len(blockData.Transactions))
			swapsChan := make(chan []types.SwapLog)
			var wg sync.WaitGroup

			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for tx := range txChan {
						if tx.Meta.Err != nil {
							continue
						}
						_swaps, err := qh.txHandler.ProcessTransaction(ctx, &tx, blockData.Timestamp, blockData.Block, blockData.IgnoreWS)
						if err != nil {
							continue
						}
						swapsChan <- _swaps
					}
				}()
			}

			go func() {
				for _, tx := range blockData.Transactions {
					txChan <- tx
				}
				close(txChan)
			}()

			go func() {
				wg.Wait()

				err = qh.insertBatch(ctx, swapsChan)
				if err != nil {
					log.Printf("Failed to insert swaps: %v", err)
					return
				}

				err = x.Ack(false)
				if err != nil {
					log.Printf("Failed to ack message: %v", err)
					return
				}

				close(swapsChan)
			}()

		case <-ctx.Done():
			return
		}
	}
}

func (qh *SolanaQueueHandler) insertBatch(ctx context.Context, swapsChan chan []types.SwapLog) error {
	const maxRetries = 3
	swaps := make([]types.SwapLog, 0)
	for result := range swapsChan {
		swaps = append(swaps, result...)
	}
	log.Printf("Inserting %d swaps", len(swaps))

	for retry := 0; retry < maxRetries; retry++ {
		if err := qh.pRepo.InsertSwaps(ctx, swaps); err != nil {
			log.Printf("Failed to insert swaps batch (attempt %d): %v", retry+1, err)
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to insert swaps after %d retries", maxRetries)
}
