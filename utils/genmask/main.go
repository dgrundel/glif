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
	flag.StringVar(&fill, "fill", "w", "mask rune for non-space cells")
	flag.StringVar(&transparent, "transparent", ".", "mask rune for spaces")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: genmask [--out path] [--fill w] [--transparent .] path/to/sprite.sprite")
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
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	fillCh := fill[0]
	transCh := transparent[0]
	maskLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) < width {
			line = line + strings.Repeat(" ", width-len(line))
		}
		var b strings.Builder
		b.Grow(width)
		for i := 0; i < width; i++ {
			if line[i] == ' ' {
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
