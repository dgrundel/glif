package assets

import (
	"os"
	"strings"
)

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
