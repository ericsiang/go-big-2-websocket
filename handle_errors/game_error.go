package handle_errors

import "fmt"

type GameError struct {
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *GameError) Error() string {
	return fmt.Sprintf("GameError_%d: %s", e.Code, e.Message)
}

var (
	ErrInvalidMove           = &GameError{Type: "Game", Code: 1, Message: "Invalid move"}
	ErrNotPlayerTurn         = &GameError{Type: "Game", Code: 2, Message: "Not player's turn"}
	ErrRoomNotFound          = &GameError{Type: "Game", Code: 3, Message: "Room not found"}
	ErrRoomFull              = &GameError{Type: "Game", Code: 4, Message: "Room is full"}
	ErrPlayerNotFound        = &GameError{Type: "Game", Code: 5, Message: "Player not found"}
	ErrInvalidGameState      = &GameError{Type: "Game", Code: 6, Message: "Invalid game state"}
	ErrFirstCardNotAllow     = &GameError{Type: "Game", Code: 7, Message: "First Card must have block three"}
	ErrPlayerAlreadyHaveRoom = &GameError{Type: "Game", Code: 8, Message: "Player already have room"}
)
