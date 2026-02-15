package main

import (
	"os"
	"strings"

	"github.com/dgrundel/glif/spriteio"
)

func (e *Editor) ensureCollisionCells() {
	if e.collisionCells == nil {
		e.collisionCells = map[Point]rune{}
	}
	spriteW, spriteH := boundsSize(e.cells)
	if spriteW == 0 || spriteH == 0 {
		return
	}
	for y := 0; y < spriteH; y++ {
		for x := 0; x < spriteW; x++ {
			p := Point{X: x, Y: y}
			if _, ok := e.collisionCells[p]; !ok {
				e.collisionCells[p] = e.collisionDefault
			}
		}
	}
}

func readCollisionMask(path string, spriteW, spriteH int, fill rune) (map[Point]rune, error) {
	f, err := spriteio.LoadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return generateFilledCells(spriteW, spriteH, fill), nil
		}
		return nil, err
	}
	rows := f.RuneRows()
	if len(rows) == 0 {
		return map[Point]rune{}, nil
	}
	out := make(map[Point]rune)
	for y, line := range rows {
		for x, ch := range line {
			out[Point{X: x, Y: y}] = ch
		}
	}
	return out, nil
}

func writeCollisionMask(path string, cells map[Point]rune, spriteW, spriteH int, fill rune) error {
	hasCollisions := false
	for _, ch := range cells {
		if isCollisionKey(ch) {
			hasCollisions = true
			break
		}
	}
	if !hasCollisions {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

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

func isCollisionKey(ch rune) bool {
	return ch != ' ' && ch != '.'
}
