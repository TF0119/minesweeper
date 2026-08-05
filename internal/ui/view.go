package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/takeru0119/minesweeper/internal/game"
)

const cellWidth = 3

// Screen identifies the active UI screen.
type Screen int

const (
	ScreenPlaying Screen = iota
	ScreenDifficultyMenu
	ScreenHelp
	ScreenGameOver
	ScreenWin
)

func (m Model) renderCell(x, y int) string {
	c := game.Coord{X: x, Y: y}
	view := m.board.CellView(c)
	isCursor := m.cursor.X == x && m.cursor.Y == y && m.screen == ScreenPlaying

	var content string
	switch view.State {
	case game.CellHidden:
		content = m.hiddenChar()
	case game.CellFlagged:
		content = m.flagChar()
	case game.CellRevealed:
		if view.ShowMine {
			content = m.mineChar()
		} else if view.Adjacent == 0 {
			content = " "
		} else {
			content = fmt.Sprintf("%d", view.Adjacent)
		}
	}

	text := centerCell(content, cellWidth)
	var styled string
	switch {
	case isCursor:
		styled = m.styles.Cursor.Render(text)
	case view.State == game.CellFlagged:
		styled = m.styles.Flagged.Render(text)
	case view.State == game.CellRevealed:
		if view.ShowMine {
			styled = m.styles.Revealed.Copy().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render(text)
		} else if view.Adjacent > 0 {
			styled = m.styles.Digits[view.Adjacent].Render(text)
		} else {
			styled = m.styles.Revealed.Render(text)
		}
	default:
		styled = m.styles.Hidden.Render(text)
	}
	return styled
}

func centerCell(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (m Model) hiddenChar() string {
	if m.config.UseEmoji {
		return "·"
	}
	return " "
}

func (m Model) flagChar() string {
	if m.config.UseEmoji {
		return "F"
	}
	return "F"
}

func (m Model) mineChar() string {
	if m.config.UseEmoji {
		return "*"
	}
	return "*"
}

func (m Model) renderBoard() string {
	var rows []string
	for y := 0; y < m.board.Height(); y++ {
		var cells []string
		for x := 0; x < m.board.Width(); x++ {
			cells = append(cells, m.renderCell(x, y))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderHUD() string {
	key := m.difficulty.Key()
	best := m.highscores.Best(key)
	bestStr := "---"
	if best >= 0 {
		bestStr = fmt.Sprintf("%03d", best)
	}
	mines := fmt.Sprintf("%03d", m.board.RemainingMines())
	time := fmt.Sprintf("%03d", m.elapsed)
	preset := m.difficulty.Preset.String()
	if m.difficulty.Preset == game.Custom {
		preset = fmt.Sprintf("custom %dx%d", m.difficulty.Width, m.difficulty.Height)
	}
	line := fmt.Sprintf(" Mines:%s  Time:%s  Best:%s  [%s] ", mines, time, bestStr, preset)
	return m.styles.HUD.Render(line)
}

func (m Model) renderStatusBar() string {
	return m.styles.StatusBar.Render(" ↑↓←→/hjkl move · Space reveal · f flag · c chord · n new · d difficulty · ? help · q quit ")
}

func (m Model) renderOverlay(title, body string) string {
	content := m.styles.Title.Render(title) + "\n\n" + body
	return m.styles.Overlay.Render(content)
}

func (m Model) renderDifficultyMenu() string {
	options := []struct {
		p    game.Preset
		label string
	}{
		{game.Beginner, "Beginner (9×9, 10 mines)"},
		{game.Intermediate, "Intermediate (16×16, 40 mines)"},
		{game.Expert, "Expert (30×16, 99 mines)"},
	}
	var lines []string
	for i, opt := range options {
		prefix := "  "
		if i == m.menuIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+opt.label)
	}
	lines = append(lines, "", "Enter: select  Esc: back")
	return m.renderOverlay("Difficulty", strings.Join(lines, "\n"))
}

func (m Model) renderHelp() string {
	body := `↑↓←→ / hjkl     Move cursor
Space / Enter    Reveal cell
f                Toggle flag
c                Chord (auto-reveal)
n                New game
d                Difficulty menu
?                This help
q / Ctrl+C       Quit

Mouse: left=reveal, right=flag`
	return m.renderOverlay("Help", body)
}

func (m Model) renderGameOver() string {
	return m.renderOverlay("Game Over", fmt.Sprintf("You hit a mine!\n\nTime: %03d seconds\n\nPress n for new game", m.elapsed))
}

func (m Model) renderWin() string {
	key := m.difficulty.Key()
	best := m.highscores.Best(key)
	msg := fmt.Sprintf("You win!\n\nTime: %03d seconds", m.elapsed)
	if best >= 0 {
		msg += fmt.Sprintf("\nBest: %03d seconds", best)
	}
	msg += "\n\nPress n for new game"
	return m.renderOverlay("Victory", msg)
}

func (m Model) checkTerminalSize() string {
	needW := m.board.Width()*cellWidth + 4
	needH := m.board.Height() + 6
	if m.width > 0 && m.width < needW {
		return fmt.Sprintf("Terminal width %d < required %d", m.width, needW)
	}
	if m.height > 0 && m.height < needH {
		return fmt.Sprintf("Terminal height %d < required %d", m.height, needH)
	}
	return ""
}

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var parts []string
	parts = append(parts, m.renderHUD())
	parts = append(parts, m.renderBoard())

	if warn := m.checkTerminalSize(); warn != "" {
		parts = append(parts, m.styles.Warning.Render("Warning: "+warn))
	}
	if m.errMsg != "" {
		parts = append(parts, m.styles.Warning.Render(m.errMsg))
	}

	switch m.screen {
	case ScreenDifficultyMenu:
		parts = append(parts, m.renderDifficultyMenu())
	case ScreenHelp:
		parts = append(parts, m.renderHelp())
	case ScreenGameOver:
		parts = append(parts, m.renderGameOver())
	case ScreenWin:
		parts = append(parts, m.renderWin())
	}

	parts = append(parts, m.renderStatusBar())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
