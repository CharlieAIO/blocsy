package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mailru/easyjson"
)

func NewNode(name string, url string) *Node {
	return &Node{
		name:    name,
		url:     url,
		counter: atomic.Int64{},
		cli: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (n *Node) GetParsedLogs(rawLogs []string) []types.LogDetails {
	return GetLogs(rawLogs)
}

func (n *Node) GetBlockMessage(ctx context.Context, blockNumber int) (*types.HTTPBlockMessage, error) {
	payload := fmt.Sprintf(`{
        "jsonrpc": "2.0",
        "id": "1",
        "method": "getBlock",
        "params": [
            %d,
            {
                "commitment": "confirmed",
                "encoding": "json",
                "showRewards": false,
                "transactionDetails": "full",
                "maxSupportedTransactionVersion": 2
            }
        ]
    }`, blockNumber)

	req, err := http.NewRequest(http.MethodPost, n.url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := n.sendRequestWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("error making request after retries: %v", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("received nil response")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	blockMessage := types.HTTPBlockMessage{}
	if err := easyjson.UnmarshalFromReader(resp.Body, &blockMessage); err != nil {
		return nil, fmt.Errorf("error decoding message: %w, block: %d, body size: %d KB", err, blockNumber, resp.ContentLength)
	}

	return &blockMessage, nil
}

func (n *Node) GetTx(ctx context.Context, hash string) (*types.SolanaTx, error) {
	payload := fmt.Sprintf(`{
        "jsonrpc": "2.0",
        "id": "1",
        "method": "getTransaction",
        "params": [
            "%s",
            {
                "commitment": "confirmed",
                "encoding": "json",
                "showRewards": false,
                "transactionDetails": "full",
                "maxSupportedTransactionVersion": 2
            }
        ]
    }`, hash)

	req, err := http.NewRequest(http.MethodPost, n.url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := n.sendRequestWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("error making request after retries: %v", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("received nil response")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode)
	}

	txMessage := types.HTTPTxMessage{}
	if err := easyjson.UnmarshalFromReader(resp.Body, &txMessage); err != nil {
		return nil, fmt.Errorf("error decoding message: %w, body size: %d KB", err, resp.ContentLength)
	}

	return &txMessage.Result, nil
}

func (n *Node) sendRequestWithRetry(req *http.Request, retries int) (*http.Response, error) {
	var res *http.Response
	var err error

	for attempts := 0; attempts < retries; attempts++ {
		res, err = n.cli.Do(req)
		if err == nil && res.StatusCode == http.StatusOK {
			return res, nil
		}

		log.Printf("Retrying request, attempt %d, error: %v", attempts+1, err)
		time.Sleep(time.Duration(100*(attempts+1)) * time.Millisecond)
	}

	return nil, err
}
