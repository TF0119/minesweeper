package game

// Coord is a cell position on the board (0-indexed).
type Coord struct {
	X, Y int
}

// InBounds reports whether c is within a board of size w×h.
func (c Coord) InBounds(w, h int) bool {
	return c.X >= 0 && c.X < w && c.Y >= 0 && c.Y < h
}

// Neighbors returns up to eight adjacent coordinates, clipped to bounds.
func (c Coord) Neighbors(w, h int) []Coord {
	out := make([]Coord, 0, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nc := Coord{X: c.X + dx, Y: c.Y + dy}
			if nc.InBounds(w, h) {
				out = append(out, nc)
			}
		}
	}
	return out
}

func (c Coord) index(w int) int {
	return c.Y*w + c.X
}
