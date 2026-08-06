package main

import (
	"flag"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
)

func TestResolveDifficultyPrecedence(t *testing.T) {
	cfg := storage.DefaultConfig()
	cfg.LastPreset = "expert"
	cfg.Custom = storage.Custom{Width: 20, Height: 10, Mines: 30}

	tests := []struct {
		name       string
		flagName   string
		w, h, m    int
		wantPreset game.Preset
		wantW      int
		wantErr    bool
	}{
		{"config default", "", 0, 0, 0, game.Expert, 30, false},
		{"flag overrides config", "beginner", 0, 0, 0, game.Beginner, 9, false},
		{"bare dimensions imply custom", "", 12, 8, 15, game.Custom, 12, false},
		{"custom fills gaps from config", "custom", 0, 0, 0, game.Custom, 20, false},
		{"unknown name errors", "wizard", 0, 0, 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := resolveDifficulty(tt.flagName, tt.w, tt.h, tt.m, cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if d.Preset != tt.wantPreset || d.Width != tt.wantW {
				t.Errorf("got %+v, want preset %v width %d", d, tt.wantPreset, tt.wantW)
			}
		})
	}
}

func TestResolveSeed(t *testing.T) {
	todaysDaily := game.DailySeed(time.Now())

	t.Run("explicit number", func(t *testing.T) {
		got, err := resolveSeed("777", false)
		if err != nil || got != game.Seed(777) {
			t.Fatalf("got %v, err %v", got, err)
		}
	})

	t.Run("daily flag", func(t *testing.T) {
		got, err := resolveSeed("", true)
		if err != nil || got != todaysDaily {
			t.Fatalf("got %v, want %v (err %v)", got, todaysDaily, err)
		}
	})

	t.Run("daily keyword", func(t *testing.T) {
		got, err := resolveSeed(game.DailyKeyword, false)
		if err != nil || got != todaysDaily {
			t.Fatalf("got %v, want %v (err %v)", got, todaysDaily, err)
		}
	})

	t.Run("conflicting flags", func(t *testing.T) {
		if _, err := resolveSeed("123", true); err == nil {
			t.Error("expected an error when -daily and -seed disagree")
		}
	})

	t.Run("invalid seed", func(t *testing.T) {
		if _, err := resolveSeed("not-a-seed", false); err == nil {
			t.Error("expected an error for a non-numeric seed")
		}
	})

	t.Run("no seed is random", func(t *testing.T) {
		if _, err := resolveSeed("", false); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRunVersionFlag(t *testing.T) {
	if err := run([]string{"-version"}); err != nil {
		t.Errorf("-version should exit cleanly, got %v", err)
	}
}

func TestBuildVersion(t *testing.T) {
	restore := version
	t.Cleanup(func() { version = restore })

	version = "v1.2.3"
	if got := buildVersion(); got != version {
		t.Errorf("buildVersion() = %q, want the linker flag value %q", got, version)
	}

	// Without a linker flag the toolchain's build info answers instead. Its
	// exact value depends on how the test binary was built; what matters is
	// that -version never prints a blank.
	version = ""
	if got := buildVersion(); got == "" {
		t.Error("buildVersion() returned an empty string")
	}
}

func TestFlagsGivenDistinguishesAbsentFromFalse(t *testing.T) {
	newSet := func() (*flag.FlagSet, *bool) {
		fs := flag.NewFlagSet("t", flag.ContinueOnError)
		b := fs.Bool("no-guess", false, "")
		return fs, b
	}

	fs, _ := newSet()
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if flagsGiven(fs)["no-guess"] {
		t.Error("an omitted flag should not count as given")
	}

	fs, value := newSet()
	if err := fs.Parse([]string{"-no-guess=false"}); err != nil {
		t.Fatal(err)
	}
	if !flagsGiven(fs)["no-guess"] {
		t.Error("-no-guess=false should count as given, so it can switch a saved preference off")
	}
	if *value {
		t.Error("-no-guess=false should parse as false")
	}
}

func TestBoardRequestedGuardsTheSavedGame(t *testing.T) {
	tests := []struct {
		name  string
		given []string
		want  bool
	}{
		{"nothing typed", nil, false},
		{"appearance only", []string{"theme", "no-color", "no-guess"}, false},
		{"a difficulty", []string{"difficulty"}, true},
		{"custom dimensions", []string{"width", "height"}, true},
		{"a seed", []string{"seed"}, true},
		{"the daily challenge", []string{"daily"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			given := make(map[string]bool, len(tt.given))
			for _, name := range tt.given {
				given[name] = true
			}
			if got := boardRequested(given); got != tt.want {
				t.Errorf("boardRequested(%v) = %v, want %v", tt.given, got, tt.want)
			}
		})
	}
}
