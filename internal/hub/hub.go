package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	sync.RWMutex
	Connections map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		Connections: make(map[string]*websocket.Conn),
	}
}
