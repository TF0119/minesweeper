package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds lipgloss styles for rendering.
type Styles struct {
	Hidden    lipgloss.Style
	Revealed  lipgloss.Style
	Flagged   lipgloss.Style
	Mine      lipgloss.Style
	Cursor    lipgloss.Style
	Digits    [9]lipgloss.Style
	HUD       lipgloss.Style
	StatusBar lipgloss.Style
	Overlay   lipgloss.Style
	Title     lipgloss.Style
	Warning   lipgloss.Style
}

// NewStyles builds color or monochrome styles.
func NewStyles(useColor bool) Styles {
	if !useColor {
		base := lipgloss.NewStyle()
		return Styles{
			Hidden:    base.Copy().Bold(true),
			Revealed:  base,
			Flagged:   base.Copy().Bold(true),
			Mine:      base.Copy().Bold(true),
			Cursor:    base.Copy().Reverse(true).Bold(true),
			HUD:       base.Copy().Bold(true),
			StatusBar: base,
			Overlay:   base.Copy().Reverse(true),
			Title:     base.Copy().Bold(true),
			Warning:   base.Copy().Bold(true),
		}
	}

	hidden := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#C0C0C0")).
		Bold(true)
	revealed := lipgloss.NewStyle().
		Background(lipgloss.Color("#D0D0D0"))

	digitColors := [9]string{
		"",
		"#0000FF",
		"#008000",
		"#FF0000",
		"#000080",
		"#800000",
		"#008080",
		"#000000",
		"#808080",
	}
	var digits [9]lipgloss.Style
	for i := 1; i <= 8; i++ {
		digits[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(digitColors[i])).
			Background(lipgloss.Color("#D0D0D0")).
			Bold(true)
	}

	return Styles{
		Hidden:    hidden,
		Revealed:  revealed,
		Flagged:   hidden.Copy().Foreground(lipgloss.Color("#FF0000")),
		Mine:      revealed.Copy().Foreground(lipgloss.Color("#FF0000")).Bold(true),
		Cursor:    lipgloss.NewStyle().Background(lipgloss.Color("#000080")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		Digits:    digits,
		HUD:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#808080")).Padding(0, 1),
		StatusBar: lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		Overlay:   lipgloss.NewStyle().Background(lipgloss.Color("#000000")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1, 2),
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")).Bold(true),
	}
}
