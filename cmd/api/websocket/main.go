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
	}()

	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	ws.mu.Lock()
	ws.clients[conn] = Client{}
	ws.mu.Unlock()

	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to client %v: %v", conn.LocalAddr(), err)
				break
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error from client %v: %v", conn.LocalAddr(), err)
			break
		}

		var clientData Client
		if err := json.Unmarshal(message, &clientData); err != nil {
			log.Printf("Failed to parse client data from %v: %v", conn.LocalAddr(), err)
			break
		}

		if !validateClientType(clientData.ClientType) {
			log.Printf("Invalid client type from client %v: %v", conn.LocalAddr(), clientData.ClientType)
			break
		}

		log.Printf("Client %v connected with type %v | wallets len %d", conn.LocalAddr(), clientData.ClientType, len(clientData.Wallets))

		ws.mu.Lock()
		ws.clients[conn] = clientData
		ws.mu.Unlock()
	}

}

func (ws *WebSocketServer) handleMessages() {
	for {
		select {
		case messageSwaps := <-ws.broadcastSwaps:
			var swaps []types.SwapLog
			if err := json.Unmarshal(messageSwaps, &swaps); err != nil {
				log.Printf("Failed to unmarshal swap message: %v", err)
				continue
			}
			ws.broadcastRelevantSwaps(swaps)

		case messagePFTokens := <-ws.broadcastPFTokens:
			ws.broadcastPFTokensToClients(messagePFTokens)
		}
	}
}

func (ws *WebSocketServer) broadcastRelevantSwaps(swaps []types.SwapLog) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for client, clientData := range ws.clients {
		if clientData.ClientType == "wallet" {
			relevantSwaps := filterRelevantSwaps(swaps, clientData.Wallets)
			if len(relevantSwaps) > 0 {
				message, err := json.Marshal(relevantSwaps)
				if err != nil {
					log.Printf("Failed to marshal relevant swaps: %v", err)
					continue
				}

				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Failed to write message to client: %v", err)
					delete(ws.clients, client)
				}
			}
		}
	}
}

func (ws *WebSocketServer) broadcastPFTokensToClients(messagePFTokens []byte) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for client, clientData := range ws.clients {
		if clientData.ClientType == "pf-tokens" {
			if err := client.WriteMessage(websocket.TextMessage, messagePFTokens); err != nil {
				log.Printf("Failed to write message to client: %v", err)
				delete(ws.clients, client)
			}
		}
	}
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
