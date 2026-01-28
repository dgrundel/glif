package assets

import (
	"strings"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/render"
)

func SpriteFromASCII(src string, style grid.Style, transparent rune) *render.Sprite {
	src = strings.Trim(src, "\n")
	lines := strings.Split(src, "\n")
	w := 0
	for _, line := range lines {
		if len(line) > w {
			w = len(line)
		}
	}
	if w == 0 {
		return &render.Sprite{W: 0, H: 0}
	}
	cells := make([]grid.Cell, w*len(lines))
	for y, line := range lines {
		for x := 0; x < w; x++ {
			ch := rune(' ')
			if x < len(line) {
				ch = rune(line[x])
			}
			cells[y*w+x] = grid.Cell{Ch: ch, Style: style}
		}
	}
	return &render.Sprite{W: w, H: len(lines), Cells: cells, Transparent: transparent}
}
