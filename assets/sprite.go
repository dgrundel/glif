package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

func LoadSprite(basePath string) (*render.Sprite, error) {
	spriteLinesRaw, err := readLines(basePath + ".sprite")
	if err != nil {
		return nil, err
	}
	maskLinesRaw, err := readLines(basePath + ".color")
	if err != nil {
		return nil, err
	}

	spriteLines := toRunesLines(spriteLinesRaw)
	maskLines := toRunesLines(maskLinesRaw)

	sw, sh := dims(spriteLines)
	mw, mh := dims(maskLines)
	if sw != mw || sh != mh {
		return nil, fmt.Errorf("sprite and color sizes differ: sprite=%dx%d color=%dx%d", sw, sh, mw, mh)
	}

	widthLines, hasWidth, err := readOptionalLines(basePath + ".width")
	if err != nil {
		return nil, err
	}
	widthMask, cellW, err := parseWidthMask(widthLines, hasWidth, spriteLines, sw, sh)
	if err != nil {
		return nil, err
	}

	pal, err := palette.Load(resolvePalettePath(basePath))
	if err != nil {
		return nil, err
	}

	collisionMask, err := loadCollisionMask(basePath, spriteLines, widthMask, cellW, sw, sh)
	if err != nil {
		return nil, err
	}

	cells := make([]grid.Cell, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		row := spriteLines[y]
		for x := 0; x < len(row); x++ {
			spr := row[x]
			mask := runeAt(maskLines[y], x)
			width := 1
			if hasWidth {
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
					cells[idx] = grid.Cell{Ch: 0, Style: entry.Style, Skip: true}
				}
			}
		}
		if hasWidth && col != cellW {
			return nil, fmt.Errorf("width mask row %d expands to %d cells, expected %d", y+1, col, cellW)
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

func resolvePalettePath(basePath string) string {
	candidate := basePath + ".palette"
	if fileExists(candidate) {
		return candidate
	}
	dir := filepath.Dir(basePath)
	for {
		candidate = filepath.Join(dir, "default.palette")
		if fileExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return candidate
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func toRunesLines(lines []string) [][]rune {
	out := make([][]rune, len(lines))
	for i, line := range lines {
		out[i] = []rune(line)
	}
	return out
}

func dims(lines [][]rune) (int, int) {
	w := 0
	for _, line := range lines {
		if len(line) > w {
			w = len(line)
		}
	}
	return w, len(lines)
}

func runeAt(line []rune, x int) rune {
	if x < 0 || x >= len(line) {
		return ' '
	}
	return line[x]
}

func loadCollisionMask(basePath string, spriteLines [][]rune, widthMask [][]int, cellW, sw, sh int) (*render.CollisionMask, error) {
	path := basePath + ".collision"
	if !fileExists(path) {
		return nil, nil
	}
	linesRaw, err := readLines(path)
	if err != nil {
		return nil, err
	}
	lines := toRunesLines(linesRaw)
	cw, ch := dims(lines)
	if cw != sw || ch != sh {
		return nil, fmt.Errorf("sprite and collision sizes differ: sprite=%dx%d collision=%dx%d", sw, sh, cw, ch)
	}
	cells := make([]bool, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		row := spriteLines[y]
		for x := 0; x < len(row); x++ {
			ch := runeAt(lines[y], x)
			width := 1
			if widthMask != nil {
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
		if widthMask != nil && col != cellW {
			return nil, fmt.Errorf("collision row %d expands to %d cells, expected %d", y+1, col, cellW)
		}
	}
	return &render.CollisionMask{W: cellW, H: sh, Cells: cells}, nil
}

func readOptionalLines(path string) ([]string, bool, error) {
	if !fileExists(path) {
		return nil, false, nil
	}
	lines, err := readLines(path)
	if err != nil {
		return nil, false, err
	}
	return lines, true, nil
}

func parseWidthMask(lines []string, hasWidth bool, spriteLines [][]rune, sw, sh int) ([][]int, int, error) {
	if !hasWidth {
		return nil, sw, nil
	}
	widthLines := toRunesLines(lines)
	if len(widthLines) != sh {
		return nil, 0, fmt.Errorf("width mask height differs: sprite=%d width=%d", sh, len(widthLines))
	}
	out := make([][]int, sh)
	expected := -1
	for y := 0; y < sh; y++ {
		if len(widthLines[y]) != len(spriteLines[y]) {
			return nil, 0, fmt.Errorf("width mask row %d length differs: sprite=%d width=%d", y+1, len(spriteLines[y]), len(widthLines[y]))
		}
		out[y] = make([]int, len(widthLines[y]))
		rowWidth := 0
		for x, ch := range widthLines[y] {
			w, err := parseWidth(ch)
			if err != nil {
				return nil, 0, fmt.Errorf("width mask row %d col %d: %v", y+1, x+1, err)
			}
			out[y][x] = w
			rowWidth += w
		}
		if expected == -1 {
			expected = rowWidth
		} else if rowWidth != expected {
			return nil, 0, fmt.Errorf("width mask row %d expands to %d cells, expected %d", y+1, rowWidth, expected)
		}
	}
	return out, expected, nil
}

func parseWidth(ch rune) (int, error) {
	switch ch {
	case '1':
		return 1, nil
	case '2':
		return 2, nil
	default:
		return 0, fmt.Errorf("invalid width %q (expected 1 or 2)", string(ch))
	}
}
