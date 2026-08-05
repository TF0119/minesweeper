package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/takeru0119/minesweeper/internal/game"
)

const highScoreVersion = 1

// Record is a single high-score entry.
type Record struct {
	Seconds int       `json:"seconds"`
	Date    time.Time `json:"date"`
	Width   int       `json:"width"`
	Height  int       `json:"height"`
	Mines   int       `json:"mines"`
}

// HighScores holds best times keyed by difficulty key.
type HighScores struct {
	Version int               `json:"version"`
	Records map[string]Record `json:"records"`
}

// DefaultHighScores returns an empty score table.
func DefaultHighScores() HighScores {
	return HighScores{
		Version: highScoreVersion,
		Records: make(map[string]Record),
	}
}

// LoadHighScores reads high scores from disk.
func LoadHighScores() (HighScores, error) {
	path, err := HighScorePath()
	if err != nil {
		return DefaultHighScores(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultHighScores(), nil
		}
		return DefaultHighScores(), err
	}
	var h HighScores
	if err := json.Unmarshal(data, &h); err != nil {
		return DefaultHighScores(), err
	}
	if h.Records == nil {
		h.Records = make(map[string]Record)
	}
	return h, nil
}

// SaveHighScores writes high scores atomically.
func SaveHighScores(h HighScores) error {
	path, err := HighScorePath()
	if err != nil {
		return err
	}
	h.Version = highScoreVersion
	return writeJSONAtomic(path, h)
}

// TryUpdate records a new best time; returns true if updated.
func (h *HighScores) TryUpdate(key string, seconds int, d game.Difficulty) bool {
	if h.Records == nil {
		h.Records = make(map[string]Record)
	}
	existing, ok := h.Records[key]
	if ok && existing.Seconds <= seconds {
		return false
	}
	h.Records[key] = Record{
		Seconds: seconds,
		Date:    time.Now().UTC(),
		Width:   d.Width,
		Height:  d.Height,
		Mines:   d.Mines,
	}
	return true
}

// Best returns the best time for a key, or -1 if none.
func (h HighScores) Best(key string) int {
	if r, ok := h.Records[key]; ok {
		return r.Seconds
	}
	return -1
}
