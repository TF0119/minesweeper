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

// NewStyles builds the style set for a theme. The monochrome set carries cell
// state through bold and reverse so the board survives terminals without
// colour, which is also why no theme is consulted in that case.
func NewStyles(theme Theme, useColor bool) Styles {
	r := lipgloss.DefaultRenderer()
	if !useColor {
		return monochromeStyles(r)
	}
	return colorStyles(r, theme.palette())
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

// color converts a palette entry into a lipgloss colour. It exists only to keep
// the style table below readable.
func color(hex string) lipgloss.Color { return lipgloss.Color(hex) }

func colorStyles(r *lipgloss.Renderer, p palette) Styles {
	hidden := r.NewStyle().
		Foreground(color(p.hiddenFg)).
		Background(color(p.hiddenBg))
	revealed := r.NewStyle().Background(color(p.revealedBg))

	var digits [9]lipgloss.Style
	for i := 1; i <= 8; i++ {
		digits[i] = revealed.Foreground(color(p.digits[i])).Bold(true)
	}

	return Styles{
		Hidden:     hidden,
		Revealed:   revealed,
		Flagged:    hidden.Foreground(color(p.flagFg)).Bold(true),
		Questioned: hidden.Foreground(color(p.questionFg)).Bold(true),
		Mine:       revealed.Foreground(color(p.mineFg)).Bold(true),
		Cursor:     r.NewStyle().Background(color(p.cursorBg)).Foreground(color(p.cursorFg)).Bold(true),
		Digits:     digits,
		HUD:        r.NewStyle().Bold(true).Foreground(color(p.hudFg)).Background(color(p.hudBg)),
		StatusBar:  r.NewStyle().Foreground(color(p.statusFg)),
		Overlay:    r.NewStyle().Background(color(p.overlayBg)).Foreground(color(p.overlayFg)).Padding(1, 2),
		Title:      r.NewStyle().Bold(true).Foreground(color(p.titleFg)),
		Warning:    r.NewStyle().Bold(true).Foreground(color(p.warningFg)),
	}
}
