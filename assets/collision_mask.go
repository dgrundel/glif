package assets

import (
	"fmt"

	"github.com/dgrundel/glif/render"
)

func buildCollisionMask(spriteLines, collisionLines [][]rune, widthMask [][]int, cellW, sw, sh int) (*render.CollisionMask, error) {
	if collisionLines == nil {
		return nil, nil
	}
	cw, ch := dims(collisionLines)
	if cw != sw || ch != sh {
		return nil, fmt.Errorf("sprite and collision sizes differ: sprite=%dx%d collision=%dx%d", sw, sh, cw, ch)
	}
	cells := make([]bool, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		for x := 0; x < sw; x++ {
			ch := runeAt(collisionLines[y], x)
			width := 1
			if widthMask != nil && x < len(widthMask[y]) {
				width = widthMask[y][x]
			}
			collides := ch != ' ' && ch != '.'
			for i := 0; i < width; i++ {
				idx := y*cellW + col
				col++
				if collides {
					cells[idx] = true
				}
			}
		}
		for col < cellW {
			col++
		}
	}
	return &render.CollisionMask{W: cellW, H: sh, Cells: cells}, nil
}
