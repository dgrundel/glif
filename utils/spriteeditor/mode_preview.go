package main

import (
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/render"
)

func (e *Editor) previewSprite() *render.Sprite {
	w, h := boundsSize(e.cells)
	if w == 0 || h == 0 {
		return nil
	}

	spriteLines := make([][]rune, h)
	for y := 0; y < h; y++ {
		row := make([]rune, w)
		for x := 0; x < w; x++ {
			row[x] = ' '
		}
		spriteLines[y] = row
	}
	for p, ch := range e.cells {
		if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
			continue
		}
		spriteLines[p.Y][p.X] = ch
	}

	widthMask := make([][]int, h)
	cellW := 0
	for y := 0; y < h; y++ {
		widthMask[y] = make([]int, w)
		rowWidth := 0
		for x := 0; x < w; x++ {
			width := 1
			if ch, ok := e.widthCells[Point{X: x, Y: y}]; ok {
				if parsed, err := parseWidthRune(ch); err == nil {
					width = parsed
				}
			}
			widthMask[y][x] = width
			rowWidth += width
		}
		if rowWidth > cellW {
			cellW = rowWidth
		}
	}

	cells := make([]grid.Cell, cellW*h)
	for y := 0; y < h; y++ {
		col := 0
		for x := 0; x < w; x++ {
			spr := runeAt(spriteLines[y], x)
			mask := e.colorDefault
			if ch, ok := e.colorCells[Point{X: x, Y: y}]; ok {
				mask = ch
			}
			width := widthMask[y][x]

			visible := mask != ' ' && mask != '.'
			style := e.previewErrorStyle
			if visible {
				if e.pal != nil {
					if entry, err := e.pal.Entry(mask); err == nil {
						style = entry.Style
						if entry.Transparent {
							visible = false
						}
					}
				}
			}

			for i := 0; i < width; i++ {
				idx := y*cellW + col
				col++
				if !visible {
					continue
				}
				if i == 0 {
					cells[idx] = grid.Cell{Ch: spr, Style: style}
				} else {
					cells[idx] = grid.Cell{Ch: 0, Style: style, Skip: true}
				}
			}
		}
		for col < cellW {
			col++
		}
	}

	return &render.Sprite{W: cellW, H: h, Cells: cells, Source: e.path}
}

func runeAt(line []rune, x int) rune {
	if x < 0 || x >= len(line) {
		return ' '
	}
	return line[x]
}
