// Package game implements Minesweeper rules as a deep, TUI-independent module.
//
// Invariants:
//   - Before the first Reveal, mines are not placed on the board.
//   - Flags may only be placed on hidden (unrevealed) cells.
//   - After Won or Lost, all mutating operations are no-ops.
package game
