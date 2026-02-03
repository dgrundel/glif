package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		fill        string
		transparent string
		mappings    mapSpec
		writeColor  bool
		writeColl   bool
	)
	flag.StringVar(&fill, "fill", "x", "color rune for non-space cells")
	flag.StringVar(&transparent, "transparent", ".", "color rune for spaces")
	flag.Var(&mappings, "map", "character mapping (repeatable, e.g. --map a=b)")
	flag.Var(&mappings, "m", "character mapping (shorthand for --map)")
	flag.BoolVar(&writeColor, "color", false, "write .color output")
	flag.BoolVar(&writeColl, "collision", false, "write .collision output")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if !writeColor && !writeColl {
		fmt.Fprintln(os.Stderr, "must provide --color and/or --collision")
		flag.Usage()
		os.Exit(2)
	}
	if len(fill) != 1 || len(transparent) != 1 {
		fmt.Fprintln(os.Stderr, "--fill and --transparent must be single characters")
		os.Exit(2)
	}

	spritePath := flag.Arg(0)
	data, err := os.ReadFile(spritePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read sprite: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{}
	}
	runeLines := make([][]rune, len(lines))
	width := 0
	for i, line := range lines {
		runes := []rune(line)
		runeLines[i] = runes
		if len(runes) > width {
			width = len(runes)
		}
	}

	fillCh := fill[0]
	transCh := transparent[0]
	maskLines := make([]string, 0, len(runeLines))
	for _, line := range runeLines {
		var b strings.Builder
		b.Grow(width)
		for i := 0; i < width; i++ {
			ch := ' '
			if i < len(line) {
				ch = line[i]
			}
			if ch == ' ' {
				b.WriteByte(transCh)
				continue
			}
			if mapped, ok := mappings[ch]; ok {
				b.WriteRune(mapped)
				continue
			}
			b.WriteByte(fillCh)
		}
		maskLines = append(maskLines, b.String())
	}

	out := strings.Join(maskLines, "\n")
	if len(lines) > 0 {
		out += "\n"
	}
	base := strings.TrimSuffix(spritePath, filepath.Ext(spritePath))
	if writeColor {
		outPath := base + ".color"
		if strings.HasSuffix(spritePath, ".animation") {
			outPath = base + ".animation.color"
		}
		if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write color: %v\n", err)
			os.Exit(1)
		}
	}
	if writeColl {
		outPath := base + ".collision"
		if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write collision: %v\n", err)
			os.Exit(1)
		}
	}
}

type mapSpec map[rune]rune

func (m *mapSpec) String() string {
	if m == nil {
		return ""
	}
	parts := make([]string, 0, len(*m))
	for k, v := range *m {
		parts = append(parts, fmt.Sprintf("%c=%c", k, v))
	}
	return strings.Join(parts, ",")
}

func (m *mapSpec) Set(value string) error {
	if *m == nil {
		*m = map[rune]rune{}
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("map entry is empty")
	}
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid map entry %q (expected a=b)", value)
	}
	left := []rune(strings.TrimSpace(parts[0]))
	right := []rune(strings.TrimSpace(parts[1]))
	if len(left) != 1 || len(right) != 1 {
		return fmt.Errorf("map entry %q must be single characters", value)
	}
	(*m)[left[0]] = right[0]
	return nil
}
