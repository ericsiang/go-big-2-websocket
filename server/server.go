package server

import (
	"big2/big2_game"
	"big2/handle_errors"
	"big2/player"
	"big2/room"
	"big2/shared"
	"errors"
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
		Game:                big2_game.NewBig2Game(),
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
	slog.Info("[in CreateRoom] ")
	for {
		roomID := room.GenerateID()
		// log.Println("roomID:",roomID)
		if !slices.Contains(rooms, roomID) {
			// log.Println(" in slices")
			room := room.NewRoom(roomID, server.Game)
			s.Rooms[roomID] = room
			return room
		} else {
			continue
		}
	}
}

func (s *Server) GetRoom(roomID string) *room.Room {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	return s.Rooms[roomID]
}

func (s *Server) ListRooms() []string {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	slog.Info("[in ListRooms] ")
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
	slog.Info("[player add] ", "player", player)
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

func (s *Server) JoinRoom(roomID string, player *player.Player) (bool, error) {
	slog.Info("[in JoinRoom] ")
	room := s.GetRoom(roomID)
	if room == nil {
		return false, errors.New("roomID not exist")
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	if len(room.Players) >= 4 {
		return false, handle_errors.ErrRoomFull
	}

	room.Players[player.ID] = player
	slog.Info("[room.Players] ", "players", room.Players)
	// player.Room = room

	s.Mu.Lock()
	s.PlayerToRoom[player.ID] = roomID
	s.Mu.Unlock()

	if len(room.Players) == 4 {
		go room.StartGame()
	}

	return true, nil
}

func (s *Server) Broadcast(message room.Message) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for _, player := range s.Players {
		err := player.Conn.WriteJSON(message)
		if err != nil {
			slog.Error("Error broadcasting to player %s: %v", player.ID, err)
		}
	}
}
