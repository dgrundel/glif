package render

import "github.com/dgrundel/glif/grid"

type Sprite struct {
	W           int
	H           int
	Cells       []grid.Cell
	Transparent rune
}

func (s *Sprite) cellAt(x, y int) grid.Cell {
	return s.Cells[y*s.W+x]
}
