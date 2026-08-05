package game

import (
	"math/rand"
	"testing"
)

func testRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func TestValidateDifficulty(t *testing.T) {
	tests := []struct {
		name    string
		d       Difficulty
		wantErr bool
	}{
		{"beginner ok", PresetDifficulty(Beginner), false},
		{"intermediate ok", PresetDifficulty(Intermediate), false},
		{"expert ok", PresetDifficulty(Expert), false},
		{"zero width", Difficulty{Width: 0, Height: 9, Mines: 1}, true},
		{"zero mines", Difficulty{Width: 9, Height: 9, Mines: 0}, true},
		{"too many mines", Difficulty{Width: 3, Height: 3, Mines: 9}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaceMinesCount(t *testing.T) {
	d := PresetDifficulty(Beginner)
	b := NewBoard(d, testRNG(42))
	if err := placeMines(b, testRNG(42), map[Coord]struct{}{}); err != nil {
		t.Fatal(err)
	}
	count := 0
	for i := range b.cells {
		if b.cells[i].HasMine {
			count++
		}
	}
	if count != d.Mines {
		t.Errorf("got %d mines, want %d", count, d.Mines)
	}
}

func TestRelocateSafeZone(t *testing.T) {
	d := PresetDifficulty(Beginner)
	for seed := int64(0); seed < 50; seed++ {
		b := NewBoard(d, testRNG(seed))
		if err := placeMines(b, testRNG(seed), map[Coord]struct{}{}); err != nil {
			t.Fatal(err)
		}
		first := Coord{X: 4, Y: 4}
		safe := safeZone(first, b.width, b.height)
		if err := relocateMinesFromSafeZone(b, safe, testRNG(seed+1)); err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		for c := range safe {
			if b.cell(c).HasMine {
				t.Fatalf("seed %d: mine in safe zone at %+v", seed, c)
			}
		}
	}
}

func TestPresetKey(t *testing.T) {
	if PresetDifficulty(Beginner).Key() != "beginner" {
		t.Error("beginner key mismatch")
	}
}
