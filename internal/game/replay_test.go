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
	flag := MarkFlag

	replay := Replay{
		Seed:       seed,
		Difficulty: d,
		Moves: []Move{
			{Kind: MoveReveal, Coord: first},
			{Kind: MoveMark, Coord: Coord{1, 1}, TargetMark: &flag},
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

func TestReplayApplyClearsMarkWithoutQuestionMarks(t *testing.T) {
	d := PresetDifficulty(Beginner)
	seed := Seed(2)
	first := Coord{X: 4, Y: 4}
	markCell := Coord{1, 1}

	live := NewBoard(d, seed)
	live.Reveal(first)
	live.CycleMark(markCell, false) // flag
	live.CycleMark(markCell, false) // clear
	if got := live.CellView(markCell).State; got != CellHidden {
		t.Fatalf("live cell = %v, want hidden", got)
	}

	none := MarkNone
	replay := Replay{
		Seed:       seed,
		Difficulty: d,
		Moves: []Move{
			{Kind: MoveReveal, Coord: first},
			{Kind: MoveMark, Coord: markCell, TargetMark: &none},
		},
	}

	watch := NewBoard(d, seed)
	replay.Apply(watch, len(replay.Moves))
	if got := watch.CellView(markCell).State; got != CellHidden {
		t.Errorf("replayed clear: state = %v, want hidden", got)
	}
}
