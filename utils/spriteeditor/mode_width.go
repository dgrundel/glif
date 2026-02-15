package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dgrundel/glif/spriteio"
)

func (e *Editor) ensureWidthCells() {
	if e.widthCells == nil {
		e.widthCells = map[Point]rune{}
	}
	spriteW, spriteH := boundsSize(e.cells)
	if spriteW == 0 || spriteH == 0 {
		return
	}
	for y := 0; y < spriteH; y++ {
		for x := 0; x < spriteW; x++ {
			p := Point{X: x, Y: y}
			if _, ok := e.widthCells[p]; !ok {
				e.widthCells[p] = '1'
			}
		}
	}
}

func readWidthMask(path string) (map[Point]rune, bool, error) {
	f, err := spriteio.LoadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	rows := f.RuneRows()
	if len(rows) == 0 {
		return map[Point]rune{}, true, nil
	}
	out := make(map[Point]rune)
	for y, line := range rows {
		for x, ch := range line {
			out[Point{X: x, Y: y}] = ch
		}
	}
	return out, true, nil
}

func writeWidthMask(path string, cells map[Point]rune, spriteW, spriteH int) error {
	allOnes := true
	for _, ch := range cells {
		if ch != '1' {
			allOnes = false
			break
		}
	}
	if allOnes {
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
			line[x] = '1'
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

func parseWidthRune(ch rune) (int, error) {
	switch ch {
	case '1':
		return 1, nil
	case '2':
		return 2, nil
	default:
		return 0, fmt.Errorf("invalid width %q (expected 1 or 2)", string(ch))
	}
}
