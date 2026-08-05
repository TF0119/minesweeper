package game

// Test-only accessors. Keeping them here rather than in board.go preserves the
// narrow public interface of Board.

func (b *Board) setStatus(s Status) { b.status = s }

func (b *Board) setCell(c Coord, mine bool, adjacent uint8, revealed, flagged bool) {
	cell := b.cell(c)
	cell.HasMine = mine
	cell.Adjacent = adjacent
	cell.Revealed = revealed
	cell.Mark = MarkNone
	if flagged {
		cell.Mark = MarkFlag
	}
	b.minesPlaced = true
}

func (b *Board) markMinesPlaced() { b.minesPlaced = true }

func (b *Board) hasMine(c Coord) bool { return b.cell(c).HasMine }

// resetReveals returns the board to its just-generated state with the mine
// layout intact, so a solver run can be replayed against it.
func (b *Board) resetReveals() {
	for i := range b.cells {
		b.cells[i].Revealed = false
		b.cells[i].Mark = MarkNone
	}
	b.status = StatusPlaying
}
