package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/storage"
)

func TestParseTheme(t *testing.T) {
	for _, want := range Themes {
		got, ok := ParseTheme(string(want))
		if !ok || got != want {
			t.Errorf("ParseTheme(%q) = %v, %v", want, got, ok)
		}
	}

	got, ok := ParseTheme("chartreuse")
	if ok {
		t.Error("an unknown name should not be reported as recognised")
	}
	if got != ThemeClassic {
		t.Errorf("fallback = %v, want %v", got, ThemeClassic)
	}
}

// Two digits sharing a colour would make the board harder to read than having
// no colour at all, which defeats the point of a theme.
func TestThemeDigitColorsAreDistinct(t *testing.T) {
	for _, theme := range Themes {
		t.Run(string(theme), func(t *testing.T) {
			p := theme.palette()
			seen := make(map[string]int, 8)
			for i := 1; i <= 8; i++ {
				c := p.digits[i]
				if c == "" {
					t.Errorf("digit %d has no colour", i)
					continue
				}
				if prev, dup := seen[c]; dup {
					t.Errorf("digits %d and %d share the colour %s", prev, i, c)
				}
				seen[c] = i
			}
		})
	}
}

// Hidden and revealed cells are told apart by their glyph, but the shades
// should not fight that by looking identical.
func TestThemeSeparatesHiddenFromRevealed(t *testing.T) {
	for _, theme := range Themes {
		p := theme.palette()
		if p.hiddenBg == p.revealedBg {
			t.Errorf("%s: hidden and revealed share the background %s", theme, p.hiddenBg)
		}
		if p.flagFg == p.questionFg {
			t.Errorf("%s: flags and question marks share the colour %s", theme, p.flagFg)
		}
	}
}

func TestThemeFromConfigFallsBackInsteadOfFailing(t *testing.T) {
	c := storage.DefaultConfig()
	c.Theme = "written-by-a-newer-version"
	if got := themeFromConfig(c); got != ThemeClassic {
		t.Errorf("themeFromConfig() = %v, want %v", got, ThemeClassic)
	}

	c.Theme = string(ThemeDark)
	if got := themeFromConfig(c); got != ThemeDark {
		t.Errorf("themeFromConfig() = %v, want %v", got, ThemeDark)
	}
}
