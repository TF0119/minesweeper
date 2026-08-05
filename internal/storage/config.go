package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/takeru0119/minesweeper/internal/game"
)

const configVersion = 1

// Custom holds custom difficulty dimensions.
type Custom struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Mines  int `json:"mines"`
}

// Config is persisted user settings.
type Config struct {
	Version    int    `json:"version"`
	LastPreset string `json:"last_preset"`
	Custom     Custom `json:"custom"`
	UseEmoji   bool   `json:"use_emoji"`
}

// DefaultConfig returns factory defaults.
func DefaultConfig() Config {
	return Config{
		Version:    configVersion,
		LastPreset: "beginner",
		Custom:     Custom{Width: 20, Height: 10, Mines: 30},
		UseEmoji:   false,
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
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return DefaultConfig(), err
	}
	if c.Version == 0 {
		c.Version = configVersion
	}
	return c, nil
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
