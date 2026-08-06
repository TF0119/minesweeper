package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TF0119/minesweeper/internal/storage/storagetest"
)

// Tests must never touch the real save files, and the check belongs next to
// the helper that promises it.
func TestIsolateConfigDirCoversThisPlatform(t *testing.T) {
	dir := storagetest.IsolateConfigDir(t)
	got, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, dir) {
		t.Errorf("ConfigDir() = %q, want it under the temporary %q", got, dir)
	}
}

func TestSaveLoadConfigRoundTrip(t *testing.T) {
	storagetest.IsolateConfigDir(t)

	cfg := DefaultConfig()
	cfg.LastPreset = "intermediate"
	cfg.UseEmoji = true
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LastPreset != "intermediate" || !loaded.UseEmoji {
		t.Errorf("round trip mismatch: %+v", loaded)
	}

	path, _ := ConfigPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not written: %v", err)
	}
}

func TestAtomicWriteNoPartialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	if err := writeJSONAtomic(path, DefaultConfig()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		t.Fatal("expected valid json file")
	}
}

func TestHighScoreUpdate(t *testing.T) {
	h := DefaultHighScores()
	d := DifficultyFromConfig(DefaultConfig())
	if !h.TryUpdate("beginner", 100, d) {
		t.Error("first record should update")
	}
	if h.TryUpdate("beginner", 120, d) {
		t.Error("slower time should not update")
	}
	if !h.TryUpdate("beginner", 50, d) {
		t.Error("faster time should update")
	}
	if h.Best("beginner") != 50 {
		t.Errorf("best = %d, want 50", h.Best("beginner"))
	}
}

// The game promises boards that deduction alone can clear, so that has to be
// what a fresh install does.
func TestConfigDefaultsToNoGuess(t *testing.T) {
	if !DefaultConfig().NoGuess {
		t.Error("DefaultConfig().NoGuess = false, want no-guess boards by default")
	}
}

// writeConfigFile drops raw JSON where LoadConfig will find it.
func writeConfigFile(t *testing.T, body string) {
	t.Helper()
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigMovesVersionOneOntoNoGuess(t *testing.T) {
	storagetest.IsolateConfigDir(t)
	// Version 1 defaulted no-guess off, so its no_guess:false is the old
	// default rather than something the player asked for.
	writeConfigFile(t, `{"version":1,"no_guess":false,"theme":"dark","last_preset":"expert"}`)

	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !c.NoGuess {
		t.Error("a version 1 config should come forward onto no-guess boards")
	}
	if c.Version != configVersion {
		t.Errorf("Version = %d, want %d after migration", c.Version, configVersion)
	}
	if c.Theme != "dark" || c.LastPreset != "expert" {
		t.Errorf("migration disturbed unrelated settings: %+v", c)
	}
}

// Once the file is at the current version, no_guess:false is a decision and
// migration must leave it alone.
func TestLoadConfigKeepsNoGuessOffOnceChosen(t *testing.T) {
	storagetest.IsolateConfigDir(t)
	writeConfigFile(t, `{"version":2,"no_guess":false}`)

	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.NoGuess {
		t.Error("an explicit no_guess:false at the current version was overwritten")
	}
}
