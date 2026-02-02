package tilemap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/render"
)

type TilesetMapping struct {
	IDs map[rune]int
	Map *Map
}

func LoadFromFiles(mapPath, tilesPath string) (*Map, error) {
	m, _, err := loadMapWithMapping(mapPath, tilesPath)
	return m, err
}

func loadMapWithMapping(mapPath, tilesPath string) (*Map, map[rune]int, error) {
	mappings, tileset, tileW, tileH, err := loadTileset(tilesPath)
	if err != nil {
		return nil, nil, err
	}

	lines, err := readLines(mapPath)
	if err != nil {
		return nil, nil, err
	}
	grid := toRuneLines(lines)
	w, h := dims(grid)
	if w == 0 || h == 0 {
		return New(0, 0, tileW, tileH, 0), mappings, nil
	}

	m := New(w, h, tileW, tileH, 0)
	m.Tileset = tileset
	for y := 0; y < h; y++ {
		line := grid[y]
		for x := 0; x < w; x++ {
			ch := runeAt(line, x)
			if ch == ' ' {
				continue
			}
			id, ok := mappings[ch]
			if !ok {
				return nil, nil, fmt.Errorf("map uses unmapped tile %q", string(ch))
			}
			m.Set(x, y, id)
		}
	}
	return m, mappings, nil
}

func loadTileset(path string) (map[rune]int, map[int]*render.Sprite, int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, 1, 1, err
	}
	defer file.Close()

	ids := map[rune]int{}
	tileset := map[int]*render.Sprite{}
	nextID := 1
	var tileW, tileH int

	scanner := bufio.NewScanner(file)
	lineNo := 0
	dir := filepath.Dir(path)
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, nil, 1, 1, fmt.Errorf("invalid tiles line %d: %q", lineNo, line)
		}
		r := []rune(fields[0])
		if len(r) != 1 {
			return nil, nil, 1, 1, fmt.Errorf("tile key must be single rune on line %d", lineNo)
		}
		key := r[0]
		name := fields[1]
		if name == "empty" || name == "none" {
			ids[key] = 0
			continue
		}
		base := name
		if !filepath.IsAbs(base) {
			base = filepath.Join(dir, base)
		}
		sprite, err := assets.LoadSprite(base)
		if err != nil {
			return nil, nil, 1, 1, fmt.Errorf("load sprite %q: %w", name, err)
		}
		if tileW == 0 {
			tileW = sprite.W
			tileH = sprite.H
		} else if sprite.W != tileW || sprite.H != tileH {
			return nil, nil, 1, 1, fmt.Errorf("sprite %q size %dx%d does not match tileset size %dx%d", name, sprite.W, sprite.H, tileW, tileH)
		}
		ids[key] = nextID
		tileset[nextID] = sprite
		nextID++
		// Note: tileset entries are stored in Map after creation.
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, 1, 1, err
	}
	if tileW == 0 {
		tileW = 1
		tileH = 1
	}
	return ids, tileset, tileW, tileH, nil
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

func toRuneLines(lines []string) [][]rune {
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
