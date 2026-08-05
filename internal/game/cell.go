package game

// Mark is the note a player has put on a hidden cell. A cell carries at most
// one, which is why this is an enum rather than a pair of booleans.
type Mark uint8

const (
	// MarkNone is an unmarked cell.
	MarkNone Mark = iota
	// MarkFlag asserts a mine. Flagged cells cannot be revealed and count
	// towards the remaining-mine display and chording.
	MarkFlag
	// MarkQuestion records a suspicion. It is a note to the player and nothing
	// more: revealing, flood fill, and chording treat it as plain hidden.
	MarkQuestion
)

// Cell holds internal board state. Mine positions are never exposed to UI.
type Cell struct {
	HasMine  bool
	Adjacent uint8
	Revealed bool
	Mark     Mark
}

// CellState is the visible state of a cell.
type CellState int

const (
	CellHidden CellState = iota
	CellRevealed
	CellFlagged
	CellQuestioned
)

// CellView is the UI-facing projection of a cell (information hiding).
type CellView struct {
	State    CellState
	Adjacent uint8
	ShowMine bool // true when game is lost — reveal all mines
}
