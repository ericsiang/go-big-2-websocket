package web_socket

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnectionPool struct {
	connections chan *websocket.Conn
	maxSize     int
}

func NewConnectionPool(maxSize int) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan *websocket.Conn, maxSize),
		maxSize:     maxSize,
	}
}

func (p *ConnectionPool) Get(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	select {
	case conn := <-p.connections:
		return conn
	default:
		// 創建新的連接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Error upgrading to WebSocket - " + err.Error())
			return nil
		}
		return conn
	}
}

func (p *ConnectionPool) Put(conn *websocket.Conn) {
	select {
	case p.connections <- conn:
		// 連接已放回池中
	default:
		// 池已滿,關閉連接
		conn.Close()
	}
}
