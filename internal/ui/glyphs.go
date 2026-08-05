package ui

// glyphs holds the characters used to draw cells. Resolving the character set
// once at construction keeps the per-cell render path branch-free.
type glyphs struct {
	hidden   string
	flag     string
	question string
	mine     string
}

// newGlyphs picks a glyph set. Hidden cells always carry a visible mark rather
// than relying on their background shade, so the board stays readable with
// -no-color, on low-contrast themes, and in terminals that drop colour.
// Both sets are single-width to avoid the column misalignment wide emoji cause.
func newGlyphs(useEmoji bool) glyphs {
	if useEmoji {
		return glyphs{hidden: "·", flag: "⚑", question: "?", mine: "✹"}
	}
	return glyphs{hidden: ".", flag: "F", question: "?", mine: "*"}
}
