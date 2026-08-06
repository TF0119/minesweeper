package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage/storagetest"
)

func TestSaveAndListReplay(t *testing.T) {
	storagetest.IsolateConfigDir(t)

	r := game.Replay{
		Seed:       game.Seed(42),
		Difficulty: game.PresetDifficulty(game.Beginner),
		NoGuess:    true,
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
	if got[0].Seed != r.Seed || !got[0].Won || got[0].Seconds != 30 || !got[0].NoGuess {
		t.Errorf("round trip = %+v, want %+v", got[0], r)
	}
}

func TestReplayOmittingNoGuessLoadsAsClassic(t *testing.T) {
	storagetest.IsolateConfigDir(t)

	dir, err := ReplaysDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Older recordings have no no_guess field.
	raw := `{
		"version": 1,
		"replay": {
			"id": "old",
			"seed": 9,
			"difficulty": {"Preset": 0, "Width": 9, "Height": 9, "Mines": 10},
			"moves": [{"kind": 0, "X": 4, "Y": 4}],
			"won": true,
			"seconds": 10,
			"played_at": "2026-08-05T12:00:00Z"
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, "old.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ListReplays(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].NoGuess {
		t.Error("missing no_guess should load as classic (false)")
	}
}
