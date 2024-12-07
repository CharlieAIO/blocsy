package websocket

import (
	"blocsy/cmd/api/routes"
	"blocsy/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	ClientType string   `json:"clientType"`
	Wallets    []string `json:"wallets"`
}

type WebSocketServer struct {
	clients           map[*websocket.Conn]Client
	broadcastSwaps    chan []byte
	broadcastPFTokens chan []byte
	upgrader          websocket.Upgrader
	mu                sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:           make(map[*websocket.Conn]Client),
		broadcastSwaps:    make(chan []byte),
		broadcastPFTokens: make(chan []byte),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (ws *WebSocketServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer func() {
		ws.mu.Lock()
		delete(ws.clients, conn)
		ws.mu.Unlock()
		conn.Close()
		log.Printf("Connection closed for client: %v", conn.RemoteAddr())
	}()

	ws.mu.Lock()
	ws.clients[conn] = Client{}
	ws.mu.Unlock()

	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to client %v: %v", conn.RemoteAddr(), err)
				break
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error from client %v: %v", conn.RemoteAddr(), err)
			break
		}

		var clientData Client
		if err := json.Unmarshal(message, &clientData); err == nil {
			if !validateClientType(clientData.ClientType) {
				log.Printf("Invalid client type from client %v: %v", conn.RemoteAddr(), clientData.ClientType)
				break
			}
			ws.mu.Lock()
			ws.clients[conn] = clientData
			ws.mu.Unlock()
		} else {
			log.Printf("Failed to parse wallets from client %v: %v", conn.RemoteAddr(), err)
		}
	}
}

func (ws *WebSocketServer) handleMessages() {
	for {
		messageSwaps := <-ws.broadcastSwaps
		var swaps []types.SwapLog
		if err := json.Unmarshal(messageSwaps, &swaps); err != nil {
			log.Printf("Failed to unmarshal swap message: %v", err)
			continue
		}

		messagePFTokens := <-ws.broadcastPFTokens
		var tokens []types.PumpFunCreation
		if err := json.Unmarshal(messagePFTokens, &tokens); err != nil {
			log.Printf("Failed to unmarshal pf tokens message: %v", err)
			continue
		}

		ws.mu.Lock()
		for client, clientData := range ws.clients {
			if clientData.ClientType == "wallet" {
				for _, swap := range swaps {
					if ws.isRelevantTransaction(swap, clientData.Wallets) {
						if err := client.WriteMessage(websocket.TextMessage, messageSwaps); err != nil {
							log.Printf("Failed to write message to client: %v", err)
							delete(ws.clients, client)
							break
						}
					}
				}
			}
			if clientData.ClientType == "pf-tokens" {
				if err := client.WriteMessage(websocket.TextMessage, messagePFTokens); err != nil {
					log.Printf("Failed to write message to client: %v", err)
					delete(ws.clients, client)
					break
				}
			}

		}
		ws.mu.Unlock()
	}
}

func validateClientType(clientType string) bool {
	return clientType == "wallet" || clientType == "pf-tokens"
}

func (ws *WebSocketServer) isRelevantTransaction(swap types.SwapLog, wallets []string) bool {
	for _, wallet := range wallets {
		if wallet == swap.Wallet {
			return true
		}
	}
	return false
}

func (ws *WebSocketServer) BroadcastSwaps(swaps []types.SwapLog) {
	message, err := json.Marshal(swaps)
	if err != nil {
		log.Printf("Failed to marshal swap: %v", err)
		return
	}
	ws.broadcastSwaps <- message
}

func (ws *WebSocketServer) BroadcastPumpFunTokens(tokens []types.PumpFunCreation) {
	message, err := json.Marshal(tokens)
	if err != nil {
		log.Printf("Failed to marshal pf tokens: %v", err)
		return
	}
	ws.broadcastPFTokens <- message
}

func (ws *WebSocketServer) RegisterRoutes(r chi.Router) {
	r.With(routes.APIKeyMiddleware).HandleFunc("/v1/ws", ws.handleConnections)
}

func (ws *WebSocketServer) Start() {
	r := chi.NewRouter()
	ws.RegisterRoutes(r)

	go ws.handleMessages()

	server := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	log.Println("WebSocket server started on :8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
