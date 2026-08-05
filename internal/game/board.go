package game

import (
	"math/rand"
)

// Status is the overall game state.
type Status int

const (
	StatusPlaying Status = iota
	StatusWon
	StatusLost
)

// ActionResult describes the outcome of a board action.
type ActionResult struct {
	Ok      bool
	Changed []Coord
	Status  Status
}

// Board is the deep module encapsulating all game rules.
type Board struct {
	width       int
	height      int
	mineCount   int
	cells       []Cell
	status      Status
	difficulty  Difficulty
	seed        Seed
	minesPlaced bool
	firstReveal bool
	rng         *rand.Rand
}

// NewBoard creates an empty board; mines are placed on the first Reveal so
// that the opening move is always safe. The same difficulty and seed always
// produce the same layout.
func NewBoard(d Difficulty, seed Seed) *Board {
	return &Board{
		width:      d.Width,
		height:     d.Height,
		mineCount:  d.Mines,
		cells:      make([]Cell, d.Width*d.Height),
		status:     StatusPlaying,
		difficulty: d,
		seed:       seed,
		rng:        seed.rand(),
	}
}

func (b *Board) cell(c Coord) *Cell {
	return &b.cells[c.index(b.width)]
}

func (b *Board) noop() ActionResult {
	return ActionResult{Ok: false, Status: b.status}
}

func (b *Board) result(changed []Coord) ActionResult {
	return ActionResult{Ok: true, Changed: changed, Status: b.status}
}

// CanReveal reports whether Reveal would have an effect.
func (b *Board) CanReveal(c Coord) bool {
	if b.status != StatusPlaying || !c.InBounds(b.width, b.height) {
		return false
	}
	cell := b.cell(c)
	return !cell.Revealed && !cell.Flagged
}

// CanFlag reports whether ToggleFlag would have an effect.
func (b *Board) CanFlag(c Coord) bool {
	if b.status != StatusPlaying || !c.InBounds(b.width, b.height) {
		return false
	}
	cell := b.cell(c)
	return !cell.Revealed
}

// CanChord reports whether Chord would have an effect.
func (b *Board) CanChord(c Coord) bool {
	if b.status != StatusPlaying || !c.InBounds(b.width, b.height) {
		return false
	}
	cell := b.cell(c)
	if !cell.Revealed || cell.Adjacent == 0 {
		return false
	}
	var flags uint8
	for _, n := range c.Neighbors(b.width, b.height) {
		if b.cell(n).Flagged {
			flags++
		}
	}
	return flags == cell.Adjacent
}

// Reveal opens a cell and may trigger flood fill.
func (b *Board) Reveal(c Coord) ActionResult {
	if !b.CanReveal(c) {
		return b.noop()
	}

	if !b.minesPlaced {
		safe := safeZone(c, b.width, b.height)
		if err := placeMines(b, b.rng, map[Coord]struct{}{}); err != nil {
			return b.noop()
		}
		if err := relocateMinesFromSafeZone(b, safe, b.rng); err != nil {
			return b.noop()
		}
		b.minesPlaced = true
		b.firstReveal = true
	}

	changed := b.revealAt(c)
	if b.status == StatusLost {
		b.revealAllMines()
		return b.result(changed)
	}
	b.checkWin()
	return b.result(changed)
}

func (b *Board) revealAt(c Coord) []Coord {
	cell := b.cell(c)
	if cell.Revealed || cell.Flagged {
		return nil
	}

	var changed []Coord
	queue := []Coord{c}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		curCell := b.cell(cur)
		if curCell.Revealed || curCell.Flagged {
			continue
		}
		curCell.Revealed = true
		changed = append(changed, cur)

		if curCell.HasMine {
			b.status = StatusLost
			return changed
		}

		if curCell.Adjacent == 0 {
			for _, n := range cur.Neighbors(b.width, b.height) {
				nb := b.cell(n)
				if !nb.Revealed && !nb.Flagged {
					queue = append(queue, n)
				}
			}
		}
	}
	return changed
}

func (b *Board) revealAllMines() {
	for i := range b.cells {
		if b.cells[i].HasMine {
			b.cells[i].Revealed = true
		}
	}
}

func (b *Board) checkWin() {
	for i := range b.cells {
		if !b.cells[i].HasMine && !b.cells[i].Revealed {
			return
		}
	}
	b.status = StatusWon
}

// ToggleFlag toggles a flag on a hidden cell.
func (b *Board) ToggleFlag(c Coord) ActionResult {
	if !b.CanFlag(c) {
		return b.noop()
	}
	cell := b.cell(c)
	cell.Flagged = !cell.Flagged
	return b.result([]Coord{c})
}

// Chord reveals adjacent hidden cells when flag count matches the number.
func (b *Board) Chord(c Coord) ActionResult {
	if b.status != StatusPlaying || !c.InBounds(b.width, b.height) {
		return b.noop()
	}
	cell := b.cell(c)
	if !cell.Revealed || cell.Adjacent == 0 {
		return b.noop()
	}

	var flags uint8
	hidden := make([]Coord, 0)
	for _, n := range c.Neighbors(b.width, b.height) {
		nc := b.cell(n)
		if nc.Flagged {
			flags++
		} else if !nc.Revealed {
			hidden = append(hidden, n)
		}
	}

	if flags != cell.Adjacent {
		return b.noop()
	}

	var allChanged []Coord
	for _, h := range hidden {
		if b.status != StatusPlaying {
			break
		}
		changed := b.revealAt(h)
		allChanged = append(allChanged, changed...)
		if b.status == StatusLost {
			b.revealAllMines()
			return b.result(allChanged)
		}
	}
	b.checkWin()
	return b.result(allChanged)
}

// Width returns board width.
func (b *Board) Width() int { return b.width }

// Height returns board height.
func (b *Board) Height() int { return b.height }

// Status returns current game status.
func (b *Board) Status() Status { return b.status }

// MineCount returns total mines.
func (b *Board) MineCount() int { return b.mineCount }

// FlagCount returns number of flagged cells.
func (b *Board) FlagCount() int {
	n := 0
	for i := range b.cells {
		if b.cells[i].Flagged {
			n++
		}
	}
	return n
}

// RemainingMines returns mine count minus flags (may be negative).
func (b *Board) RemainingMines() int {
	return b.mineCount - b.FlagCount()
}

// CellView returns the UI-facing cell state.
func (b *Board) CellView(c Coord) CellView {
	if !c.InBounds(b.width, b.height) {
		return CellView{State: CellHidden}
	}
	cell := b.cell(c)
	showMine := b.status == StatusLost && cell.HasMine

	if cell.Flagged && b.status == StatusPlaying {
		return CellView{State: CellFlagged}
	}
	if cell.Revealed || showMine {
		return CellView{State: CellRevealed, Adjacent: cell.Adjacent, ShowMine: showMine}
	}
	return CellView{State: CellHidden}
}

// Difficulty returns the board difficulty.
func (b *Board) Difficulty() Difficulty { return b.difficulty }

// Seed returns the seed this board was generated from.
func (b *Board) Seed() Seed { return b.seed }

// ElapsedReady reports whether the timer should run (first reveal done).
func (b *Board) ElapsedReady() bool { return b.firstReveal }
