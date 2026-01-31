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

func LoadMaskedSprite(basePath string) (*render.Sprite, error) {
	spriteLinesRaw, err := readLines(basePath + ".sprite")
	if err != nil {
		return nil, err
	}
	maskLinesRaw, err := readLines(basePath + ".mask")
	if err != nil {
		return nil, err
	}

	spriteLines := toRunesLines(spriteLinesRaw)
	maskLines := toRunesLines(maskLinesRaw)

	sw, sh := dims(spriteLines)
	mw, mh := dims(maskLines)
	if sw != mw || sh != mh {
		return nil, fmt.Errorf("sprite and mask sizes differ: sprite=%dx%d mask=%dx%d", sw, sh, mw, mh)
	}

	pal, err := loadPalette(resolvePalettePath(basePath))
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

func resolvePalettePath(basePath string) string {
	candidate := basePath + ".palette"
	if fileExists(candidate) {
		return candidate
	}
	dir := filepath.Dir(basePath)
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
	if strings.HasPrefix(name, "#") {
		r, g, b, ok := parseHexColor(name)
		if !ok {
			return 0, fmt.Errorf("invalid hex color: %q", name)
		}
		return tcell.NewRGBColor(int32(r), int32(g), int32(b)), nil
	}
	c := tcell.GetColor(name)
	if c == tcell.ColorDefault {
		return 0, fmt.Errorf("unknown color: %q", name)
	}
	return c, nil
}

func parseHexColor(s string) (uint8, uint8, uint8, bool) {
	if len(s) == 4 { // #rgb
		r, ok1 := hexNibble(s[1])
		g, ok2 := hexNibble(s[2])
		b, ok3 := hexNibble(s[3])
		if !ok1 || !ok2 || !ok3 {
			return 0, 0, 0, false
		}
		return r * 17, g * 17, b * 17, true
	}
	if len(s) == 7 { // #rrggbb
		r1, ok1 := hexNibble(s[1])
		r2, ok2 := hexNibble(s[2])
		g1, ok3 := hexNibble(s[3])
		g2, ok4 := hexNibble(s[4])
		b1, ok5 := hexNibble(s[5])
		b2, ok6 := hexNibble(s[6])
		if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 {
			return 0, 0, 0, false
		}
		return r1<<4 | r2, g1<<4 | g2, b1<<4 | b2, true
	}
	return 0, 0, 0, false
}

func hexNibble(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
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
