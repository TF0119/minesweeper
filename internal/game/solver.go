package game

import "sort"

// This file answers one question: can this board be cleared by reasoning
// alone? It is the machinery behind no-guess generation, and it deliberately
// reasons the way a player does. A board it cannot finish is a board where a
// human would have to guess, even if a perfect prover could crack it.

// constraint says that exactly need of cells hold mines. Every revealed number
// on the frontier produces one.
type constraint struct {
	cells []Coord
	need  int
}

// solvable reports whether the board can be cleared from first without ever
// guessing. It plays the game out on a copy, so the caller's board is
// untouched.
func solvable(src *Board, first Coord) bool {
	b := src.cloneWithMines()
	if res := b.Reveal(first); !res.Ok {
		return false
	}
	for b.status == StatusPlaying && b.deduceOnce() {
	}
	return b.status == StatusWon
}

// cloneWithMines copies the board as an already-generated position, so the
// clone reveals cells instead of laying out mines again.
func (b *Board) cloneWithMines() *Board {
	clone := *b
	clone.cells = make([]Cell, len(b.cells))
	copy(clone.cells, b.cells)
	clone.minesPlaced = true
	return &clone
}

// deduceOnce applies every rule once and acts on what it learned. It reports
// whether anything was learned, which is how the caller knows to stop.
func (b *Board) deduceOnce() bool {
	hidden := b.hiddenCells()
	safe, mines := deduce(b.frontierConstraints(), hidden, b.RemainingMines())
	if len(safe) == 0 && len(mines) == 0 {
		return false
	}

	for _, c := range mines {
		b.cell(c).Mark = MarkFlag
	}
	for _, c := range safe {
		b.Reveal(c)
		if b.status != StatusPlaying {
			break
		}
	}
	return true
}

// hiddenCells lists the cells still to be resolved: unrevealed and unflagged.
func (b *Board) hiddenCells() []Coord {
	var out []Coord
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			c := Coord{X: x, Y: y}
			cell := b.cell(c)
			if !cell.Revealed && cell.Mark != MarkFlag {
				out = append(out, c)
			}
		}
	}
	return out
}

// frontierConstraints turns each revealed number that still borders hidden
// cells into a constraint, with already-flagged mines discounted.
func (b *Board) frontierConstraints() []constraint {
	var cs []constraint
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			c := Coord{X: x, Y: y}
			cell := b.cell(c)
			if !cell.Revealed || cell.Adjacent == 0 {
				continue
			}
			var cells []Coord
			need := int(cell.Adjacent)
			for _, n := range c.Neighbors(b.width, b.height) {
				switch nc := b.cell(n); {
				case nc.Mark == MarkFlag:
					need--
				case !nc.Revealed:
					cells = append(cells, n)
				}
			}
			if len(cells) > 0 {
				cs = append(cs, constraint{cells: cells, need: need})
			}
		}
	}
	return cs
}

// deduce derives everything the given constraints imply. It is pure so the
// reasoning can be tested without a board.
func deduce(cs []constraint, hidden []Coord, minesLeft int) (safe, mines []Coord) {
	safeSet := make(map[Coord]struct{})
	mineSet := make(map[Coord]struct{})

	mark := func(cells []Coord, need int) {
		switch {
		case need == 0:
			for _, c := range cells {
				safeSet[c] = struct{}{}
			}
		case need == len(cells):
			for _, c := range cells {
				mineSet[c] = struct{}{}
			}
		}
	}

	for _, c := range cs {
		mark(c.cells, c.need)
	}
	for _, pair := range overlappingPairs(cs) {
		a, b := cs[pair[0]], cs[pair[1]]
		if rest, ok := subtract(b, a); ok {
			mark(rest, b.need-a.need)
		}
	}

	// The mine counter is a constraint too, and it is how most endgames close.
	mark(hidden, minesLeft)

	// A cell deduced both ways means the constraints contradict each other,
	// which only happens if a rule above is wrong. Trust neither.
	for c := range mineSet {
		if _, both := safeSet[c]; both {
			return nil, nil
		}
	}
	return keys(safeSet), keys(mineSet)
}

// overlappingPairs lists ordered constraint index pairs that share a cell.
// Constraints that share nothing can never be subsets of one another, so
// skipping them keeps the pairwise pass cheap on large boards.
func overlappingPairs(cs []constraint) [][2]int {
	byCell := make(map[Coord][]int)
	for i, c := range cs {
		for _, cell := range c.cells {
			byCell[cell] = append(byCell[cell], i)
		}
	}

	seen := make(map[[2]int]struct{})
	var pairs [][2]int
	for _, idxs := range byCell {
		for _, i := range idxs {
			for _, j := range idxs {
				if i == j {
					continue
				}
				p := [2]int{i, j}
				if _, dup := seen[p]; dup {
					continue
				}
				seen[p] = struct{}{}
				pairs = append(pairs, p)
			}
		}
	}
	return pairs
}

// subtract returns b.cells minus a.cells when a is a strict subset of b. The
// remainder then holds exactly b.need-a.need mines, which is the reasoning
// behind patterns like 1-2-1.
func subtract(b, a constraint) ([]Coord, bool) {
	if len(a.cells) >= len(b.cells) {
		return nil, false
	}
	in := make(map[Coord]struct{}, len(b.cells))
	for _, c := range b.cells {
		in[c] = struct{}{}
	}
	for _, c := range a.cells {
		if _, ok := in[c]; !ok {
			return nil, false
		}
		delete(in, c)
	}
	return keys(in), true
}

// keys returns the set in reading order. The deductions themselves do not
// depend on ordering, but a stable order keeps generation reproducible from a
// seed and keeps failures debuggable.
func keys(m map[Coord]struct{}) []Coord {
	if len(m) == 0 {
		return nil
	}
	out := make([]Coord, 0, len(m))
	for c := range m {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}
