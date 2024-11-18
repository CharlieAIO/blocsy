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

type WebSocketServer struct {
	clients   map[*websocket.Conn][]string
	broadcast chan []byte
	upgrader  websocket.Upgrader
	mu        sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*websocket.Conn][]string),
		broadcast: make(chan []byte),
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
	ws.clients[conn] = []string{}
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

		var wallets []string
		if err := json.Unmarshal(message, &wallets); err == nil {
			ws.mu.Lock()
			ws.clients[conn] = wallets
			ws.mu.Unlock()
			log.Printf("Client %v subscribed to wallets: %v", conn.RemoteAddr(), wallets)
		} else {
			log.Printf("Failed to parse wallets from client %v: %v", conn.RemoteAddr(), err)
		}
	}
}

func (ws *WebSocketServer) handleMessages() {
	for {
		message := <-ws.broadcast
		var swaps []types.SwapLog

		if err := json.Unmarshal(message, &swaps); err != nil {
			log.Printf("Failed to unmarshal swap message: %v", err)
			continue
		}

		ws.mu.Lock()
		for client, wallets := range ws.clients {
			for _, swap := range swaps {
				if ws.isRelevantTransaction(swap, wallets) {
					if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Printf("Failed to write message to client: %v", err)
						delete(ws.clients, client)
						break
					}
				}
			}
		}
		ws.mu.Unlock()
	}
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
	ws.broadcast <- message
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
