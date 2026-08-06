package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/TF0119/minesweeper/internal/game"
)

// configVersion 2 made no-guess boards the default. Files written by version 1
// carry an explicit no_guess:false that was the old default rather than a
// choice, so loading one moves it onto the new default. See migrate.
const configVersion = 2

// Custom holds custom difficulty dimensions.
type Custom struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Mines  int `json:"mines"`
}

// Config is persisted user settings.
type Config struct {
	Version       int    `json:"version"`
	LastPreset    string `json:"last_preset"`
	Custom        Custom `json:"custom"`
	UseEmoji      bool   `json:"use_emoji"`
	QuestionMarks bool   `json:"question_marks"`
	NoGuess       bool   `json:"no_guess"`
	Theme         string `json:"theme"`
}

// DefaultConfig returns factory defaults.
func DefaultConfig() Config {
	return Config{
		Version:       configVersion,
		LastPreset:    "beginner",
		Custom:        Custom{Width: 20, Height: 10, Mines: 30},
		UseEmoji:      false,
		QuestionMarks: true,
		NoGuess:       true,
		Theme:         "classic",
	}
}

// LoadConfig reads config from disk or returns defaults if missing.
func LoadConfig() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}
	// Decode onto the defaults so settings the file predates keep their
	// intended value instead of silently becoming the zero value.
	c := DefaultConfig()
	if err := json.Unmarshal(data, &c); err != nil {
		return DefaultConfig(), err
	}
	return migrate(c), nil
}

// migrate brings an older config file forward. Decoding happens onto the
// defaults, so a file that predates the version field arrives already carrying
// the current one and needs nothing.
func migrate(c Config) Config {
	if c.Version < 2 {
		// Version 1 defaulted no-guess off, so a no_guess:false written by it
		// records the old default rather than a decision. Version 2 promises
		// boards that deduction alone can clear, and anyone who wants the coin
		// flips back can say so in Settings.
		c.NoGuess = true
	}
	c.Version = configVersion
	return c
}

// SaveConfig writes config atomically.
func SaveConfig(c Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	c.Version = configVersion
	return writeJSONAtomic(path, c)
}

// DifficultyFromConfig resolves difficulty using config defaults.
func DifficultyFromConfig(c Config) game.Difficulty {
	preset, ok := game.PresetFromString(c.LastPreset)
	if !ok {
		preset = game.Beginner
	}
	if preset == game.Custom {
		return game.Difficulty{
			Preset: game.Custom,
			Width:  c.Custom.Width,
			Height: c.Custom.Height,
			Mines:  c.Custom.Mines,
		}
	}
	return game.PresetDifficulty(preset)
}

func writeJSONAtomic(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
