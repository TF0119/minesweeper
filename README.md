# minesweeper

Terminal Minesweeper with a testable game core separated from the TUI.

Classic rules — first-click safety, flood fill, chord, timer, high scores — plus
shareable board seeds and a daily challenge.

```
 Mines:007  Time:042  Best:038  [beginner]  seed 1487233901
   1  1
   1     1  1  1
   1  1  1     2  F
         1  1  3  2
 arrows/hjkl move · space reveal · f flag · c chord · n new · r restart · ? help · q quit
```

## Install

Download a binary from the [releases page](https://github.com/TF0119/minesweeper/releases),
or build it yourself with Go 1.22+:

```bash
go install github.com/TF0119/minesweeper/cmd/minesweeper@latest
```

## Usage

```bash
minesweeper                          # last difficulty, random board
minesweeper -difficulty expert
minesweeper -difficulty custom -width 20 -height 10 -mines 30
minesweeper -daily                   # today's challenge, same board for everyone
minesweeper -seed 1487233901         # replay a specific board
minesweeper -no-color
```

Settings and high scores live in `~/.config/minesweeper/`. Options are resolved
in one order: built-in defaults, then the config file, then command-line flags.

### Seeds and the daily challenge

Every board is generated from a seed shown in the status line. Pass that number
back with `-seed` to replay the exact same layout, or press `r` in game to retry
the board you just lost. `-daily` derives the seed from the current UTC date, so
everyone who plays on the same day gets the same board.

Seeds pin the mine layout, not your first click: the opening move is always safe,
so the same seed can still start differently depending on where you click.

## Controls

| Key | Action |
|-----|--------|
| Arrows / `hjkl` | Move cursor |
| Space / Enter | Reveal |
| `f` | Toggle flag |
| `c` | Chord (reveal neighbours once flags match the number) |
| `n` | New board |
| `r` | Restart the same seed |
| `d` | Difficulty menu |
| `?` | Help |
| `q` / Ctrl+C | Quit |

Mouse: left click reveals, right click or Shift+left click flags. Some terminals
(notably Windows Terminal and WSL) intercept right-click for paste, so Shift+left
click is the reliable option there.

Boards larger than the window scroll to follow the cursor, so Expert works on
small terminals.

## Architecture

Game rules live in `internal/game`, a package with no TUI dependencies and full
unit-test coverage. The bubbletea layer translates input and draws; it never sees
mine positions, only the `CellView` projection.

```
cmd/minesweeper → internal/ui → internal/game
                              → internal/storage
```

See [docs/design.md](docs/design.md) for the design decisions behind the split.

## Development

```bash
go test ./...
go build -o bin/minesweeper ./cmd/minesweeper
```

Issues and pull requests are welcome.

## License

[MIT](LICENSE)
