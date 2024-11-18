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
	queueName  = "solana-tx"
	numWorkers = 750
)

func NewSolanaQueueHandler(txHandler *TxHandler, pRepo SwapsRepo) *SolanaQueueHandler {
	qh := &SolanaQueueHandler{
		txHandler: txHandler,
		pRepo:     pRepo,
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

	blockChan := make(chan amqp.Delivery, numWorkers)
	var workerWg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go qh.solanaWorker(ctx, &workerWg, blockChan)
	}

	err := qh.ch.Qos(
		10,
		0,
		false,
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
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
				blockChan <- d
			}
		}
		close(blockChan)
	}()

	workerWg.Wait()

	runtime.GC()
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

func (qh *SolanaQueueHandler) solanaWorker(ctx context.Context, wg *sync.WaitGroup, rabbitChan <-chan amqp.Delivery) {
	defer wg.Done()
	for {
		select {
		case x, ok := <-rabbitChan:
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

			swaps := make([]types.SwapLog, 0)
			for _, tx := range blockData.Transactions {
				_swaps, err := qh.txHandler.ProcessTransaction(ctx, &tx, blockData.Timestamp, blockData.Block)
				if err != nil {
					continue
				}
				swaps = append(swaps, _swaps...)
			}

			err = qh.insertBatch(ctx, swaps)
			if err != nil {
				log.Printf("Failed to insert swaps: %v", err)
				return
			}

			err = x.Ack(false)
			if err != nil {
				log.Printf("Failed to ack message: %v", err)
				continue
			}

		case <-ctx.Done():
			return
		}
	}
}

func (qh *SolanaQueueHandler) insertBatch(ctx context.Context, swaps []types.SwapLog) error {
	const maxRetries = 3

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
