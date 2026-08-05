# Minesweeper CLI — Design Document

## Overview

Terminal Minesweeper with a testable game core separated from the TUI.
Stack: Go, bubbletea, lipgloss.

## Layer Architecture (APoSD)

| Layer | Package | Vocabulary | Responsibility |
|-------|---------|------------|----------------|
| CLI | `cmd/minesweeper` | flags, exit codes | Parse args, invoke UI |
| UI | `internal/ui` | tea.Msg, lipgloss | Input, rendering, timer |
| Game | `internal/game` | Coord, Board, ActionResult | Rules, state machine |
| Storage | `internal/storage` | JSON, XDG paths | Config, high scores |

**Principle**: Complexity lives in `internal/game`. UI never sees mine positions directly—only `CellView`.

## First-Click Safety — Design It Twice

| Option | Summary | Pros | Cons | Chosen |
|--------|---------|------|------|--------|
| A: Deferred generation | Mines placed on first Reveal | Trivially safe | First Reveal does heavy work | |
| B: Relocation | Place all mines, swap mines out of safe zone | Simple generator | Swap algorithm needed | **Yes** |

### Algorithm B (chosen)

1. On first `Reveal(c)`, define safe zone `S = {c} ∪ neighbors(c)`.
2. For each mine in `S`, swap with a non-mine cell outside `S`.
3. Recompute adjacent counts after swaps.
4. Reject boards where `mines > width*height - 9` via `Difficulty.Validate()`.

## Invariants

- Before first Reveal: `minesPlaced == false`.
- During play: flags only on hidden cells.
- After Won/Lost: all actions are no-op (`Status != Playing`).

## Configuration Priority

```
defaults → ~/.config/minesweeper/config.json → CLI flags
```

No environment variables.

## Action API (Define Errors Out of Existence)

Invalid operations return `ActionResult{Ok: false}` rather than errors.
Callers use `CanReveal`, `CanFlag`, `CanChord` for UI affordances.
