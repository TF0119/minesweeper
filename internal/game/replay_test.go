package game

import "testing"

func TestReplayApplyReproducesMoves(t *testing.T) {
	d := PresetDifficulty(Beginner)
	seed := Seed(2)
	first := Coord{X: 4, Y: 4}

	// Play a short game by hand.
	live := NewBoard(d, seed)
	live.Reveal(first)
	live.CycleMark(Coord{1, 1}, true)

	replay := Replay{
		Seed:       seed,
		Difficulty: d,
		Moves: []Move{
			{Kind: MoveReveal, Coord: first},
			{Kind: MoveMark, Coord: Coord{1, 1}},
		},
	}

	watch := NewBoard(d, seed)
	replay.Apply(watch, len(replay.Moves))

	if got := watch.CellView(first).State; got != CellRevealed {
		t.Errorf("replayed reveal: state = %v", got)
	}
	if got := watch.CellView(Coord{1, 1}).State; got != CellFlagged {
		t.Errorf("replayed mark: state = %v", got)
	}
}
