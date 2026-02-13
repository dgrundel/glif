package main

import (
	"os"
	"strings"

	"github.com/dgrundel/glif/grid"
)

func (e *Editor) ensureColorCells() {
	if e.colorCells == nil {
		e.colorCells = map[Point]rune{}
	}
	spriteW, spriteH := boundsSize(e.cells)
	if spriteW == 0 || spriteH == 0 {
		return
	}
	for y := 0; y < spriteH; y++ {
		for x := 0; x < spriteW; x++ {
			p := Point{X: x, Y: y}
			if _, ok := e.colorCells[p]; !ok {
				e.colorCells[p] = e.colorDefault
			}
		}
	}
}

func (e *Editor) colorStyleFor(key rune, base grid.Style) grid.Style {
	if e.pal == nil {
		return e.previewErrorStyle.Resolve(base)
	}
	entry, err := e.pal.Entry(key)
	if err != nil {
		return e.previewErrorStyle.Resolve(base)
	}
	return entry.Style.Resolve(base)
}

func readColorMask(path string, spriteW, spriteH int, fill rune) (map[Point]rune, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return generateFilledCells(spriteW, spriteH, fill), nil
		}
		return nil, err
	}
	text := string(data)
	if strings.HasSuffix(text, "\n") {
		text = strings.TrimRight(text, "\n")
	}
	if text == "" {
		return map[Point]rune{}, nil
	}
	lines := strings.Split(text, "\n")
	out := make(map[Point]rune)
	for y, line := range lines {
		for x, ch := range []rune(line) {
			out[Point{X: x, Y: y}] = ch
		}
	}
	return out, nil
}

func writeColorMask(path string, cells map[Point]rune, spriteW, spriteH int, fill rune) error {
	w, h := spriteW, spriteH
	if w == 0 || h == 0 {
		w, h = boundsSize(cells)
	}
	if w == 0 || h == 0 {
		return os.WriteFile(path, []byte(""), 0o644)
	}
	lines := make([][]rune, h)
	for y := 0; y < h; y++ {
		line := make([]rune, w)
		for x := 0; x < w; x++ {
			line[x] = fill
		}
		lines[y] = line
	}
	for p, ch := range cells {
		if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
			continue
		}
		lines[p.Y][p.X] = ch
	}
	parts := make([]string, h)
	for y := 0; y < h; y++ {
		parts[y] = string(lines[y])
	}
	out := strings.Join(parts, "\n")
	return os.WriteFile(path, []byte(out), 0o644)
}
