# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2026-08-05

### Added

- Shareable board seeds: the seed is shown in the status line, `-seed` replays a
  board, and `r` retries the current one
- Daily challenge via `-daily`, seeded from the UTC date so every player gets the
  same board
- Board scrolling, so Expert fits terminals narrower than the board
- Pre-built binaries for Linux, macOS, and Windows via GoReleaser
- CI across three operating systems with `-race` tests, `go vet`, and gofmt checks

### Fixed

- Cursor keys and `hjkl` did not move the cursor
- Right-click did not place a flag on terminals that report the button on release
  or use the X10 encoding; Shift+left click now works everywhere
- The timer ran several times too fast after the first finished game, because the
  tick chain was never stopped

### Changed

- `NewBoard` takes a `Seed` instead of a `*rand.Rand`, keeping random number
  generation inside the game package
- Test-only helpers moved out of `Board`'s public interface
- The help overlay is generated from the key map, so bindings have one definition
- Errors are reported on stderr with a non-zero exit code instead of `log.Fatal`

## [0.1.0] - 2026-08-05

### Added

- Full-featured terminal Minesweeper TUI (bubbletea + lipgloss)
- Game core with first-click safety, flood fill, and chord
- Difficulty presets: Beginner, Intermediate, Expert, Custom
- Timer and per-difficulty high scores
- Keyboard and mouse controls
