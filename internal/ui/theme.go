package ui

// Theme names a colour palette. The board is readable without colour at all,
// so a theme only changes how pleasant it looks — never what it means.
type Theme string

const (
	// ThemeClassic is the light grey of the original Windows game.
	ThemeClassic Theme = "classic"
	// ThemeDark suits dark terminals, where the classic light cells glare.
	ThemeDark Theme = "dark"
	// ThemeColorblind swaps the red/green digits for the Okabe-Ito palette,
	// which stays distinguishable with the common forms of colour blindness.
	ThemeColorblind Theme = "colorblind"
)

// Themes lists every theme in presentation order.
var Themes = []Theme{ThemeClassic, ThemeDark, ThemeColorblind}

// ParseTheme resolves a theme name, reporting whether it was recognised. The
// caller decides what an unknown name means: a typo on the command line is
// worth an error, whereas a stale config file should just fall back.
func ParseTheme(s string) (Theme, bool) {
	for _, t := range Themes {
		if string(t) == s {
			return t, true
		}
	}
	return ThemeClassic, false
}

// palette is every colour a theme supplies. Collecting them in one struct means
// adding a theme is filling in a table, not editing the rendering code.
type palette struct {
	hiddenFg   string
	hiddenBg   string
	revealedBg string
	flagFg     string
	questionFg string
	mineFg     string
	cursorFg   string
	cursorBg   string
	hudFg      string
	hudBg      string
	statusFg   string
	overlayFg  string
	overlayBg  string
	titleFg    string
	warningFg  string
	// digits holds the colour of each adjacency count, indexed 1 through 8.
	digits [9]string
}

func (t Theme) palette() palette {
	switch t {
	case ThemeDark:
		return darkPalette
	case ThemeColorblind:
		return colorblindPalette
	default:
		return classicPalette
	}
}

var classicPalette = palette{
	hiddenFg:   "#8C8C8C",
	hiddenBg:   "#C0C0C0",
	revealedBg: "#E0E0E0",
	flagFg:     "#CC0000",
	questionFg: "#0000CC",
	mineFg:     "#CC0000",
	cursorFg:   "#FFFFFF",
	cursorBg:   "#000080",
	hudFg:      "#FFFFFF",
	hudBg:      "#606060",
	statusFg:   "#909090",
	overlayFg:  "#FFFFFF",
	overlayBg:  "#1C1C1C",
	titleFg:    "#FFD700",
	warningFg:  "#FF8800",
	digits: [9]string{
		1: "#0000FF", 2: "#008000", 3: "#FF0000", 4: "#000080",
		5: "#800000", 6: "#008080", 7: "#000000", 8: "#808080",
	},
}

var darkPalette = palette{
	hiddenFg:   "#5A5A66",
	hiddenBg:   "#2E2E36",
	revealedBg: "#1B1B21",
	flagFg:     "#FF6B6B",
	questionFg: "#7AA2F7",
	mineFg:     "#FF6B6B",
	cursorFg:   "#101014",
	cursorBg:   "#7AA2F7",
	hudFg:      "#E6E6EA",
	hudBg:      "#3A3A44",
	statusFg:   "#71717A",
	overlayFg:  "#E6E6EA",
	overlayBg:  "#26262E",
	titleFg:    "#F2C97D",
	warningFg:  "#F2994A",
	digits: [9]string{
		1: "#7AA2F7", 2: "#9ECE6A", 3: "#F7768E", 4: "#BB9AF7",
		5: "#E0AF68", 6: "#2AC3DE", 7: "#C0CAF5", 8: "#9AA5CE",
	},
}

// colorblindPalette uses the Okabe-Ito qualitative palette, chosen so that no
// two digits collapse into the same colour under deuteranopia or protanopia.
var colorblindPalette = palette{
	hiddenFg:   "#8C8C8C",
	hiddenBg:   "#C0C0C0",
	revealedBg: "#E8E8E8",
	flagFg:     "#D55E00",
	questionFg: "#0072B2",
	mineFg:     "#D55E00",
	cursorFg:   "#FFFFFF",
	cursorBg:   "#0072B2",
	hudFg:      "#FFFFFF",
	hudBg:      "#4D4D4D",
	statusFg:   "#909090",
	overlayFg:  "#FFFFFF",
	overlayBg:  "#1C1C1C",
	titleFg:    "#E69F00",
	warningFg:  "#D55E00",
	digits: [9]string{
		1: "#0072B2", 2: "#009E73", 3: "#D55E00", 4: "#CC79A7",
		5: "#56B4E9", 6: "#8C6D1F", 7: "#000000", 8: "#767676",
	},
}
