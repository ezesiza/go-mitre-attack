package websockets

import (
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

func TestNewWebSocketServer(t *testing.T) {
	tests := []struct {
		name string
		want *WebSocketServer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWebSocketServer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWebSocketServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebSocketServer_Broadcast(t *testing.T) {
	type fields struct {
		upgrader websocket.Upgrader
		clients  map[*websocket.Conn]bool
		lock     sync.Mutex
	}
	type args struct {
		message []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WebSocketServer{
				upgrader: tt.fields.upgrader,
				clients:  tt.fields.clients,
				lock:     tt.fields.lock,
			}
			s.Broadcast(tt.args.message)
		})
	}
}

func TestWebSocketServer_HandleWs(t *testing.T) {
	type fields struct {
		upgrader websocket.Upgrader
		clients  map[*websocket.Conn]bool
		lock     sync.Mutex
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WebSocketServer{
				upgrader: tt.fields.upgrader,
				clients:  tt.fields.clients,
				lock:     tt.fields.lock,
			}
			s.HandleWs(tt.args.w, tt.args.r)
		})
	}
}

func TestWebSocketServer_Start(t *testing.T) {
	type fields struct {
		upgrader websocket.Upgrader
		clients  map[*websocket.Conn]bool
		lock     sync.Mutex
	}
	type args struct {
		addr string
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WebSocketServer{
				upgrader: tt.fields.upgrader,
				clients:  tt.fields.clients,
				lock:     tt.fields.lock,
			}
			if err := s.Start(tt.args.addr, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebSocketServer_Stop(t *testing.T) {
	type fields struct {
		upgrader websocket.Upgrader
		clients  map[*websocket.Conn]bool
		lock     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WebSocketServer{
				upgrader: tt.fields.upgrader,
				clients:  tt.fields.clients,
				lock:     tt.fields.lock,
			}
			s.Stop()
		})
	}
}

func TestWebSocketServer_readPump(t *testing.T) {
	type fields struct {
		upgrader websocket.Upgrader
		clients  map[*websocket.Conn]bool
		lock     sync.Mutex
	}
	type args struct {
		conn *websocket.Conn
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WebSocketServer{
				upgrader: tt.fields.upgrader,
				clients:  tt.fields.clients,
				lock:     tt.fields.lock,
			}
			s.readPump(tt.args.conn)
		})
	}
}
