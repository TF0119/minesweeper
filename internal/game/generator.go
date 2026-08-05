package game

import (
	"fmt"
	"math/rand"
)

func placeMines(b *Board, rng *rand.Rand, forbidden map[Coord]struct{}) error {
	total := b.width * b.height
	if b.mineCount > total-len(forbidden) {
		return fmt.Errorf("cannot place %d mines with %d forbidden cells", b.mineCount, len(forbidden))
	}

	indices := make([]int, 0, total-len(forbidden))
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			c := Coord{X: x, Y: y}
			if _, skip := forbidden[c]; skip {
				continue
			}
			indices = append(indices, c.index(b.width))
		}
	}

	rng.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	for i := 0; i < b.mineCount; i++ {
		idx := indices[i]
		b.cells[idx].HasMine = true
	}

	computeAdjacent(b)
	return nil
}

func computeAdjacent(b *Board) {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			c := Coord{X: x, Y: y}
			idx := c.index(b.width)
			if b.cells[idx].HasMine {
				b.cells[idx].Adjacent = 0
				continue
			}
			var count uint8
			for _, n := range c.Neighbors(b.width, b.height) {
				if b.cells[n.index(b.width)].HasMine {
					count++
				}
			}
			b.cells[idx].Adjacent = count
		}
	}
}

func relocateMinesFromSafeZone(b *Board, safe map[Coord]struct{}, rng *rand.Rand) error {
	outsideNonMines := make([]int, 0)
	safeMines := make([]int, 0)

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			c := Coord{X: x, Y: y}
			idx := c.index(b.width)
			_, inSafe := safe[c]
			if inSafe && b.cells[idx].HasMine {
				safeMines = append(safeMines, idx)
			} else if !inSafe && !b.cells[idx].HasMine {
				outsideNonMines = append(outsideNonMines, idx)
			}
		}
	}

	if len(safeMines) > len(outsideNonMines) {
		return fmt.Errorf("not enough non-mine cells outside safe zone")
	}

	rng.Shuffle(len(outsideNonMines), func(i, j int) {
		outsideNonMines[i], outsideNonMines[j] = outsideNonMines[j], outsideNonMines[i]
	})

	for i, mineIdx := range safeMines {
		swapIdx := outsideNonMines[i]
		b.cells[mineIdx].HasMine = false
		b.cells[swapIdx].HasMine = true
	}

	computeAdjacent(b)
	return nil
}

// maxNoGuessAttempts bounds the search for a layout that needs no guessing.
// Most boards are found in a handful of tries; the cap exists so that a hard
// combination of size and mine density cannot stall the opening move. Running
// out means the player gets an ordinary board, which is worse than promised
// but far better than a game that refuses to start.
const maxNoGuessAttempts = 400

// placeSolvableMines lays out mines that can be cleared from first by pure
// deduction, and reports whether it managed to. Each attempt reshuffles from
// the board's own generator, so a seed still reproduces its board exactly.
func placeSolvableMines(b *Board, safe map[Coord]struct{}, first Coord, rng *rand.Rand) (bool, error) {
	for attempt := 0; attempt < maxNoGuessAttempts; attempt++ {
		clearMines(b)
		if err := placeMines(b, rng, safe); err != nil {
			return false, err
		}
		if solvable(b, first) {
			return true, nil
		}
	}
	return false, nil
}

func clearMines(b *Board) {
	for i := range b.cells {
		b.cells[i].HasMine = false
		b.cells[i].Adjacent = 0
	}
}

func safeZone(first Coord, w, h int) map[Coord]struct{} {
	safe := make(map[Coord]struct{})
	safe[first] = struct{}{}
	for _, n := range first.Neighbors(w, h) {
		safe[n] = struct{}{}
	}
	return safe
}
