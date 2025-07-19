package interfaces

// EventProducer defines the interface for publishing events
type EventProducer interface {
	PublishEvent(event interface{}) error
	Close() error
}

// WebSocketServer defines the interface for websocket broadcasting

type WebSocketServer interface {
	Start(addr, path string) error
	Stop()
	Broadcast(message []byte)
}

// EventProcessor defines the interface for processing events
type EventProcessor interface {
	ProcessEvent(event interface{}) error
}
