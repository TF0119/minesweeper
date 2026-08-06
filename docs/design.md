# Minesweeper CLI — Design Document

This file records the decisions behind the structure, following the ideas in
*A Philosophy of Software Design*. It is meant to explain why the code looks the
way it does, not to restate what the code already says.

## Layer architecture

| Layer | Package | Vocabulary | Responsibility |
|-------|---------|------------|----------------|
| CLI | `cmd/minesweeper` | flags, exit codes | Parse arguments, resolve options |
| UI | `internal/ui` | `tea.Msg`, lipgloss styles | Input, rendering, timer |
| Game | `internal/game` | `Coord`, `Board`, `ActionResult` | Rules and state machine |
| Storage | `internal/storage` | JSON, XDG paths | Config, high scores, statistics |

Complexity lives in `internal/game`. The UI never sees mine positions; it reads
`CellView`, which exposes only what a player is allowed to know. That is what
makes it safe for the renderer to be simple.

## First-click safety: designed twice

| Option | Summary | Pros | Cons | Chosen |
|--------|---------|------|------|--------|
| A: deferred generation | Place mines during the first `Reveal` | Trivially safe | Ties layout to reveal logic | |
| B: relocation | Place all mines, then move any inside the safe zone | Generator stays independent of reveal | Needs a swap step | Yes |

Algorithm B, as implemented:

1. On the first `Reveal(c)`, the safe zone is `{c} ∪ neighbors(c)`.
2. Each mine inside the zone swaps places with a non-mine cell outside it.
3. Adjacency counts are recomputed after the swaps.
4. `Difficulty.Validate` rejects boards where `mines > width*height - 9`, so step
   2 always has somewhere to move mines to.

## Seeds

`NewBoard` takes a `Seed`, not a `*rand.Rand`. Callers say *which board* they
want; how randomness is produced stays inside the package. This is what let the
CLI, the UI, and the tests all move to seeds without any of them importing
`math/rand`.

`Seed` is a `uint32` rather than an `int64` because its purpose is to be shared.
Ten digits can be read aloud or pasted into a chat message; twenty cannot.

The daily challenge is `FNV-1a` over the UTC date string. A hash of the date,
rather than the date's numeric value, keeps consecutive days from producing
neighbouring seeds.

A seed pins the mine layout but not the first click, since relocation depends on
where the player opens. Two players on the same daily seed can therefore see
different boards if they start in different places — an acceptable trade for
keeping the opening move safe.

## No-guess boards

A classic board can force a coin flip, and losing to one is the least satisfying
way for a game to end. `internal/game/solver.go` decides whether a layout can be
cleared by reasoning alone, and `NewNoGuessBoard` keeps generating until the
answer is yes.

The solver applies three rules until nothing more can be learned:

1. A number whose remaining count equals its hidden neighbours marks them all as
   mines; a number already satisfied by its flags clears them all.
2. When one number's cells are a subset of another's, the difference between the
   two holds the difference between their counts. This is the reasoning behind
   patterns like 1-2-1.
3. The mine counter is a constraint over every hidden cell, which is how most
   endgames close.

The solver stops at what a player can work out, not at what is theoretically
decidable. A layout that only yields to exhaustive enumeration is one a human
would guess on, so calling it unsolvable is the honest answer.

Two design consequences:

- `deduce` is a pure function over constraints, so the reasoning is tested
  directly rather than through generated boards.
- The classic generator's use of the random source is frozen. No-guess
  generation draws from the same source but only on its own path, because
  changing the classic draw order would hand every previously shared seed a
  different board.

The search is bounded. When it runs out of attempts the player gets an ordinary
board and the status line says `guess needed`: a quietly broken promise is worse
than a stated one.

## Marks

A cell carries at most one mark, so `Mark` is an enum rather than a `Flagged`
and `Questioned` pair. The pair would have a fourth state that means nothing,
and every reader would have to work out that it cannot happen.

A flag changes the rules — it blocks reveals and counts towards chords. A
question mark deliberately changes nothing at all; it is a note, and revealing
the cell clears it. Keeping `?` inert is what makes it safe to add: no rule
elsewhere in the package needs to know about it.

Whether `f` cycles through `?` is a preference, so it is a parameter of
`CycleMark` rather than state on the board. The board defines what marks mean;
the UI decides which ones a player can reach.

## Statistics

`Tally` stores counts and nothing else. Win rate and average time are computed
on read, because a stored aggregate that disagrees with the counts it summarises
is a bug that survives every save. Averages cover won games only: a game lost to
a mine says nothing about pace.

A game is counted when it reaches a result. Abandoning a board never appears as
a loss, which keeps the numbers honest enough to be worth looking at.

## Themes

`internal/ui/theme.go` holds a table of palettes; `styles.go` builds lipgloss
styles from whichever one is selected. Adding a theme means adding a row, not
editing rendering code.

Colour is never the only carrier of state. Hidden cells have their own glyph and
terminals without colour get a separate monochrome style set, so a theme only
changes how the board looks, never what it means. The colourblind palette is
Okabe-Ito, and a test asserts that no two adjacency digits share a colour.

An unknown theme name behaves differently depending on where it came from: a
typo on the command line is an error, while a name in the config file falls back
to `classic`. A config file written by a newer version should not stop the game
from starting.

## Menu

The game opens directly onto the board. A title screen would add a step between
the player and Minesweeper for no gain. The menu is opt-in via `m` or `Esc` and
collects settings, statistics, and navigation that used to be scattered across
single-key shortcuts.

`screenStack` remembers nested screens so `Esc` unwinds Menu → Watch →
timelapse without trapping the player on the hub.

## Timelapse

`r` during play restarts the same seed — a new attempt at the same layout.
**Watch** replays a recorded game move by move. The two must stay separate in
labelling and docs or players will expect the wrong thing.

Recordings live in `internal/game` as `Replay` (seed + moves). The UI appends
moves during play and `storage` persists them on disk. Playback rebuilds a board
from the seed and applies moves in order; timelapse is just that loop on a timer
in the UI layer, not logic in the game package.

## Resuming an unfinished game

Quitting mid-game writes `session.json`; the next launch replays it and puts the
player back on the same board. The file stores the seed and the moves rather
than the board, which is the same trick recordings use: one description of what
happened instead of two that can drift apart. It also means the restore path is
`Replay.Apply`, already exercised by the timelapse.

Two rules keep a stale game from surprising anyone. Finishing clears the file,
so a won board never comes back. Quitting with nothing to resume also clears it,
or starting a fresh board and closing it would resurrect the game before that.

The clock counts time spent playing, so a resumed game continues from the saved
figure rather than restarting or charging the player for the hours it was
closed. Naming a board with `-seed`, `-daily`, or `-difficulty` skips the saved
game: an explicit request should win over an implicit one.

## Errors defined out of existence

Board actions return `ActionResult{Ok: false}` rather than an error when they do
nothing. Clicking a revealed cell is not a failure; it is a no-op, and the UI
should not need an error branch for it. `CanReveal`, `CanMark`, and `CanChord`
exist for callers that want to know in advance, for example to grey out an
affordance.

Errors are reserved for situations the caller genuinely cannot proceed from, such
as an unparseable seed or an impossible difficulty.

## Viewport

Boards can be wider than the terminal, and `internal/ui/viewport.go` is where all
the arithmetic for that lives: sizing the window, scrolling to follow the cursor,
and mapping terminal coordinates back to board coordinates. Rendering and mouse
handling each call one small method rather than repeating offset maths.

`toBoard` owns the knowledge that row 0 is the HUD. When the layout changed
earlier, the mouse offset was wrong in a second place; concentrating it here is
what prevents that class of bug from returning.

## Timer

The timer is derived from a wall-clock start time, not accumulated from ticks. An
earlier version incremented a counter on each tick and rescheduled unconditionally,
which left a tick chain alive after a game ended; the next game then received
several ticks per second. Ticks now only drive redraws, and stopping the timer
stops the chain.

## Configuration

One order, applied once at startup:

```
defaults → ~/.config/minesweeper/config.json → command-line flags
```

There are no environment variables. Every additional source of configuration is
another thing a reader has to check before they can predict behaviour.

## Invariants

- Before the first `Reveal`, no mines are placed.
- Only hidden cells carry a mark, and revealing a cell clears it.
- A flag blocks revealing; a question mark never changes what an action does.
- After `Won` or `Lost`, mutating operations are no-ops.

## Deliberately out of scope

In-game hints and network play would each be a new source of complexity in a
package that is currently easy to hold in your head. The solver exists to
generate boards, not to play them; wiring it to a hint key would put the answer
one keystroke away from every position and change what the game is.

If they are added, they belong beside `internal/game` and should depend on it,
not modify it.
