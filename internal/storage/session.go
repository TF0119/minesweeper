package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
)

const sessionVersion = 1

// Session is a game that was still in progress when the player quit. The board
// itself is not stored: the seed and the moves rebuild it exactly, so there is
// one description of what happened rather than two that can disagree.
type Session struct {
	Version    int             `json:"version"`
	Seed       game.Seed       `json:"seed"`
	Difficulty game.Difficulty `json:"difficulty"`
	NoGuess    bool            `json:"no_guess"`
	Moves      []game.Move     `json:"moves"`
	Seconds    int             `json:"seconds"`
	Cursor     game.Coord      `json:"cursor"`
	SavedAt    time.Time       `json:"saved_at"`
}

// SessionPath returns the path to session.json.
func SessionPath() (string, error) {
	return dataFile("session.json")
}

// SaveSession writes the unfinished game atomically.
func SaveSession(s Session) error {
	path, err := SessionPath()
	if err != nil {
		return err
	}
	s.Version = sessionVersion
	if s.SavedAt.IsZero() {
		s.SavedAt = time.Now()
	}
	s.SavedAt = s.SavedAt.UTC()
	return writeJSONAtomic(path, s)
}

// LoadSession reads the unfinished game. The bool reports whether one was
// found; a missing file is the normal case and not an error.
func LoadSession() (Session, bool, error) {
	path, err := SessionPath()
	if err != nil {
		return Session{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, false, nil
		}
		return Session{}, false, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return Session{}, false, err
	}
	// A game with no moves has nothing to restore, and a board that cannot be
	// built is worse than starting fresh.
	if len(s.Moves) == 0 || s.Difficulty.Validate() != nil {
		return Session{}, false, nil
	}
	return s, true, nil
}

// ClearSession removes any saved game. Removing one that is not there is not
// an error: callers clear unconditionally rather than checking first.
func ClearSession() error {
	path, err := SessionPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
