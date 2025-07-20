package websockets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	lock     sync.Mutex
	srv      *http.Server // Add http.Server for graceful shutdown
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *WebSocketServer) HandleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade error: %v", err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.clients[conn] = true

	go s.readPump(conn)
}

func (s *WebSocketServer) readPump(conn *websocket.Conn) {
	defer func() {
		s.lock.Lock()
		delete(s.clients, conn)
		s.lock.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Websocket read error: %v", err)
			break
		}
	}
}

func (s *WebSocketServer) Broadcast(message []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Websocket write error: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

// Compile-time check to ensure WebSocketServer implements the interface used by events/processor.go
var _ interface{ Broadcast([]byte) } = (*WebSocketServer)(nil)

func (s *WebSocketServer) Start(addr, path string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(path, s.HandleWs)
	s.srv = &http.Server{Addr: addr, Handler: mux}
	fmt.Printf("Websocket server listening at ws://%s%s", addr, path)
	return s.srv.ListenAndServe()
}

func (s *WebSocketServer) Stop() {
	if s.srv != nil {
		s.srv.Shutdown(context.Background())
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	for client := range s.clients {
		client.Close()
		delete(s.clients, client)
	}
}
