// Package game implements Minesweeper rules as a deep, TUI-independent module.
//
// Invariants:
//   - Before the first Reveal, mines are not placed on the board.
//   - Only hidden cells carry a mark; revealing a cell clears it.
//   - A flag blocks revealing; a question mark never changes what an action does.
//   - After Won or Lost, all mutating operations are no-ops.
package game
