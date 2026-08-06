package game

import "time"

// MoveKind is a single player action in a replay.
type MoveKind int

const (
	MoveReveal MoveKind = iota
	MoveMark
	MoveChord
)

// Move is one recorded action in order.
type Move struct {
	Kind MoveKind `json:"kind"`
	Coord
	// TargetMark is the mark left on the cell after a MoveMark. Older replays
	// omit this and fall back to a single flag cycle.
	TargetMark *Mark `json:"target_mark,omitempty"`
}

// Replay is a finished game that can be watched again. The seed pins the mine
// layout; the moves pin what the player did. NoGuess records which generator
// built the board: classic and no-guess draw differently from the same seed.
// Older JSON omits the field and loads as classic.
type Replay struct {
	ID         string     `json:"id"`
	Seed       Seed       `json:"seed"`
	Difficulty Difficulty `json:"difficulty"`
	NoGuess    bool       `json:"no_guess"`
	Moves      []Move     `json:"moves"`
	Won        bool       `json:"won"`
	Seconds    int        `json:"seconds"`
	PlayedAt   time.Time  `json:"played_at"`
}

// Apply replays moves onto a fresh board in order, up to but not including
// index n. The board must already have been generated from the replay's seed
// and had its opening move applied the same way the original game did.
func (r Replay) Apply(b *Board, n int) {
	if n > len(r.Moves) {
		n = len(r.Moves)
	}
	for i := 0; i < n; i++ {
		m := r.Moves[i]
		switch m.Kind {
		case MoveReveal:
			b.Reveal(m.Coord)
		case MoveMark:
			if m.TargetMark != nil {
				b.SetMark(m.Coord, *m.TargetMark)
			} else {
				b.CycleMark(m.Coord, true)
			}
		case MoveChord:
			b.Chord(m.Coord)
		}
	}
}
