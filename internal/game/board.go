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

	noGuess         bool // ask the generator for a board that needs no guessing
	noGuessVerified bool // and whether it actually found one
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

// NewNoGuessBoard is NewBoard for a board that can be cleared by deduction
// alone, so the player is never asked to flip a coin. Laying one out means
// searching, which costs time on the opening move and is not always possible;
// when the search comes up empty the board falls back to an ordinary layout
// and NoGuess reports false.
func NewNoGuessBoard(d Difficulty, seed Seed) *Board {
	b := NewBoard(d, seed)
	b.noGuess = true
	return b
}

// NoGuess reports whether this board is known to be solvable without guessing.
// It is only meaningful once mines are placed, which happens on the first
// reveal.
func (b *Board) NoGuess() bool { return b.noGuessVerified }

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
	return !cell.Revealed && cell.Mark != MarkFlag
}

// CanMark reports whether CycleMark would have an effect.
func (b *Board) CanMark(c Coord) bool {
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
		if b.cell(n).Mark == MarkFlag {
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
		if err := b.placeMinesAround(c); err != nil {
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

// placeMinesAround lays out mines with the 3x3 around the opening move kept
// clear, which is what makes the first click safe and always a cascade.
func (b *Board) placeMinesAround(first Coord) error {
	safe := safeZone(first, b.width, b.height)
	if b.noGuess {
		verified, err := placeSolvableMines(b, safe, first, b.rng)
		b.noGuessVerified = verified
		return err
	}

	// The classic path places mines everywhere and then moves the ones that
	// landed in the safe zone. It draws from the generator in a specific order,
	// so it is kept as-is: changing it would hand every existing seed a
	// different board.
	if err := placeMines(b, b.rng, map[Coord]struct{}{}); err != nil {
		return err
	}
	return relocateMinesFromSafeZone(b, safe, b.rng)
}

func (b *Board) revealAt(c Coord) []Coord {
	cell := b.cell(c)
	if cell.Revealed || cell.Mark == MarkFlag {
		return nil
	}

	var changed []Coord
	queue := []Coord{c}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		curCell := b.cell(cur)
		if curCell.Revealed || curCell.Mark == MarkFlag {
			continue
		}
		curCell.Revealed = true
		// A mark is a note about a hidden cell; opening it answers the question.
		curCell.Mark = MarkNone
		changed = append(changed, cur)

		if curCell.HasMine {
			b.status = StatusLost
			return changed
		}

		if curCell.Adjacent == 0 {
			for _, n := range cur.Neighbors(b.width, b.height) {
				nb := b.cell(n)
				if !nb.Revealed && nb.Mark != MarkFlag {
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

// CycleMark advances the mark on a hidden cell. With questions enabled the
// cycle is none, flag, question, none; otherwise it is a plain flag toggle.
// Keeping the choice in the caller lets the preference live with the UI while
// the board owns what each mark means.
func (b *Board) CycleMark(c Coord, questions bool) ActionResult {
	if !b.CanMark(c) {
		return b.noop()
	}
	cell := b.cell(c)
	cell.Mark = nextMark(cell.Mark, questions)
	return b.result([]Coord{c})
}

func nextMark(m Mark, questions bool) Mark {
	switch m {
	case MarkNone:
		return MarkFlag
	case MarkFlag:
		if questions {
			return MarkQuestion
		}
		return MarkNone
	default:
		return MarkNone
	}
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
		if nc.Mark == MarkFlag {
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
		if b.cells[i].Mark == MarkFlag {
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

	if b.status == StatusPlaying {
		switch cell.Mark {
		case MarkFlag:
			return CellView{State: CellFlagged}
		case MarkQuestion:
			return CellView{State: CellQuestioned}
		}
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
