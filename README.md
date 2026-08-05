<div align="center">

# minesweeper

**Terminal Minesweeper with a testable game core separated from the TUI.**

Classic rules — first-click safety, flood fill, chord, timer, high scores —
plus shareable board seeds and a daily challenge.

[![CI](https://github.com/TF0119/minesweeper/actions/workflows/ci.yml/badge.svg)](https://github.com/TF0119/minesweeper/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/TF0119/minesweeper?logo=github)](https://github.com/TF0119/minesweeper/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/TF0119/minesweeper)](https://goreportcard.com/report/github.com/TF0119/minesweeper)
[![Go Reference](https://pkg.go.dev/badge/github.com/TF0119/minesweeper.svg)](https://pkg.go.dev/github.com/TF0119/minesweeper)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

<img src="docs/demo.gif" alt="Playing a seeded beginner board: the opening move cascades, a mine is flagged, and a chord opens the rest." width="820">

</div>

## Install

Download a binary for your platform from the
[releases page](https://github.com/TF0119/minesweeper/releases/latest), or build
it yourself with Go 1.25+:

```bash
go install github.com/TF0119/minesweeper/cmd/minesweeper@latest
```

That installs into `$(go env GOPATH)/bin` — usually `~/go/bin`, which has to be
on your `PATH`. Running the same command again upgrades to the newest tagged
release; `@latest` follows release tags, not `main`. Check what you have with
`minesweeper -version`.

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
small terminals. Cell state never depends on colour alone, so `-no-color`,
`NO_COLOR=1`, and monochrome terminals stay playable.

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
make test     # go test -race ./...
make build    # binary at bin/minesweeper
make lint     # gofmt, go vet, golangci-lint
make demo     # re-record docs/demo.gif (needs vhs)
```

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for the
workflow and [docs/design.md](docs/design.md) for the reasoning you are expected
to work with.

## License

[MIT](LICENSE) © Takeru Fukuda
