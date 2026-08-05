package game

import (
	"fmt"
	"strings"
)

// Preset identifies a built-in difficulty level.
type Preset int

const (
	Beginner Preset = iota
	Intermediate
	Expert
	Custom
)

// Difficulty describes board dimensions and mine count.
type Difficulty struct {
	Preset Preset
	Width  int
	Height int
	Mines  int
}

// PresetDifficulty returns the standard settings for a preset.
func PresetDifficulty(p Preset) Difficulty {
	switch p {
	case Beginner:
		return Difficulty{Preset: Beginner, Width: 9, Height: 9, Mines: 10}
	case Intermediate:
		return Difficulty{Preset: Intermediate, Width: 16, Height: 16, Mines: 40}
	case Expert:
		return Difficulty{Preset: Expert, Width: 30, Height: 16, Mines: 99}
	default:
		return Difficulty{Preset: Custom, Width: 9, Height: 9, Mines: 10}
	}
}

// Validate checks that the difficulty is playable with first-click safety.
func (d Difficulty) Validate() error {
	if d.Width < 1 || d.Height < 1 {
		return fmt.Errorf("board size must be positive")
	}
	total := d.Width * d.Height
	if d.Mines < 1 {
		return fmt.Errorf("mine count must be at least 1")
	}
	if d.Mines > total-9 {
		return fmt.Errorf("mine count %d exceeds maximum %d for first-click safety", d.Mines, total-9)
	}
	return nil
}

// Key returns a stable string for high-score storage.
func (d Difficulty) Key() string {
	switch d.Preset {
	case Beginner:
		return "beginner"
	case Intermediate:
		return "intermediate"
	case Expert:
		return "expert"
	default:
		return fmt.Sprintf("custom_%dx%d_%d", d.Width, d.Height, d.Mines)
	}
}

// PresetFromString parses a preset name.
func PresetFromString(s string) (Preset, bool) {
	switch strings.ToLower(s) {
	case "beginner", "b":
		return Beginner, true
	case "intermediate", "i":
		return Intermediate, true
	case "expert", "e":
		return Expert, true
	case "custom", "c":
		return Custom, true
	default:
		return Beginner, false
	}
}

// String returns a human-readable preset name.
func (p Preset) String() string {
	switch p {
	case Beginner:
		return "beginner"
	case Intermediate:
		return "intermediate"
	case Expert:
		return "expert"
	case Custom:
		return "custom"
	default:
		return "unknown"
	}
}
