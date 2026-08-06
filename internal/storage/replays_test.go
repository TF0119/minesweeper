package storage

import (
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
)

func TestSaveAndListReplay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	r := game.Replay{
		Seed:       game.Seed(42),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
		Won:        true,
		Seconds:    30,
		PlayedAt:   time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
	}
	if err := SaveReplay(r); err != nil {
		t.Fatal(err)
	}

	got, err := ListReplays(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Seed != r.Seed || !got[0].Won || got[0].Seconds != 30 {
		t.Errorf("round trip = %+v, want %+v", got[0], r)
	}
}
