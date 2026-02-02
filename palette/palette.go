package palette

import (
	"fmt"
	"os"
	"strings"

	"github.com/dgrundel/glif/grid"
	"github.com/gdamore/tcell/v3"
)

type Entry struct {
	Style       grid.Style
	Transparent bool
}

type Palette struct {
	entries map[rune]Entry
}

func Load(path string) (*Palette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return &Palette{entries: map[rune]Entry{}}, nil
	}
	lines := strings.Split(text, "\n")
	entries := map[rune]Entry{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("invalid palette line %d: %q", i+1, line)
		}
		keyRunes := []rune(fields[0])
		if len(keyRunes) != 1 {
			return nil, fmt.Errorf("palette key must be single rune on line %d", i+1)
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

		entry := Entry{Style: grid.Style{Fg: fg, Bg: bg}}
		for _, opt := range fields[3:] {
			opt = strings.ToLower(opt)
			switch opt {
			case "bold":
				entry.Style.Bold = true
			case "transparent":
				entry.Transparent = true
			}
		}
		entries[key] = entry
	}

	return &Palette{entries: entries}, nil
}

func MustLoad(path string) *Palette {
	pal, err := Load(path)
	if err != nil {
		panic(err)
	}
	return pal
}

func (p *Palette) Entry(key rune) (Entry, error) {
	if p == nil {
		return Entry{}, fmt.Errorf("palette is nil")
	}
	entry, ok := p.entries[key]
	if !ok {
		return Entry{}, fmt.Errorf("palette missing key %q", string(key))
	}
	return entry, nil
}

func (p *Palette) Style(key rune) (grid.Style, error) {
	entry, err := p.Entry(key)
	if err != nil {
		return grid.Style{}, err
	}
	return entry.Style, nil
}

func (p *Palette) MustStyle(key rune) grid.Style {
	style, err := p.Style(key)
	if err != nil {
		panic(err)
	}
	return style
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
