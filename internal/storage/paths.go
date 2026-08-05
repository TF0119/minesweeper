package storage

import (
	"os"
	"path/filepath"
)

const appName = "minesweeper"

// ConfigDir returns the application config directory.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// ConfigPath returns the path to config.json.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// HighScorePath returns the path to highscores.json.
func HighScorePath() (string, error) {
	return dataFile("highscores.json")
}

// StatsPath returns the path to stats.json.
func StatsPath() (string, error) {
	return dataFile("stats.json")
}

func dataFile(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}
