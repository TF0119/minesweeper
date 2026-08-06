# Changelog

All notable changes to this project will be documented in this file.

## [0.4.2] - 2026-08-06

### Fixed

- Watching a replay of another difficulty left the live board drawn at the
  replay's size with the cursor outside it, so keys stopped doing anything.
  The viewport and cursor are now handed back when playback ends
- Resizing the terminal during a timelapse clipped the replay board to the live
  board's size, and the scroll indicator described the wrong board
- The timelapse HUD showed the live board's mine count, time, and seed instead
  of the replay's
- `q` on the statistics and help screens quit the game instead of closing the
  screen, unlike every other sub-screen
- Opening the menu from statistics or help and picking Resume returned to that
  screen instead of the board
- Replays cycled marks as if question marks were always on, so a flag that had
  been cleared during play came back as `?` during playback
- `+` and `-` during a timelapse only took effect on the next frame
- Actions that changed nothing were recorded, spending timelapse frames
  showing nothing happen
- The difficulty menu highlighted Beginner while a custom board was in play; it
  now names the custom board instead of implying a preset

## [0.4.1] - 2026-08-06

### Fixed

- Menu navigation after Watch: nested screens now unwind correctly instead of
  trapping you on the hub after a timelapse
- Game over and victory overlays now accept `m` / `Esc` to open the menu
- `q` exits timelapse playback back to the replay list, same as `Esc`

## [0.4.0] - 2026-08-06

### Added

- In-game menu on `m` or `Esc`: new game, daily challenge, difficulty,
  statistics, settings, help, and quit. Launch still drops you straight into a
  board — the menu never blocks startup
- Settings screen in the menu: theme, no-guess, question marks, and emoji glyphs,
  saved to config immediately
- Timelapse watch: finished games are recorded and listed under Menu → Watch.
  Moves play back automatically; `Space` pauses, `+`/`-` adjust speed, `r`
  restarts. Distinct from `r` during play, which restarts the same seed for
  another attempt

### Changed

- Status bar hints now point at `m menu` instead of listing every shortcut

## [0.3.0] - 2026-08-05

### Added

- No-guess boards via `-no-guess`: a logic solver checks each candidate layout
  and the generator retries until it finds one that can be cleared by deduction
  alone. Only 13 of 200 random Expert boards pass that bar, so this is the
  difference between finishing a game and flipping a coin. The status line says
  `guess needed` when the search comes up empty
- Question marks: `f` now cycles a cell through flag, `?`, and clear. A `?` is a
  note only — it does not block a reveal and does not count towards a chord.
  Set `"question_marks": false` in the config for a plain flag toggle
- Statistics on `s`: games played, win rate, average winning time, and current
  and best streak, tracked per difficulty in `stats.json`. Only finished games
  count
- Themes via `-theme`: `classic`, `dark`, and `colorblind`. The colourblind
  palette is Okabe-Ito, chosen so no two adjacency digits share a colour under
  the common forms of colour blindness
- Homebrew tap and Scoop bucket, published by GoReleaser on release

### Changed

- The status bar now shrinks to fit the terminal instead of wrapping and pushing
  the board off screen

### Fixed

- Right-click flagged a cell on both press and release, so the mark cycled twice
  and `F` flashed for a moment. Right-click now acts on release only; legacy
  X10 terminals that emit a single press event still work
- Homebrew install failed on Linux because the cask postflight called macOS-only
  `xattr`. The hook now runs only on macOS
- `minesweeper -version` reported `dev` for binaries installed with `go install`.
  It now falls back to the module version the toolchain records
- A config file written before a setting existed no longer silently reads that
  setting as its zero value

## [0.2.1] - 2026-08-05

### Fixed

- Hidden cells were drawn as blank space and told apart from empty revealed
  cells by their background shade alone, which made `-no-color`, `NO_COLOR=1`,
  and colourless terminals unplayable. Hidden cells now carry a visible glyph,
  and `NO_COLOR`/`TERM=dumb` switch to the monochrome styles so the cursor keeps
  its highlight

### Added

- Demo recording in the README, generated from `docs/demo.tape` with VHS
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, and issue and pull
  request templates
- `Makefile` with the same checks CI runs, plus `.editorconfig`
- golangci-lint in CI, weekly Dependabot updates, and read-only workflow
  permissions

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
