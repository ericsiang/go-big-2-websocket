package handle_errors

import "fmt"

type GameError struct {
	Code    int
	Message string
}

func (e *GameError) Error() string {
	return fmt.Sprintf("Game Error %d: %s", e.Code, e.Message)
}

var (
	ErrInvalidMove      = &GameError{Code: 1, Message: "Invalid move"}
	ErrNotPlayerTurn    = &GameError{Code: 2, Message: "Not player's turn"}
	ErrRoomFull         = &GameError{Code: 3, Message: "Room is full"}
	ErrPlayerNotFound   = &GameError{Code: 4, Message: "Player not found"}
	ErrInvalidGameState = &GameError{Code: 5, Message: "Invalid game state"}
)
