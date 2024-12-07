package service

import (
	"github.com/gorilla/websocket"
	"github.com/wrale/wrale-fleet/user/api/types"
)

type wsService struct {}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() types.WebSocketService {
	return &wsService{}
}

func (s *wsService) HandleConnection(conn *websocket.Conn) error {
	// TODO: Implement real-time updates for v1.0
	return nil
}