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
		outPath     string
		fill        string
		transparent string
	)
	flag.StringVar(&outPath, "out", "", "output .mask path (defaults to same base name)")
	flag.StringVar(&fill, "fill", "x", "mask rune for non-space cells")
	flag.StringVar(&transparent, "transparent", ".", "mask rune for spaces")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: genmask [--out path] [--fill x] [--transparent .] path/to/sprite.sprite")
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
			} else {
				b.WriteByte(fillCh)
			}
		}
		maskLines = append(maskLines, b.String())
	}

	if outPath == "" {
		base := strings.TrimSuffix(spritePath, filepath.Ext(spritePath))
		outPath = base + ".mask"
	}

	out := strings.Join(maskLines, "\n")
	if len(lines) > 0 {
		out += "\n"
	}
	if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write mask: %v\n", err)
		os.Exit(1)
	}
}
