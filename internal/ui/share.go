package ui

import (
	"fmt"
	"strings"

	"github.com/TF0119/minesweeper/internal/game"
)

// repoURL is where players land when they share a result. Putting it on the
// card turns a win into a discovery path for the project.
const repoURL = "https://github.com/TF0119/minesweeper"

// buildShareCard formats a copy-pasteable result for the current win. Daily
// boards lead with the date so two players can compare; other boards lead with
// the seed so a friend can replay the same layout.
func (m Model) buildShareCard() string {
	var head string
	if m.isDailyBoard() {
		head = "minesweeper · daily " + game.DailyDate(timeNow())
	} else {
		head = "minesweeper · seed " + m.board.Seed().String()
	}

	parts := []string{m.difficultyLabel()}
	if m.boardNoGuess {
		parts = append(parts, "no-guess")
	}
	parts = append(parts, fmt.Sprintf("%ds", m.elapsed))

	return head + "\n" + strings.Join(parts, " · ") + "\n" + repoURL
}

// difficultyLabel is the human name used on the share card. Custom boards
// include size and mine count so the line stays meaningful without a preset.
func (m Model) difficultyLabel() string {
	if m.difficulty.Preset == game.Custom {
		return fmt.Sprintf("custom %dx%d/%d",
			m.difficulty.Width, m.difficulty.Height, m.difficulty.Mines)
	}
	return m.difficulty.Preset.String()
}

// ShareCard returns the result printed after the TUI exits, or empty when
// there is nothing worth pasting (no win, or a newer board already started).
func (m Model) ShareCard() string {
	if !m.quitting {
		return ""
	}
	return m.shareCard
}
