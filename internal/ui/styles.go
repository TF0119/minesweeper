package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds lipgloss styles for rendering.
type Styles struct {
	Hidden     lipgloss.Style
	Revealed   lipgloss.Style
	Flagged    lipgloss.Style
	Questioned lipgloss.Style
	Mine       lipgloss.Style
	Cursor     lipgloss.Style
	Digits     [9]lipgloss.Style
	HUD        lipgloss.Style
	StatusBar  lipgloss.Style
	Overlay    lipgloss.Style
	Title      lipgloss.Style
	Warning    lipgloss.Style
}

// digitColors are the classic Minesweeper number colours, indexed by the
// adjacent mine count.
var digitColors = [9]string{
	1: "#0000FF",
	2: "#008000",
	3: "#FF0000",
	4: "#000080",
	5: "#800000",
	6: "#008080",
	7: "#000000",
	8: "#808080",
}

// NewStyles builds the style set. The monochrome set carries cell state
// through bold and reverse so the board survives terminals without colour.
func NewStyles(useColor bool) Styles {
	if !useColor {
		return monochromeStyles(lipgloss.DefaultRenderer())
	}
	return colorStyles(lipgloss.DefaultRenderer())
}

func monochromeStyles(r *lipgloss.Renderer) Styles {
	base := r.NewStyle()
	bold := base.Bold(true)
	return Styles{
		Hidden:     base,
		Revealed:   base,
		Flagged:    bold,
		Questioned: bold,
		Mine:       bold,
		Cursor:     base.Reverse(true),
		Digits:     [9]lipgloss.Style{1: bold, 2: bold, 3: bold, 4: bold, 5: bold, 6: bold, 7: bold, 8: bold},
		HUD:        bold,
		StatusBar:  base,
		Overlay:    base.Reverse(true).Padding(1, 2),
		Title:      bold,
		Warning:    bold,
	}
}

func colorStyles(r *lipgloss.Renderer) Styles {
	hidden := r.NewStyle().
		Foreground(lipgloss.Color("#8C8C8C")).
		Background(lipgloss.Color("#C0C0C0"))
	revealed := r.NewStyle().
		Background(lipgloss.Color("#E0E0E0"))

	var digits [9]lipgloss.Style
	for i := 1; i <= 8; i++ {
		digits[i] = revealed.
			Foreground(lipgloss.Color(digitColors[i])).
			Bold(true)
	}

	return Styles{
		Hidden:     hidden,
		Revealed:   revealed,
		Flagged:    hidden.Foreground(lipgloss.Color("#CC0000")).Bold(true),
		Questioned: hidden.Foreground(lipgloss.Color("#0000CC")).Bold(true),
		Mine:       revealed.Foreground(lipgloss.Color("#CC0000")).Bold(true),
		Cursor:     r.NewStyle().Background(lipgloss.Color("#000080")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		Digits:     digits,
		HUD:        r.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#606060")),
		StatusBar:  r.NewStyle().Foreground(lipgloss.Color("#909090")),
		Overlay:    r.NewStyle().Background(lipgloss.Color("#1C1C1C")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1, 2),
		Title:      r.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700")),
		Warning:    r.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8800")),
	}
}
