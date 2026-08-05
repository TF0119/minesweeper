package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/TF0119/minesweeper/internal/ui"
)

// version is injected into release builds via -ldflags "-X main.version=...".
// It stays empty for binaries produced any other way.
var version string

// buildVersion reports the version to show the user. Release archives carry it
// in a linker flag, but `go install module@version` sets no flags, so fall back
// to the module version the toolchain embeds. Without either — a build from an
// untagged tree — say so rather than inventing a number.
func buildVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "devel"
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "minesweeper:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("minesweeper", flag.ContinueOnError)
	var (
		difficulty  = fs.String("difficulty", "", "beginner, intermediate, expert or custom")
		width       = fs.Int("width", 0, "custom board width")
		height      = fs.Int("height", 0, "custom board height")
		mines       = fs.Int("mines", 0, "custom mine count")
		seedFlag    = fs.String("seed", "", `board seed: a number, or "daily" for today's challenge`)
		daily       = fs.Bool("daily", false, `shorthand for -seed daily`)
		noGuess     = fs.Bool("no-guess", false, "only generate boards solvable without guessing")
		theme       = fs.String("theme", "", "colour theme: classic, dark or colorblind")
		noColor     = fs.Bool("no-color", false, "disable colors")
		showVersion = fs.Bool("version", false, "print version and exit")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	given := flagsGiven(fs)

	if *showVersion {
		fmt.Println("minesweeper", buildVersion())
		return nil
	}

	config := loadConfig()
	scores := loadHighScores()
	stats := loadStats()

	d, err := resolveDifficulty(*difficulty, *width, *height, *mines, config)
	if err != nil {
		return err
	}
	if err := d.Validate(); err != nil {
		return err
	}

	seed, err := resolveSeed(*seedFlag, *daily)
	if err != nil {
		return err
	}

	if given["no-guess"] {
		config.NoGuess = *noGuess
	}
	if given["theme"] {
		// A typo here is worth reporting; a stale name in the config file is
		// not, because that must never stop the game from starting.
		if _, ok := ui.ParseTheme(*theme); !ok {
			return fmt.Errorf("unknown theme %q, want one of: %s", *theme, themeNames())
		}
		config.Theme = *theme
	}

	config = rememberPreferences(config, d)

	return ui.Run(ui.Options{
		Difficulty: d,
		Seed:       seed,
		Config:     config,
		HighScores: scores,
		Stats:      stats,
		NoColor:    *noColor,
	})
}

// loadConfig falls back to defaults so a damaged file never blocks play.
func loadConfig() storage.Config {
	c, err := storage.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "minesweeper: using default config:", err)
		return storage.DefaultConfig()
	}
	return c
}

func loadHighScores() storage.HighScores {
	h, err := storage.LoadHighScores()
	if err != nil {
		fmt.Fprintln(os.Stderr, "minesweeper: using empty high scores:", err)
		return storage.DefaultHighScores()
	}
	return h
}

func themeNames() string {
	names := make([]string, 0, len(ui.Themes))
	for _, t := range ui.Themes {
		names = append(names, string(t))
	}
	return strings.Join(names, ", ")
}

func loadStats() storage.Stats {
	s, err := storage.LoadStats()
	if err != nil {
		fmt.Fprintln(os.Stderr, "minesweeper: using empty statistics:", err)
		return storage.DefaultStats()
	}
	return s
}

// resolveDifficulty applies the precedence documented in the README:
// defaults, then the config file, then flags.
func resolveDifficulty(name string, w, h, mines int, c storage.Config) (game.Difficulty, error) {
	if name != "" {
		preset, ok := game.PresetFromString(name)
		if !ok {
			return game.Difficulty{}, fmt.Errorf("unknown difficulty %q", name)
		}
		if preset != game.Custom {
			return game.PresetDifficulty(preset), nil
		}
		return customDifficulty(w, h, mines, c), nil
	}
	if w > 0 || h > 0 || mines > 0 {
		return customDifficulty(w, h, mines, c), nil
	}
	return storage.DifficultyFromConfig(c), nil
}

func customDifficulty(w, h, mines int, c storage.Config) game.Difficulty {
	if w == 0 {
		w = c.Custom.Width
	}
	if h == 0 {
		h = c.Custom.Height
	}
	if mines == 0 {
		mines = c.Custom.Mines
	}
	return game.Difficulty{Preset: game.Custom, Width: w, Height: h, Mines: mines}
}

func resolveSeed(seedFlag string, daily bool) (game.Seed, error) {
	if daily && seedFlag != "" && seedFlag != game.DailyKeyword {
		return 0, fmt.Errorf("-daily conflicts with -seed %s", seedFlag)
	}
	if daily {
		seedFlag = game.DailyKeyword
	}
	if seedFlag == "" {
		return game.RandomSeed(), nil
	}
	return game.ParseSeed(seedFlag, time.Now())
}

// flagsGiven reports which flags the user actually typed. A bool flag left out
// and a bool flag set to false are indistinguishable by value, and the two mean
// different things when a saved preference is involved.
func flagsGiven(fs *flag.FlagSet) map[string]bool {
	given := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })
	return given
}

// rememberPreferences persists the settings that should still apply next launch.
func rememberPreferences(c storage.Config, d game.Difficulty) storage.Config {
	c.LastPreset = d.Preset.String()
	if d.Preset == game.Custom {
		c.Custom = storage.Custom{Width: d.Width, Height: d.Height, Mines: d.Mines}
	}
	if err := storage.SaveConfig(c); err != nil {
		fmt.Fprintln(os.Stderr, "minesweeper: could not save config:", err)
	}
	return c
}
