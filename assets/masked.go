package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type PaletteEntry struct {
	Style       grid.Style
	Transparent bool
}

type Palette map[rune]PaletteEntry

func LoadMaskedSprite(spritePath, maskPath, palPath string) (*render.Sprite, error) {
	spriteLines, err := readLines(spritePath)
	if err != nil {
		return nil, err
	}
	maskLines, err := readLines(maskPath)
	if err != nil {
		return nil, err
	}

	sw, sh := dims(spriteLines)
	mw, mh := dims(maskLines)
	if sw != mw || sh != mh {
		return nil, fmt.Errorf("sprite and mask sizes differ: sprite=%dx%d mask=%dx%d", sw, sh, mw, mh)
	}

	pal, err := loadPalette(resolvePalettePath(spritePath, palPath))
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

			entry, ok := pal[mask]
			if !ok {
				return nil, fmt.Errorf("palette missing key %q", string(mask))
			}
			if entry.Transparent {
				cells[y*sw+x] = grid.Cell{Ch: 0}
				continue
			}
			cells[y*sw+x] = grid.Cell{Ch: spr, Style: entry.Style}
		}
	}

	return &render.Sprite{W: sw, H: sh, Cells: cells}, nil
}

func resolvePalettePath(spritePath, palPath string) string {
	if palPath != "" {
		return palPath
	}
	dir := filepath.Dir(spritePath)
	base := strings.TrimSuffix(filepath.Base(spritePath), filepath.Ext(spritePath))
	candidate := filepath.Join(dir, base+".palette")
	if fileExists(candidate) {
		return candidate
	}
	return filepath.Join(dir, "default.palette")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadPalette(path string) (Palette, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	pal := Palette{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("invalid palette line: %q", line)
		}
		keyRunes := []rune(fields[0])
		if len(keyRunes) != 1 {
			return nil, fmt.Errorf("palette key must be single rune: %q", fields[0])
		}
		key := keyRunes[0]

		fg, err := parseColor(fields[1])
		if err != nil {
			return nil, err
		}
		bg, err := parseColor(fields[2])
		if err != nil {
			return nil, err
		}

		entry := PaletteEntry{Style: grid.Style{Fg: fg, Bg: bg}}
		for _, opt := range fields[3:] {
			opt = strings.ToLower(opt)
			switch opt {
			case "bold":
				entry.Style.Bold = true
			case "transparent":
				entry.Transparent = true
			}
		}
		pal[key] = entry
	}
	return pal, nil
}

func parseColor(name string) (tcell.Color, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "reset", "default":
		return tcell.ColorReset, nil
	}
	c := tcell.GetColor(name)
	if c == tcell.ColorDefault {
		return 0, fmt.Errorf("unknown color: %q", name)
	}
	return c, nil
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

func dims(lines []string) (int, int) {
	w := 0
	for _, line := range lines {
		if len(line) > w {
			w = len(line)
		}
	}
	return w, len(lines)
}

func runeAt(line string, x int) rune {
	if x < 0 || x >= len(line) {
		return ' '
	}
	return rune(line[x])
}
