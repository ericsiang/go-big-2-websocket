package server

import (
	"big2/big2_game"
	"big2/room"
	"errors"
	"log/slog"
	"slices"
	"sync"
)

type Server struct {
	Players             map[string]*room.Player
	DisconnectedPlayers map[string]*room.Player
	Rooms               map[string]*room.Room
	PlayerToRoom        map[string]string // 新增：用于快速查找玩家所在的房间 map[playerID]RoomID
	Mu                  sync.Mutex
}

func NewServer() *Server {
	return &Server{
		Players:             make(map[string]*room.Player),
		DisconnectedPlayers: make(map[string]*room.Player),
		Rooms:               make(map[string]*room.Room),
		PlayerToRoom:        make(map[string]string),
	}
}

func (s *Server) CreateRoom() *room.Room {
	rooms := s.ListRooms()
	s.Mu.Lock()
	defer s.Mu.Unlock()
	slog.Info("in CreateRoom")
	for {
		roomID := room.GenerateID()
		// log.Println("roomID:",roomID)
		if !slices.Contains(rooms, roomID) {
			// log.Println(" in slices")
			room := &room.Room{
				ID:      roomID,
				Players: make(map[string]*room.Player, 4),
				Game:    &room.Big2Game{Deck: &big2_game.Deck{}},
			}
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
	slog.Info("in ListRooms")
	rooms := make([]string, 0, len(s.Rooms))
	for roomID := range s.Rooms {
		rooms = append(rooms, roomID)
	}

	return rooms
}

func (s *Server) AddPlayer(player *room.Player) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Players[player.ID] = player
	slog.Info("player add :", "player", player)
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

func (s *Server) JoinRoom(roomID string, player *room.Player) (bool, error) {
	slog.Info("in JoinRoom")
	room := s.GetRoom(roomID)
	if room == nil {
		return false, errors.New("roomID not exist")
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	if len(room.Players) >= 4 {
		return false, errors.New("room player already full")
	}

	room.Players[player.ID] = player
	slog.Info("room.Players : ", "players", room.Players)
	player.Room = room

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
