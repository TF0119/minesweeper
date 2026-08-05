package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/takeru0119/minesweeper/internal/game"
	"github.com/takeru0119/minesweeper/internal/storage"
	"github.com/takeru0119/minesweeper/internal/ui"
)

const version = "0.1.0"

func main() {
	difficultyFlag := flag.String("difficulty", "", "beginner|intermediate|expert|custom")
	widthFlag := flag.Int("width", 0, "custom board width")
	heightFlag := flag.Int("height", 0, "custom board height")
	minesFlag := flag.Int("mines", 0, "custom mine count")
	noColor := flag.Bool("no-color", false, "disable colors")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println("minesweeper", version)
		os.Exit(0)
	}

	cfg, err := storage.LoadConfig()
	if err != nil {
		log.Printf("config: %v (using defaults)", err)
		cfg = storage.DefaultConfig()
	}

	scores, err := storage.LoadHighScores()
	if err != nil {
		log.Printf("highscores: %v (using defaults)", err)
		scores = storage.DefaultHighScores()
	}

	d := resolveDifficulty(*difficultyFlag, *widthFlag, *heightFlag, *minesFlag, cfg)
	if err := d.Validate(); err != nil {
		log.Fatal(err)
	}

	cfg.LastPreset = d.Preset.String()
	if d.Preset == game.Custom {
		cfg.Custom = storage.Custom{Width: d.Width, Height: d.Height, Mines: d.Mines}
	}
	_ = storage.SaveConfig(cfg)

	if err := ui.Run(ui.Options{
		Difficulty: d,
		Config:     cfg,
		HighScores: scores,
		NoColor:    *noColor,
	}); err != nil {
		log.Fatal(err)
	}
}

func resolveDifficulty(flagDiff string, w, h, mines int, cfg storage.Config) game.Difficulty {
	if flagDiff != "" {
		p, ok := game.PresetFromString(flagDiff)
		if !ok {
			log.Fatalf("unknown difficulty: %s", flagDiff)
		}
		if p == game.Custom {
			return customDifficulty(w, h, mines, cfg)
		}
		return game.PresetDifficulty(p)
	}
	if w > 0 || h > 0 || mines > 0 {
		return customDifficulty(w, h, mines, cfg)
	}
	return storage.DifficultyFromConfig(cfg)
}

func customDifficulty(w, h, mines int, cfg storage.Config) game.Difficulty {
	if w == 0 {
		w = cfg.Custom.Width
	}
	if h == 0 {
		h = cfg.Custom.Height
	}
	if mines == 0 {
		mines = cfg.Custom.Mines
	}
	return game.Difficulty{Preset: game.Custom, Width: w, Height: h, Mines: mines}
}
