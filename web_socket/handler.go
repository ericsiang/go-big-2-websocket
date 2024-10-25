package web_socket

import (
	"big2/big2_card"
	"big2/handle_errors"
	"big2/player"
	"big2/room"
	"big2/server"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(server *server.Server, w http.ResponseWriter, r *http.Request) {
	// 建立 webscket conn
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("[HandleWebSocket Error]", "upgrader", "Error upgrading to WebSocket - "+err.Error())
		return
	}
	defer func() {
		slog.Info("[conn close]")
		conn.Close()
	}()
	// 建立 Player
	players := server.ListPlayers()
	// 從 Query 取得 player_id ， 用來重連使用
	playerID := r.URL.Query().Get("player_id")
	// 检查是否是重连
	server.Mu.Lock()
	roomID, exists := server.PlayerToRoom[playerID]
	slog.Info("[HandleWebSocket]", "PlayerToRoom", server.PlayerToRoom)
	server.Mu.Unlock()
	var newPlayer *player.Player
	slog.Info("[HandleWebSocket]", "newPlayer", newPlayer)
	if exists {
		newPlayer = player.NewPlayer(playerID, conn)
		currentRoom := server.GetRoom(roomID)
		if currentRoom != nil {
			slog.Error("[HandleWebSocket Error]", "Reconnect-GetRoom", handle_errors.ErrRoomNotFound.Error())
			err := send(conn, "error", handle_errors.ErrRoomNotFound.Error())
			if err != nil {
				return
			}
			currentRoom.ReconnectPlayer(newPlayer)
		}
	} else {
		for {
			playerID := room.GenerateID()
			if !slices.Contains(players, playerID) {
				newPlayer = player.NewPlayer(playerID, conn)
				server.AddPlayer(newPlayer)
				newPlayer.StartHeartbeat()
				break
			} else {
				continue
			}
		}
	}
	slog.Info("[HandleWebSocket]", "after newPlayer", newPlayer)
	roomList := server.ListRooms()
	err = send(conn, "room_list", roomList)
	if err != nil {
		return
	}

	for {
		var msg room.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			slog.Error("[HandleWebSocket Error]", "reading_message_error", err.Error())
			break
		}
		// slog.Info("[msg]", "message", msg)
		switch msg.Type {
		case "heartbeat_resp":
			newPlayer.HandleHeartbeatResponse()
		case "create_room":
			createRoom, err := server.CreateRoom(newPlayer)
			if err != nil {
				slog.Warn("[HandleWebSocket Warn]", "CreateRoom", err.Error())
				err = send(conn, "create_room_error", err.Error())
				if err != nil {
					return
				}
				continue
			}
			err = server.JoinRoom(createRoom.GetID(), newPlayer)
			if err != nil {
				slog.Error("[HandleWebSocket Error]", "create_room_error", err.Error())
				continue
			}
			err = send(conn, "room_created", createRoom.GetID())
			if err != nil {
				return
			}
		case "join_room":
			roomID, ok := msg.Content.(string)
			if !ok {
				slog.Warn("[HandleWebSocket Warn]", "join_room", handle_errors.ErrRoomNotFound.Error())
				err = send(conn, "join_room_error", handle_errors.ErrRoomNotFound.Error())
				if err != nil {
					return
				}
				continue
			}

			err := server.JoinRoom(roomID, newPlayer)
			if err != nil {
				slog.Warn("[HandleWebSocket Warn]", "join_room[JoinRoom]", err.Error())
				err = send(conn, "join_room_error", err.Error())
				if err != nil {
					return
				}
				continue
			}
			err = send(conn, "room_joined", roomID)
			if err != nil {
				return
			}
		case "leave_room":

		case "list_room":
			rooms := server.ListRooms()
			err = send(conn, "room_list", rooms)
			if err != nil {
				return
			}
		case "list_player":
			players := server.ListPlayers()
			err = send(conn, "player_list", players)
			if err != nil {
				return
			}
		case "list_room_player":
			playerRoom := newPlayer.GetRoom()
			if playerRoom == nil {
				err = send(conn, "list_room_player_error", handle_errors.ErrRoomNotFound.Error())
				if err != nil {
					return
				}
				continue
			}
			slog.Info("[HandleWebSocket]", "list_room_player", playerRoom)
			players := playerRoom.GetPlayers()
			err = send(conn, "room_player_list", players)
			if err != nil {
				return
			}
		case "broadcast":
			msg, ok := msg.Content.(string)
			if !ok {
				err = send(conn, "broadcast_error", "broadcast message content error")
				if err != nil {
					return
				}
			}
			server.Broadcast("broadcast_all", msg)
		case "game_action_first_play":
			slog.Info("[HandleWebSocket]", "game_action_first_play", "Received game action from player_id = "+newPlayer.GetID())
			// 將 msg.Content 轉成 big2_card.Card
			cards, err := contentToCard2(msg.Content)
			if err != nil {
				slog.Error("[HandleWebSocket Error]", "game_action_first_play", "contentToCard2", "error", err.Error())
				err = send(conn, "game_action_first_play_error", "content error")
				if err != nil {
					return
				}
				continue
			}
			slog.Info("[HandleWebSocket]", "game_action_first_play", cards)
			// 判斷牌型
			err = server.Game.CheckCard("first", cards)
			if err != nil {
				slog.Warn("[HandleWebSocket Warn]", "game_action_first_play", "CheckCard", "error", err.Error())
				err = send(conn, "game_action_first_play_error", err.Error())
				if err != nil {
					return
				}
				continue
			}

			server.Game.PlayCards(newPlayer, cards)
			server.Game.SetLastPlayerCard(cards)
			server.Game.SetLastPlayer(newPlayer)
			nextPlayer := server.Game.GetNextPlayer()
			send(nextPlayer.GetConn(), "game_action_next_play", "")
		case "game_action_next_play":

		default:
			slog.Debug("Unknown message type: %s", "msg.Type", msg.Type)
		}

	}

}

// 方式一 用 json 處理，代码简洁，易于理解，性能略差
func contentToCard(content interface{}) ([]big2_card.Card, error) {

	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	slog.Info("[contentToCard]", "contentToCard", jsonBytes)
	var cards []big2_card.Card
	if err = json.Unmarshal(jsonBytes, &cards); err != nil {
		return nil, err
	}

	return cards, nil
}

// 方式二 性能更好，代码较长，需要手动处理类型转换
func contentToCard2(content interface{}) ([]big2_card.Card, error) {

	// 首先断言是否为切片
	contentSlice, ok := content.([]interface{})
	if !ok {
		return nil, fmt.Errorf("content is not a slice")
	}

	cards := make([]big2_card.Card, len(contentSlice))
	for i, item := range contentSlice {
		// 斷言每個元素是否為 map[string]interface{} 的 type
		cardMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d is not a map", i)
		}

		// 斷言 suit 為 float64 的 type
		suit, ok := cardMap["suit"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid suit at index %d", i)
		}

		// 斷言 value 為 float64 的 type
		value, ok := cardMap["value"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value at index %d", i)
		}

		// 每個元素的 suit 跟 value 轉成 big2_card.Card 的 type
		cards[i] = big2_card.Card{
			Suit:  big2_card.Suit(int(suit)),
			Value: int(value),
		}
	}

	return cards, nil
}

func send(conn *websocket.Conn, sendType string, content interface{}) error {
	err := conn.WriteJSON(room.Message{
		Type:    sendType,
		Content: content,
	})
	if err != nil {
		slog.Error("[send]", "err", "Error sending message - "+err.Error(), "sendType", sendType, "content", content)
		return err
	}

	return nil
}
