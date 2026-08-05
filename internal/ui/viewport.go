package ui

import "github.com/TF0119/minesweeper/internal/game"

// Rows of chrome drawn around the board: the HUD above, and the scroll
// indicator plus status bar below.
const chromeRows = 3

// viewport is the window of board cells currently drawn. Expert boards are
// wider than many terminals, so the board is scrolled rather than truncated.
// A zero viewport means "not sized yet"; callers use fit before rendering.
type viewport struct {
	offsetX, offsetY int
	cols, rows       int
}

// fit sizes the window to the terminal while keeping it within the board.
// Non-positive terminal dimensions mean the size is not known yet, in which
// case the whole board is shown.
func fit(v viewport, termW, termH, boardW, boardH int) viewport {
	v.cols = boardW
	v.rows = boardH

	if termW > 0 {
		if c := termW / cellWidth; c < boardW {
			v.cols = max(c, 1)
		}
	}
	if termH > 0 {
		if r := termH - chromeRows; r < boardH {
			v.rows = max(r, 1)
		}
	}

	v.offsetX = clamp(v.offsetX, 0, boardW-v.cols)
	v.offsetY = clamp(v.offsetY, 0, boardH-v.rows)
	return v
}

// follow scrolls by the smallest amount that brings c into view.
func (v viewport) follow(c game.Coord, boardW, boardH int) viewport {
	v.offsetX = scrollAxis(v.offsetX, v.cols, c.X, boardW)
	v.offsetY = scrollAxis(v.offsetY, v.rows, c.Y, boardH)
	return v
}

func scrollAxis(offset, span, target, total int) int {
	if span >= total {
		return 0
	}
	if target < offset {
		offset = target
	} else if target >= offset+span {
		offset = target - span + 1
	}
	return clamp(offset, 0, total-span)
}

// scrolls reports whether any part of the board is off screen.
func (v viewport) scrolls(boardW, boardH int) bool {
	return v.cols < boardW || v.rows < boardH
}

// toBoard maps a terminal cell to a board coordinate, reporting false when the
// position falls outside the drawn board.
func (v viewport) toBoard(termX, termY, boardW, boardH int) (game.Coord, bool) {
	row := termY - 1 // row 0 is the HUD
	if row < 0 || row >= v.rows {
		return game.Coord{}, false
	}
	col := termX / cellWidth
	if col < 0 || col >= v.cols {
		return game.Coord{}, false
	}
	c := game.Coord{X: v.offsetX + col, Y: v.offsetY + row}
	if !c.InBounds(boardW, boardH) {
		return game.Coord{}, false
	}
	return c, true
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
