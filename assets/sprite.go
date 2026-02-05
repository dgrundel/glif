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

	pal, err := palette.Load(resolvePalettePath(basePath))
	if err != nil {
		return nil, err
	}

	collisionMask, err := loadCollisionMask(basePath, sw, sh)
	if err != nil {
		return nil, err
	}

	cells := make([]grid.Cell, sw*sh)
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			spr := runeAt(spriteLines[y], x)
			mask := runeAt(maskLines[y], x)

			if mask == ' ' || mask == '.' {
				cells[y*sw+x] = grid.Cell{Ch: 0}
				continue
			}

			entry, err := pal.Entry(mask)
			if err != nil {
				return nil, err
			}
			if entry.Transparent {
				cells[y*sw+x] = grid.Cell{Ch: 0}
				continue
			}
			cells[y*sw+x] = grid.Cell{Ch: spr, Style: entry.Style}
		}
	}

	return &render.Sprite{W: sw, H: sh, Cells: cells, Collision: collisionMask, Source: basePath}, nil
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

func loadCollisionMask(basePath string, sw, sh int) (*render.CollisionMask, error) {
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
	cells := make([]bool, sw*sh)
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			ch := runeAt(lines[y], x)
			if ch != ' ' && ch != '.' {
				cells[y*sw+x] = true
			}
		}
	}
	return &render.CollisionMask{W: sw, H: sh, Cells: cells}, nil
}
