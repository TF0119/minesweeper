# minesweeper

Terminal Minesweeper with a **testable game core** separated from the TUI.

Classic Windows-style rules: first-click safety, flood fill, chord, timer, and high scores.

```
 Mines:010  Time:042  Best:038  [beginner]
 ┌───┬───┬───┐
 │   │ 1 │   │
 ├───┼───┼───┤
 │ F │ 2 │ * │
 └───┴───┴───┘
 ↑↓←→/hjkl move · Space reveal · f flag · c chord · ? help · q quit
```

## Install

Requires Go 1.22+.

```bash
go install github.com/takeru0119/minesweeper/cmd/minesweeper@latest
```

Or build from source:

```bash
git clone https://github.com/takeru0119/minesweeper.git
cd minesweeper
go build -o minesweeper ./cmd/minesweeper
./minesweeper
```

## Usage

```bash
minesweeper                              # last difficulty or Beginner
minesweeper --difficulty expert
minesweeper --difficulty custom --width 20 --height 10 --mines 30
minesweeper --no-color
```

Config and high scores are stored in `~/.config/minesweeper/`.

## Controls

| Key | Action |
|-----|--------|
| Arrow / hjkl | Move cursor |
| Space / Enter | Reveal |
| f | Toggle flag |
| c | Chord |
| n | New game |
| d | Difficulty menu |
| ? | Help |
| q / Ctrl+C | Quit |
| Mouse left | Reveal |
| Mouse right | Flag |

## Architecture

Game logic lives in `internal/game` (pure Go, fully unit-tested). The bubbletea UI is a thin adapter. See [docs/design.md](docs/design.md) for design decisions (APoSD).

```
cmd/minesweeper → internal/ui → internal/game
                              → internal/storage
```

## Development

```bash
go test ./...
go build -o bin/minesweeper ./cmd/minesweeper
```

Contributions welcome — open an issue or PR.

## License

[MIT](LICENSE)
