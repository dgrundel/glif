package assets

import (
	"fmt"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/spriteio"
)

func LoadSprite(basePath string) (*render.Sprite, error) {
	src, err := spriteio.LoadSpriteSource(basePath)
	if err != nil {
		return nil, err
	}
	spriteLines := src.Sprite
	maskLines := src.Color

	sw, sh := dims(spriteLines)
	mw, mh := dims(maskLines)
	if sw != mw || sh != mh {
		return nil, fmt.Errorf("sprite and color sizes differ: sprite=%dx%d color=%dx%d", sw, sh, mw, mh)
	}

	widthMask, cellW, err := spriteio.ParseWidthMask(spriteLines, src.Width)
	if err != nil {
		return nil, err
	}

	pal, err := palette.Load(resolvePalettePath(basePath))
	if err != nil {
		return nil, err
	}

	collisionMask, err := buildCollisionMask(spriteLines, src.Collision, widthMask, cellW, sw, sh)
	if err != nil {
		return nil, err
	}

	cells := make([]grid.Cell, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		for x := 0; x < sw; x++ {
			spr := runeAt(spriteLines[y], x)
			mask := runeAt(maskLines[y], x)
			width := 1
			if widthMask != nil && x < len(widthMask[y]) {
				width = widthMask[y][x]
			}

			visible := mask != ' ' && mask != '.'
			var entry palette.Entry
			if visible {
				entry, err = pal.Entry(mask)
				if err != nil {
					return nil, err
				}
				if entry.Transparent {
					visible = false
				}
			}

			for i := 0; i < width; i++ {
				idx := y*cellW + col
				col++
				if !visible {
					continue
				}
				if i == 0 {
					cells[idx] = grid.Cell{Ch: spr, Style: entry.Style}
				} else {
					cells[idx] = grid.SkipCell(entry.Style)
				}
			}
		}
		for col < cellW {
			col++
		}
	}

	return &render.Sprite{W: cellW, H: sh, Cells: cells, Collision: collisionMask, Source: basePath}, nil
}

func MustLoadSprite(basePath string) *render.Sprite {
	sprite, err := LoadSprite(basePath)
	if err != nil {
		panic(err)
	}
	return sprite
}
