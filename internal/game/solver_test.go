package game

import (
	"testing"
	"time"
)

func TestDeduceSimpleRules(t *testing.T) {
	a, b, c := Coord{0, 0}, Coord{1, 0}, Coord{2, 0}

	t.Run("satisfied constraint clears its cells", func(t *testing.T) {
		safe, mines := deduce([]constraint{{cells: []Coord{a, b}, need: 0}}, nil, 5)
		if len(mines) != 0 || len(safe) != 2 {
			t.Fatalf("safe = %v, mines = %v, want both cells safe", safe, mines)
		}
	})

	t.Run("full constraint marks its cells", func(t *testing.T) {
		safe, mines := deduce([]constraint{{cells: []Coord{a, b}, need: 2}}, nil, 5)
		if len(safe) != 0 || len(mines) != 2 {
			t.Fatalf("safe = %v, mines = %v, want both cells mined", safe, mines)
		}
	})

	t.Run("ambiguous constraint yields nothing", func(t *testing.T) {
		safe, mines := deduce([]constraint{{cells: []Coord{a, b, c}, need: 1}}, nil, 5)
		if len(safe) != 0 || len(mines) != 0 {
			t.Fatalf("safe = %v, mines = %v, want no deduction", safe, mines)
		}
	})
}

// The subset rule is what lets a player read patterns like 1-2-1 instead of
// guessing, so it gets its own test.
func TestDeduceSubsetRule(t *testing.T) {
	shared, extra := Coord{0, 0}, Coord{1, 0}
	// One mine among {shared}; one mine among {shared, extra}. The extra cell
	// is therefore safe, even though neither constraint says so on its own.
	cs := []constraint{
		{cells: []Coord{shared}, need: 1},
		{cells: []Coord{shared, extra}, need: 1},
	}
	safe, mines := deduce(cs, nil, 5)

	if len(safe) != 1 || safe[0] != extra {
		t.Errorf("safe = %v, want just %v", safe, extra)
	}
	if len(mines) != 1 || mines[0] != shared {
		t.Errorf("mines = %v, want just %v", mines, shared)
	}
}

func TestDeduceUsesTheMineCounter(t *testing.T) {
	hidden := []Coord{{0, 0}, {1, 0}}

	safe, mines := deduce(nil, hidden, 0)
	if len(mines) != 0 || len(safe) != 2 {
		t.Errorf("with no mines left every hidden cell is safe: safe = %v, mines = %v", safe, mines)
	}

	safe, mines = deduce(nil, hidden, 2)
	if len(safe) != 0 || len(mines) != 2 {
		t.Errorf("with as many mines as cells every cell is mined: safe = %v, mines = %v", safe, mines)
	}
}

// The generator's promise is only as good as the solver's verdict, so check
// the verdict against an actual play-through.
func TestNoGuessBoardsAreSolvedWithoutGuessing(t *testing.T) {
	for _, preset := range []Preset{Beginner, Intermediate, Expert} {
		t.Run(preset.String(), func(t *testing.T) {
			d := PresetDifficulty(preset)
			first := Coord{X: d.Width / 2, Y: d.Height / 2}

			for seed := uint32(1); seed <= 5; seed++ {
				b := NewNoGuessBoard(d, Seed(seed))
				b.Reveal(first)
				if !b.NoGuess() {
					t.Errorf("seed %d: generator could not find a no-guess layout", seed)
					continue
				}
				b.resetReveals()
				if !solvable(b, first) {
					t.Errorf("seed %d: board reported no-guess but replay needed a guess", seed)
				}
			}
		})
	}
}

// A seed has to mean the same board every time, or -seed and the daily
// challenge are worthless.
func TestNoGuessGenerationIsReproducible(t *testing.T) {
	d := PresetDifficulty(Intermediate)
	first := Coord{X: 8, Y: 8}

	layout := func() []bool {
		b := NewNoGuessBoard(d, Seed(4242))
		b.Reveal(first)
		out := make([]bool, len(b.cells))
		for i := range b.cells {
			out[i] = b.cells[i].HasMine
		}
		return out
	}

	first1, second := layout(), layout()
	for i := range first1 {
		if first1[i] != second[i] {
			t.Fatalf("cell %d differs between two generations of the same seed", i)
		}
	}
}

// Generation happens while the player waits for their opening click, so the
// search has to stay comfortably under a noticeable pause.
func TestNoGuessGenerationIsFastEnoughForTheOpeningMove(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test")
	}
	const budget = 2 * time.Second

	for _, preset := range []Preset{Beginner, Intermediate, Expert} {
		t.Run(preset.String(), func(t *testing.T) {
			d := PresetDifficulty(preset)
			start := time.Now()
			b := NewNoGuessBoard(d, Seed(7))
			b.Reveal(Coord{X: d.Width / 2, Y: d.Height / 2})
			if elapsed := time.Since(start); elapsed > budget {
				t.Errorf("generation took %v, over the %v budget", elapsed, budget)
			}
		})
	}
}
