package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
)

func TestFitShowsWholeBoardWhenItFits(t *testing.T) {
	v := fit(viewport{}, 200, 60, 30, 16)
	if v.cols != 30 || v.rows != 16 {
		t.Errorf("cols,rows = %d,%d want 30,16", v.cols, v.rows)
	}
	if v.scrolls(30, 16) {
		t.Error("should not scroll when the board fits")
	}
}

func TestFitClipsToTerminal(t *testing.T) {
	// 40 columns / 3 per cell = 13 cells; 12 rows - 3 chrome = 9 rows.
	v := fit(viewport{}, 40, 12, 30, 16)
	if v.cols != 13 || v.rows != 9 {
		t.Errorf("cols,rows = %d,%d want 13,9", v.cols, v.rows)
	}
	if !v.scrolls(30, 16) {
		t.Error("should scroll when the board is clipped")
	}
}

func TestFitUnknownTerminalSizeShowsWholeBoard(t *testing.T) {
	v := fit(viewport{}, 0, 0, 30, 16)
	if v.cols != 30 || v.rows != 16 {
		t.Errorf("cols,rows = %d,%d want 30,16", v.cols, v.rows)
	}
}

func TestFitKeepsOffsetInRange(t *testing.T) {
	v := viewport{offsetX: 25, offsetY: 12}
	v = fit(v, 40, 12, 30, 16)
	if v.offsetX+v.cols > 30 || v.offsetY+v.rows > 16 {
		t.Errorf("offset out of range: %+v", v)
	}
}

func TestFollowScrollsMinimally(t *testing.T) {
	v := fit(viewport{}, 40, 12, 30, 16) // cols=13 rows=9

	v = v.follow(game.Coord{X: 20, Y: 0}, 30, 16)
	if v.offsetX != 20-13+1 {
		t.Errorf("offsetX = %d, want %d", v.offsetX, 20-13+1)
	}

	// Moving back left scrolls just enough to expose the cursor.
	v = v.follow(game.Coord{X: 2, Y: 0}, 30, 16)
	if v.offsetX != 2 {
		t.Errorf("offsetX = %d, want 2", v.offsetX)
	}

	// A cursor already in view leaves the offset alone.
	before := v.offsetX
	v = v.follow(game.Coord{X: 5, Y: 0}, 30, 16)
	if v.offsetX != before {
		t.Errorf("offsetX moved to %d for a visible cursor", v.offsetX)
	}
}

func TestToBoardAppliesOffset(t *testing.T) {
	v := fit(viewport{}, 40, 12, 30, 16)
	v = v.follow(game.Coord{X: 20, Y: 12}, 30, 16)

	got, ok := v.toBoard(0, 1, 30, 16)
	if !ok {
		t.Fatal("top-left board cell should map")
	}
	if got != (game.Coord{X: v.offsetX, Y: v.offsetY}) {
		t.Errorf("got %+v, want %+v", got, game.Coord{X: v.offsetX, Y: v.offsetY})
	}

	if _, ok := v.toBoard(0, 0, 30, 16); ok {
		t.Error("HUD row should not map to a cell")
	}
	if _, ok := v.toBoard(999, 1, 30, 16); ok {
		t.Error("column beyond the window should not map")
	}
}
