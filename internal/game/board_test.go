package game

import (
	"testing"
)

func TestFirstClickNeverMine(t *testing.T) {
	d := PresetDifficulty(Beginner)
	for seed := uint32(0); seed < 100; seed++ {
		for y := 0; y < d.Height; y++ {
			for x := 0; x < d.Width; x++ {
				b := NewBoard(d, Seed(seed*1000+uint32(y*d.Width+x)))
				first := Coord{X: x, Y: y}
				res := b.Reveal(first)
				if !res.Ok {
					t.Fatalf("seed %d coord %+v: reveal failed", seed, first)
				}
				if b.Status() == StatusLost {
					t.Fatalf("seed %d coord %+v: first click hit mine", seed, first)
				}
				for _, n := range first.Neighbors(d.Width, d.Height) {
					if b.cell(n).HasMine {
						t.Fatalf("seed %d coord %+v: neighbor %+v has mine", seed, first, n)
					}
				}
			}
		}
	}
}

func TestFloodFill(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 5, Height: 5, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	// center empty region
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			b.setCell(Coord{x, y}, false, 0, false, false)
		}
	}
	b.setCell(Coord{4, 4}, true, 0, false, false)
	computeAdjacent(b)

	res := b.Reveal(Coord{0, 0})
	if !res.Ok {
		t.Fatal("reveal failed")
	}
	if len(res.Changed) != 24 {
		t.Errorf("flood fill changed %d cells, want 24", len(res.Changed))
	}
}

func TestRevealRevealedNoOp(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()
	b.setCell(Coord{0, 0}, false, 1, true, false)
	res := b.Reveal(Coord{0, 0})
	if res.Ok {
		t.Error("expected no-op on revealed cell")
	}
}

func TestFlagToggle(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	res := b.CycleMark(Coord{1, 1}, false)
	if !res.Ok || b.CellView(Coord{1, 1}).State != CellFlagged {
		t.Error("flag toggle failed")
	}
	res = b.CycleMark(Coord{1, 1}, false)
	if !res.Ok || b.CellView(Coord{1, 1}).State != CellHidden {
		t.Error("flag untoggle failed")
	}
}

func TestFlagOnRevealed(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()
	b.setCell(Coord{0, 0}, false, 1, true, false)
	if b.CanMark(Coord{0, 0}) {
		t.Error("CanMark should be false on revealed cell")
	}
}

func TestChordSuccess(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 3, Height: 3, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	// 1 at center, mines at corners flagged
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.setCell(Coord{x, y}, false, 0, false, false)
		}
	}
	b.setCell(Coord{1, 1}, false, 2, true, false)
	b.setCell(Coord{0, 0}, true, 0, false, true)
	b.setCell(Coord{2, 0}, true, 0, false, true)
	b.setCell(Coord{0, 2}, false, 0, false, false)
	b.setCell(Coord{2, 2}, false, 0, false, false)

	res := b.Chord(Coord{1, 1})
	if !res.Ok {
		t.Fatal("chord failed")
	}
	if b.CellView(Coord{0, 2}).State != CellRevealed {
		t.Error("expected 0,2 revealed")
	}
}

func TestChordWrongFlag(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 3, Height: 3, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.setCell(Coord{x, y}, false, 0, false, false)
		}
	}
	b.setCell(Coord{1, 1}, false, 1, true, false)
	b.setCell(Coord{0, 0}, true, 0, false, false)
	// no flags placed — count mismatch

	if b.CanChord(Coord{1, 1}) {
		t.Error("CanChord should be false when flag count mismatches")
	}
	res := b.Chord(Coord{1, 1})
	if res.Ok {
		t.Error("chord should fail when flag count wrong")
	}
}

func TestChordHitsMine(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 3, Height: 3, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.setCell(Coord{x, y}, false, 0, false, false)
		}
	}
	b.setCell(Coord{1, 1}, false, 1, true, false)
	b.setCell(Coord{0, 0}, true, 0, false, true)
	b.setCell(Coord{2, 0}, true, 0, false, false) // unflagged mine

	res := b.Chord(Coord{1, 1})
	if !res.Ok || b.Status() != StatusLost {
		t.Error("chord should lose when revealing unflagged mine")
	}
}

func TestWinCondition(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 2, Height: 2, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()

	b.setCell(Coord{0, 0}, false, 1, false, false)
	b.setCell(Coord{1, 0}, false, 1, false, false)
	b.setCell(Coord{0, 1}, false, 1, false, false)
	b.setCell(Coord{1, 1}, true, 0, false, false)

	b.Reveal(Coord{0, 0})
	b.Reveal(Coord{1, 0})
	res := b.Reveal(Coord{0, 1})
	if b.Status() != StatusWon {
		t.Errorf("expected win, got status %v, res=%+v", b.Status(), res)
	}
}

func TestLossOnMine(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 2, Height: 2, Mines: 1}
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()
	b.setCell(Coord{1, 1}, true, 0, false, false)

	res := b.Reveal(Coord{1, 1})
	if !res.Ok || b.Status() != StatusLost {
		t.Error("expected loss on mine")
	}
}

func TestNoOpAfterGameOver(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()
	b.setStatus(StatusWon)
	res := b.Reveal(Coord{0, 0})
	if res.Ok {
		t.Error("expected no-op after win")
	}
}

func TestCoordNeighbors(t *testing.T) {
	c := Coord{0, 0}
	ns := c.Neighbors(3, 3)
	if len(ns) != 3 {
		t.Errorf("corner has 3 neighbors, got %d", len(ns))
	}
}

func TestRemainingMines(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, Seed(1))
	b.markMinesPlaced()
	b.CycleMark(Coord{0, 0}, false)
	b.CycleMark(Coord{1, 0}, false)
	if b.RemainingMines() != 8 {
		t.Errorf("remaining mines = %d, want 8", b.RemainingMines())
	}
}

func TestCycleMark(t *testing.T) {
	tests := []struct {
		name      string
		questions bool
		want      []CellState
	}{
		{"questions on", true, []CellState{CellFlagged, CellQuestioned, CellHidden}},
		{"questions off", false, []CellState{CellFlagged, CellHidden, CellFlagged}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBoard(PresetDifficulty(Beginner), Seed(1))
			c := Coord{X: 1, Y: 1}
			for i, want := range tt.want {
				if res := b.CycleMark(c, tt.questions); !res.Ok {
					t.Fatalf("step %d: CycleMark reported no effect", i)
				}
				if got := b.CellView(c).State; got != want {
					t.Errorf("step %d: state = %v, want %v", i, got, want)
				}
			}
		})
	}
}

// A question mark is a note, not a shield: unlike a flag it must not change
// what reveal, flood fill, or chord do.
func TestQuestionMarkDoesNotProtectTheCell(t *testing.T) {
	b := NewBoard(Difficulty{Preset: Custom, Width: 3, Height: 1, Mines: 1}, Seed(1))
	b.setCell(Coord{0, 0}, false, 0, false, false)
	b.setCell(Coord{1, 0}, false, 1, false, false)
	b.setCell(Coord{2, 0}, true, 0, false, false)

	q := Coord{1, 0}
	b.CycleMark(q, true)
	b.CycleMark(q, true)
	if got := b.CellView(q).State; got != CellQuestioned {
		t.Fatalf("setup: state = %v, want CellQuestioned", got)
	}

	if !b.CanReveal(q) {
		t.Error("a questioned cell should still be revealable")
	}
	if got := b.RemainingMines(); got != 1 {
		t.Errorf("RemainingMines = %d, want 1: a ? is not a flag", got)
	}
	if res := b.Reveal(q); !res.Ok || b.CellView(q).State != CellRevealed {
		t.Errorf("revealing a questioned cell failed: %+v", res)
	}
}

func TestFloodFillClearsQuestionMarksButStopsAtFlags(t *testing.T) {
	b := NewBoard(Difficulty{Preset: Custom, Width: 3, Height: 1, Mines: 0}, Seed(1))
	for x := 0; x < 3; x++ {
		b.setCell(Coord{x, 0}, false, 0, false, false)
	}
	questioned, flagged := Coord{1, 0}, Coord{2, 0}
	b.CycleMark(questioned, true)
	b.CycleMark(questioned, true)
	b.CycleMark(flagged, true)

	b.Reveal(Coord{0, 0})

	if got := b.CellView(questioned).State; got != CellRevealed {
		t.Errorf("questioned cell state = %v, want CellRevealed", got)
	}
	if got := b.CellView(flagged).State; got != CellFlagged {
		t.Errorf("flagged cell state = %v, want CellFlagged: flags stop flood fill", got)
	}
}
