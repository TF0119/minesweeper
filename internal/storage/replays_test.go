package storage

import (
	"fmt"
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

func TestListReplaysSkipsCorruptAndFillsLimit(t *testing.T) {
	storagetest.IsolateConfigDir(t)
	dir, err := ReplaysDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Newer corrupt file would steal a slot if limit applied before parse.
	corrupt := filepath.Join(dir, "20260806-120000-won-seed9.json")
	if err := os.WriteFile(corrupt, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	older := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
		Won:        true,
		Seconds:    5,
		PlayedAt:   time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
	}
	if err := SaveReplay(older); err != nil {
		t.Fatal(err)
	}

	got, err := ListReplays(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Seed != older.Seed {
		t.Fatalf("list = %+v, want the valid older replay", got)
	}
	if _, err := os.Stat(corrupt); !os.IsNotExist(err) {
		t.Error("corrupt replay file should have been removed")
	}
}

func TestDeleteReplay(t *testing.T) {
	storagetest.IsolateConfigDir(t)

	r := game.Replay{
		ID:         "to-delete",
		Seed:       game.Seed(2),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal}},
		PlayedAt:   time.Date(2026, 8, 6, 1, 0, 0, 0, time.UTC),
	}
	if err := SaveReplay(r); err != nil {
		t.Fatal(err)
	}
	if err := DeleteReplay(r.ID); err != nil {
		t.Fatal(err)
	}
	got, err := ListReplays(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("list = %+v, want empty after delete", got)
	}
	if err := DeleteReplay("missing"); err != nil {
		t.Errorf("missing id: %v", err)
	}
	if err := DeleteReplay("../etc"); err == nil {
		t.Error("path separators in id should be rejected")
	}
}

func TestSaveReplayPrunesBeyondMax(t *testing.T) {
	storagetest.IsolateConfigDir(t)

	for i := 0; i < MaxReplays+3; i++ {
		r := game.Replay{
			Seed:       game.Seed(i + 1),
			Difficulty: game.PresetDifficulty(game.Beginner),
			Moves:      []game.Move{{Kind: game.MoveReveal}},
			PlayedAt:   time.Date(2026, 8, 6, 0, 0, i, 0, time.UTC),
		}
		r.ID = fmt.Sprintf("20260806-0000%02d-won-seed%d", i, i+1)
		if err := SaveReplay(r); err != nil {
			t.Fatal(err)
		}
	}

	dir, err := ReplaysDir()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != MaxReplays {
		t.Errorf("files on disk = %d, want %d", len(entries), MaxReplays)
	}
	got, err := ListReplays(MaxReplays + 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != MaxReplays {
		t.Errorf("list = %d, want %d", len(got), MaxReplays)
	}
}
