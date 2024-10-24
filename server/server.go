package server

import (
	"big2/game"
	"big2/handle_errors"
	"big2/player"
	"big2/room"
	"big2/shared"
	"log/slog"
	"slices"
	"sync"
)

type Server struct {
	Mu                  sync.Mutex
	Game                shared.Game
	Players             map[string]shared.Player
	DisconnectedPlayers map[string]shared.Player
	Rooms               map[string]shared.Room
	PlayerToRoom        map[string]string // 新增：用于快速查找玩家所在的房间 map[playerID]RoomID
}

func NewServer() *Server {
	return &Server{
		Game:                game.NewBig2Game(),
		Players:             make(map[string]shared.Player),
		DisconnectedPlayers: make(map[string]shared.Player),
		Rooms:               make(map[string]shared.Room),
		PlayerToRoom:        make(map[string]string),
	}
}

func (s *Server) CreateRoom() *room.Room {
	rooms := s.ListRooms()
	s.Mu.Lock()
	defer s.Mu.Unlock()
	slog.Info("[CreateRoom] ")
	for {
		roomID := room.GenerateID()
		if !slices.Contains(rooms, roomID) {
			room := room.NewRoom(roomID, s.Game)
			s.Rooms[roomID] = room
			return room
		} else {
			continue
		}
	}
}

func (s *Server) GetRoom(roomID string) shared.Room {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	return s.Rooms[roomID]
}

func (s *Server) ListRooms() []string {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	slog.Info("[ListRooms] ")
	rooms := make([]string, 0, len(s.Rooms))
	for roomID := range s.Rooms {
		rooms = append(rooms, roomID)
	}

	return rooms
}

func (s *Server) AddPlayer(player shared.Player) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Players[player.GetID()] = player
	slog.Info("[AddPlayer] ", "player", player)
}

func (s *Server) ListPlayers() []string {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	players := make([]string, 0, len(s.Players))
	for ID := range s.Players {
		players = append(players, ID)
	}

	return players
}

func (s *Server) JoinRoom(roomID string, player *player.Player) error {
	slog.Info("[JoinRoom] ")
	room := s.GetRoom(roomID)
	if room == nil {
		return handle_errors.ErrRoomNotFound
	}

	roomPlayers := room.GetPlayers()
	if len(roomPlayers) >= 4 {
		return handle_errors.ErrRoomFull
	}

	room.AddPlayer(player)
	slog.Info("[JoinRoom] ", "players", roomPlayers)

	s.Mu.Lock()
	s.PlayerToRoom[player.GetID()] = roomID
	s.Mu.Unlock()

	if len(roomPlayers) == 4 {
		go room.StartGame()
	}

	return nil
}

func (s *Server) Broadcast(msgType string, content interface{}) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	message := room.NewMessage()
	message.SetMessage(msgType, content)
	for _, player := range s.Players {
		err := player.GetConn().WriteJSON(message)
		if err != nil {
			slog.Error("[Server Broadcast Error]", "broadcasting to player %s: %v", player.GetID(), "error", err.Error())
		}
	}
}
