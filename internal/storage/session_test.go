package storage

import (
	"os"
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
)

func sampleSession() Session {
	return Session{
		Seed:       game.Seed(4242),
		Difficulty: game.PresetDifficulty(game.Intermediate),
		NoGuess:    true,
		Moves: []game.Move{
			{Kind: game.MoveReveal, Coord: game.Coord{X: 3, Y: 3}},
			{Kind: game.MoveChord, Coord: game.Coord{X: 4, Y: 3}},
		},
		Seconds: 42,
		Cursor:  game.Coord{X: 5, Y: 6},
	}
}

func TestSaveAndLoadSessionRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	want := sampleSession()
	if err := SaveSession(want); err != nil {
		t.Fatal(err)
	}

	got, ok, err := LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("saved session was not found")
	}
	if got.Seed != want.Seed || got.Difficulty != want.Difficulty {
		t.Errorf("board = %v %+v, want %v %+v", got.Seed, got.Difficulty, want.Seed, want.Difficulty)
	}
	if got.NoGuess != want.NoGuess || got.Seconds != want.Seconds || got.Cursor != want.Cursor {
		t.Errorf("session = %+v, want no-guess/seconds/cursor of %+v", got, want)
	}
	if len(got.Moves) != len(want.Moves) {
		t.Fatalf("moves = %d, want %d", len(got.Moves), len(want.Moves))
	}
	for i := range want.Moves {
		if got.Moves[i] != want.Moves[i] {
			t.Errorf("move %d = %+v, want %+v", i, got.Moves[i], want.Moves[i])
		}
	}
	if got.SavedAt.IsZero() {
		t.Error("SavedAt should be stamped on save")
	}
}

func TestLoadSessionMissingIsNotAnError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, ok, err := LoadSession()
	if err != nil {
		t.Fatalf("missing session reported an error: %v", err)
	}
	if ok {
		t.Error("no session should have been found")
	}
}

func TestLoadSessionRejectsNothingToResume(t *testing.T) {
	tests := []struct {
		name string
		s    Session
	}{
		{"no moves", Session{Seed: 1, Difficulty: game.PresetDifficulty(game.Beginner)}},
		{"unbuildable board", Session{
			Seed:       1,
			Difficulty: game.Difficulty{Preset: game.Custom, Width: 0, Height: 0, Mines: 99},
			Moves:      []game.Move{{Kind: game.MoveReveal}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			if err := SaveSession(tt.s); err != nil {
				t.Fatal(err)
			}
			if _, ok, err := LoadSession(); ok || err != nil {
				t.Errorf("ok = %v, err = %v; want the session ignored", ok, err)
			}
		})
	}
}

func TestClearSession(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := ClearSession(); err != nil {
		t.Errorf("clearing a missing session: %v", err)
	}
	if err := SaveSession(sampleSession()); err != nil {
		t.Fatal(err)
	}
	if err := ClearSession(); err != nil {
		t.Fatal(err)
	}
	path, _ := SessionPath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("session file still present: %v", err)
	}
	if _, ok, _ := LoadSession(); ok {
		t.Error("cleared session still loads")
	}
}
