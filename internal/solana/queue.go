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

func NewSolanaQueueHandler(txHandler *TxHandler, pRepo SwapsRepo) *QueueHandler {
	qh := &QueueHandler{
		txHandler: txHandler,
		pRepo:     pRepo,
		workers:   200,
	}
	qh.connectToRabbitMQ()

	return qh
}

func (qh *QueueHandler) connectToRabbitMQ() {
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

	log.Printf("Connected to RabbitMQ (is closed: %v)", qh.conn.IsClosed())
}

func (qh *QueueHandler) Close() {
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

func (qh *QueueHandler) reconnect() bool {
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

func (qh *QueueHandler) setPrefetch(value int) {
	err := qh.ch.Qos(
		value,
		0,
		false,
	)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
	}
}

func (qh *QueueHandler) ListenToSolanaQueue(ctx context.Context) {
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

	qh.setPrefetch(100)

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

	qh.workerWg.Wait()
	runtime.GC()
}

func (qh *QueueHandler) startWorker(workerID int, parentCtx context.Context) {
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

func (qh *QueueHandler) stopWorker(workerID int) {
	qh.mu.Lock()
	defer qh.mu.Unlock()

	if cancel, exists := qh.workerPool[workerID]; exists {
		cancel()
		delete(qh.workerPool, workerID)
		log.Printf("Worker %d stopped", workerID)
	}
}

func (qh *QueueHandler) AddToSolanaQueue(blockData types.BlockData) {
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

func (qh *QueueHandler) solanaWorker(ctx context.Context) {
	defer qh.workerWg.Done()
	for {
		select {
		case x, ok := <-qh.rabbitChan:
			if !ok {
				log.Println("Channel closed")
				return
			}

			blockData := types.BlockData{}
			unmarshallErr := blockData.UnmarshalJSON(x.Body)
			if unmarshallErr != nil {
				log.Printf("Failed to unmarshal tx: %v", unmarshallErr)
				nackErr := x.Nack(false, false)
				if nackErr != nil {
					log.Printf("Failed to nack message: %v", nackErr)
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

			for i := 0; i < txWorkers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for tx := range txChan {
						processedSwaps, err := qh.txHandler.ProcessTransaction(ctx, &tx, blockData.Timestamp, blockData.Block, blockData.IgnoreWS)
						if err != nil {
							continue
						}
						swapsChan <- processedSwaps

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
				close(swapsChan)
			}()

			swaps := make([]types.SwapLog, 0)
			for result := range swapsChan {
				swaps = append(swaps, result...)
			}

			if len(swaps) > 0 {
				err := qh.insertBatch(ctx, swaps)
				if err != nil {
					log.Printf("Failed to insert swaps: %v", err)
					return
				}
			}

			err := x.Ack(false)
			if err != nil {
				log.Printf("Failed to ack message: %v", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (qh *QueueHandler) insertBatch(ctx context.Context, swaps []types.SwapLog) error {
	const maxRetries = 3

	for retry := 0; retry < maxRetries; retry++ {
		log.Printf("Inserting swaps batch (attempt %d) %+v", retry+1, swaps[0])
		if err := qh.pRepo.InsertSwaps(ctx, swaps); err != nil {
			log.Printf("Failed to insert swaps batch (attempt %d): %v", retry+1, err)
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to insert swaps after %d retries", maxRetries)
}
