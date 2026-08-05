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
| Storage | `internal/storage` | JSON, XDG paths | Config and high scores |

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

## Errors defined out of existence

Board actions return `ActionResult{Ok: false}` rather than an error when they do
nothing. Clicking a revealed cell is not a failure; it is a no-op, and the UI
should not need an error branch for it. `CanReveal`, `CanFlag`, and `CanChord`
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
- Flags may only be placed on hidden cells.
- After `Won` or `Lost`, mutating operations are no-ops.

## Deliberately out of scope

Solver hints, replays, and network play would each be a new source of complexity
in a package that is currently easy to hold in your head. If they are added, they
belong beside `internal/game` and should depend on it, not modify it.
