package game

import (
	"math/rand"
	"testing"
)

func TestFirstClickNeverMine(t *testing.T) {
	d := PresetDifficulty(Beginner)
	for seed := int64(0); seed < 100; seed++ {
		for y := 0; y < d.Height; y++ {
			for x := 0; x < d.Width; x++ {
				b := NewBoard(d, rand.New(rand.NewSource(seed*1000+int64(y*d.Width+x))))
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
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	// center empty region
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			b.SetCellForTest(Coord{x, y}, false, 0, false, false)
		}
	}
	b.SetCellForTest(Coord{4, 4}, true, 0, false, false)
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
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()
	b.SetCellForTest(Coord{0, 0}, false, 1, true, false)
	res := b.Reveal(Coord{0, 0})
	if res.Ok {
		t.Error("expected no-op on revealed cell")
	}
}

func TestFlagToggle(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	res := b.ToggleFlag(Coord{1, 1})
	if !res.Ok || b.CellView(Coord{1, 1}).State != CellFlagged {
		t.Error("flag toggle failed")
	}
	res = b.ToggleFlag(Coord{1, 1})
	if !res.Ok || b.CellView(Coord{1, 1}).State != CellHidden {
		t.Error("flag untoggle failed")
	}
}

func TestFlagOnRevealed(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()
	b.SetCellForTest(Coord{0, 0}, false, 1, true, false)
	if b.CanFlag(Coord{0, 0}) {
		t.Error("CanFlag should be false on revealed cell")
	}
}

func TestChordSuccess(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 3, Height: 3, Mines: 1}
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	// 1 at center, mines at corners flagged
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.SetCellForTest(Coord{x, y}, false, 0, false, false)
		}
	}
	b.SetCellForTest(Coord{1, 1}, false, 2, true, false)
	b.SetCellForTest(Coord{0, 0}, true, 0, false, true)
	b.SetCellForTest(Coord{2, 0}, true, 0, false, true)
	b.SetCellForTest(Coord{0, 2}, false, 0, false, false)
	b.SetCellForTest(Coord{2, 2}, false, 0, false, false)

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
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.SetCellForTest(Coord{x, y}, false, 0, false, false)
		}
	}
	b.SetCellForTest(Coord{1, 1}, false, 1, true, false)
	b.SetCellForTest(Coord{0, 0}, true, 0, false, false)
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
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			b.SetCellForTest(Coord{x, y}, false, 0, false, false)
		}
	}
	b.SetCellForTest(Coord{1, 1}, false, 1, true, false)
	b.SetCellForTest(Coord{0, 0}, true, 0, false, true)
	b.SetCellForTest(Coord{2, 0}, true, 0, false, false) // unflagged mine

	res := b.Chord(Coord{1, 1})
	if !res.Ok || b.Status() != StatusLost {
		t.Error("chord should lose when revealing unflagged mine")
	}
}

func TestWinCondition(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 2, Height: 2, Mines: 1}
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()

	b.SetCellForTest(Coord{0, 0}, false, 1, false, false)
	b.SetCellForTest(Coord{1, 0}, false, 1, false, false)
	b.SetCellForTest(Coord{0, 1}, false, 1, false, false)
	b.SetCellForTest(Coord{1, 1}, true, 0, false, false)

	b.Reveal(Coord{0, 0})
	b.Reveal(Coord{1, 0})
	res := b.Reveal(Coord{0, 1})
	if b.Status() != StatusWon {
		t.Errorf("expected win, got status %v, res=%+v", b.Status(), res)
	}
}

func TestLossOnMine(t *testing.T) {
	d := Difficulty{Preset: Custom, Width: 2, Height: 2, Mines: 1}
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()
	b.SetCellForTest(Coord{1, 1}, true, 0, false, false)

	res := b.Reveal(Coord{1, 1})
	if !res.Ok || b.Status() != StatusLost {
		t.Error("expected loss on mine")
	}
}

func TestNoOpAfterGameOver(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()
	b.SetStatusForTest(StatusWon)
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
	b := NewBoard(d, rand.New(rand.NewSource(1)))
	b.MarkMinesPlacedForTest()
	b.ToggleFlag(Coord{0, 0})
	b.ToggleFlag(Coord{1, 0})
	if b.RemainingMines() != 8 {
		t.Errorf("remaining mines = %d, want 8", b.RemainingMines())
	}
}
