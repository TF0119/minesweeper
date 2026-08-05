package storage

import (
	"encoding/json"
	"os"
)

const statsVersion = 1

// Tally is the running record for one difficulty.
//
// Only counts are stored. Win rate and average time are derived on demand,
// because a stored total that disagrees with the counts it summarises is a bug
// waiting to happen.
type Tally struct {
	Played        int `json:"played"`
	Won           int `json:"won"`
	CurrentStreak int `json:"current_streak"`
	BestStreak    int `json:"best_streak"`
	WonSeconds    int `json:"won_seconds"`
}

// WinRate is the fraction of finished games won, in the range 0 to 1. An
// unplayed difficulty rates zero rather than being undefined, so callers never
// have to special-case it.
func (t Tally) WinRate() float64 {
	if t.Played == 0 {
		return 0
	}
	return float64(t.Won) / float64(t.Played)
}

// AverageWinSeconds is the mean time of won games, or -1 when there are none.
// Losses are excluded: a game abandoned to a mine says nothing about pace.
func (t Tally) AverageWinSeconds() int {
	if t.Won == 0 {
		return -1
	}
	return t.WonSeconds / t.Won
}

// Stats holds per-difficulty tallies.
type Stats struct {
	Version int              `json:"version"`
	Tallies map[string]Tally `json:"tallies"`
}

// DefaultStats returns an empty record.
func DefaultStats() Stats {
	return Stats{Version: statsVersion, Tallies: make(map[string]Tally)}
}

// LoadStats reads stats from disk, or returns an empty record when there are
// none yet.
func LoadStats() (Stats, error) {
	path, err := StatsPath()
	if err != nil {
		return DefaultStats(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultStats(), nil
		}
		return DefaultStats(), err
	}
	s := DefaultStats()
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultStats(), err
	}
	if s.Tallies == nil {
		s.Tallies = make(map[string]Tally)
	}
	return s, nil
}

// SaveStats writes stats atomically.
func SaveStats(s Stats) error {
	path, err := StatsPath()
	if err != nil {
		return err
	}
	s.Version = statsVersion
	return writeJSONAtomic(path, s)
}

// For returns the tally for a difficulty, zero-valued when it has never been
// played.
func (s Stats) For(key string) Tally { return s.Tallies[key] }

// RecordWin adds a won game.
func (s *Stats) RecordWin(key string, seconds int) {
	t := s.begin(key)
	t.Won++
	t.WonSeconds += seconds
	t.CurrentStreak++
	if t.CurrentStreak > t.BestStreak {
		t.BestStreak = t.CurrentStreak
	}
	s.Tallies[key] = t
}

// RecordLoss adds a lost game and ends the streak.
func (s *Stats) RecordLoss(key string) {
	t := s.begin(key)
	t.CurrentStreak = 0
	s.Tallies[key] = t
}

// begin counts a finished game. Only games that reach a result are counted, so
// starting a fresh board mid-game never shows up as a loss.
func (s *Stats) begin(key string) Tally {
	if s.Tallies == nil {
		s.Tallies = make(map[string]Tally)
	}
	t := s.Tallies[key]
	t.Played++
	return t
}
