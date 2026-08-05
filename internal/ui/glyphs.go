package ui

// glyphs holds the characters used to draw cells. Resolving the emoji setting
// once at construction keeps the per-cell render path branch-free.
type glyphs struct {
	hidden string
	flag   string
	mine   string
}

// newGlyphs picks a glyph set. Both sets are single-width to avoid the column
// misalignment that wide emoji cause in many terminals.
func newGlyphs(useEmoji bool) glyphs {
	if useEmoji {
		return glyphs{hidden: "·", flag: "⚑", mine: "✹"}
	}
	return glyphs{hidden: " ", flag: "F", mine: "*"}
}
