package web_socket

import (
	"big2/room"
	"big2/server"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(server *server.Server, w http.ResponseWriter, r *http.Request, pool *ConnectionPool) {

	conn := pool.Get(w, r)
	defer pool.Put(conn)
	// conn, err := upgrader.Upgrade(w, r, nil)
	// if err != nil {
	// 	slog.Error("Error upgrading to WebSocket - " + err.Error())
	// 	return
	// }
	// defer conn.Close()

	player := &room.Player{
		Conn:          conn,
		LastHeartbeat: time.Now(),
	}
	players := server.ListPlayers()
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		for {
			playerID := room.GenerateID()
			if !slices.Contains(players, playerID) {
				player.ID = playerID
				break
			} else {
				continue
			}
		}
		server.AddPlayer(player)
	}
	player.StartHeartbeat()

	var currentRoom *room.Room
	// 检查是否是重连
	server.Mu.Lock()
	roomID, exists := server.PlayerToRoom[playerID]
	server.Mu.Unlock()

	if exists {
		currentRoom = server.GetRoom(roomID)
		currentRoom.ReconnectPlayer(player)
	}

	roomList := server.ListRooms()
	err := conn.WriteJSON(room.Message{
		Type:    "room_list",
		Content: roomList,
	})
	if err != nil {
		slog.Error("Error sending room list - " + err.Error())
		return
	}

	for {
		var msg room.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			slog.Error("Error reading message - " + err.Error())
			conn.Close()
			break
		}
		slog.Info("msg:", "message", msg)
		switch msg.Type {
		case "create_room":
			// log.Println("in create_room")
			createRoom := server.CreateRoom()
			// log.Println("create_room :", createRoom)
			_, err := server.JoinRoom(createRoom.ID, player)
			if err != nil {
				slog.Error("Error join room message - " + err.Error())
				continue
			}
			err = conn.WriteJSON(room.Message{
				Type:    "room_created",
				Content: createRoom.ID,
			})
			if err != nil {
				slog.Error("Error write message - " + err.Error())
				continue
			}
		case "join_room":
			roomID, ok := msg.Content.(string)
			if !ok {
				conn.WriteJSON(room.Message{
					Type: "error", Content: "Invalid room ID",
				})
				continue
			}

			ok, err := server.JoinRoom(roomID, player)
			if err != nil {
				conn.WriteJSON(room.Message{
					Type: "error", Content: err.Error(),
				})
				continue
			}
			if ok {
				conn.WriteJSON(room.Message{Type: "room_joined", Content: roomID})
			} else {
				conn.WriteJSON(room.Message{
					Type: "error", Content: "Failed to join room",
				})
			}
		case "list_room":
			rooms := server.ListRooms()
			conn.WriteJSON(room.Message{
				Type: "room_list", Content: rooms,
			})
		case "list_player":
			players := server.ListPlayers()
			conn.WriteJSON(room.Message{
				Type: "player_list", Content: players,
			})
		case "list_room_player":
			playerRoom := player.Room
			if playerRoom == nil {
				conn.WriteJSON(room.Message{
					Type: "error", Content: "Invalid room ID",
				})
				continue
			}

			players := playerRoom.ListPlayers()
			conn.WriteJSON(room.Message{
				Type: "room_player_list", Content: players,
			})
		case "broadcast":
			msg, ok := msg.Content.(string)
			if !ok {
				conn.WriteJSON(room.Message{
					Type: "error", Content: "broadcast error",
				})
				continue
			}

			server.Broadcast(room.Message{
				Type:    "broadcast_all",
				Content: msg,
			})
		case "game_action":
			slog.Debug("Received game action from player %s: %v", player.ID, msg.Content)
		default:
			slog.Debug("Unknown message type: %s", "msg.Type", msg.Type)
		}

	}

}
