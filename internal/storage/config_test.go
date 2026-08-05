package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	orig := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer os.Setenv("XDG_CONFIG_HOME", orig)

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
