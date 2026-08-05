# Contributing

Thanks for taking the time to look at this project. Bug reports, ideas, and
patches are all welcome.

## Getting started

You need Go 1.25 or newer. Everything else is standard library or already
vendored through `go.mod`.

```bash
git clone https://github.com/TF0119/minesweeper.git
cd minesweeper
make test
make run
```

`make help` lists the rest of the targets.

## Before you open a pull request

```bash
make check   # fmt check, vet, race tests, goreleaser config
```

CI runs the same checks on Linux, macOS, and Windows, so a green `make check`
usually means a green build.

## How the code is organised

The one rule that shapes everything else: **game rules never depend on the
terminal.**

| Package | Responsibility |
|---------|----------------|
| `internal/game` | Board state and rules. No I/O, no bubbletea, no colour. |
| `internal/ui` | Input handling and rendering. Sees `CellView`, never mine positions. |
| `internal/storage` | Config and high scores on disk. |
| `cmd/minesweeper` | Flag parsing and wiring. |

If a change makes `internal/game` import something from `internal/ui`, or makes
the UI reach for a mine position directly, it is going the wrong way. Read
[docs/design.md](docs/design.md) before restructuring anything — it records why
the boundaries are where they are.

## What good changes look like

- **Rules changes come with tests in `internal/game`.** That package is pure, so
  there is no excuse for an untested rule.
- **State is never carried by colour alone.** A cell must be identifiable on a
  monochrome terminal. There is a regression test for this.
- **New flags need a default that makes them unnecessary.** Every option is a
  question the user has to answer; add one only when no default can be right.
- **Prefer deleting.** A smaller interface that does the same job is a better
  patch than a larger one.

## Commit messages

Keep the subject in the imperative and scoped to one change, for example
`ui: keep the cursor visible without colour`. Split unrelated work into separate
commits so each one can be reverted on its own.

## Re-recording the demo

`docs/demo.gif` is generated from `docs/demo.tape` with
[VHS](https://github.com/charmbracelet/vhs):

```bash
make demo
```

The tape pins the board with `-seed 2`, so a re-recording produces the same
game. Only re-record when the UI actually changed.

## Reporting bugs

Open an issue with the terminal, OS, and the seed shown in the status line. The
seed is usually enough to reproduce a board bug exactly.
