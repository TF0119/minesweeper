package game

// Cell holds internal board state. Mine positions are never exposed to UI.
type Cell struct {
	HasMine  bool
	Adjacent uint8
	Revealed bool
	Flagged  bool
}

// CellState is the visible state of a cell.
type CellState int

const (
	CellHidden CellState = iota
	CellRevealed
	CellFlagged
)

// CellView is the UI-facing projection of a cell (information hiding).
type CellView struct {
	State    CellState
	Adjacent uint8
	ShowMine bool // true when game is lost — reveal all mines
}
