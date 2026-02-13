package assets

import "fmt"

func parseWidthMask(lines []string, hasWidth bool, spriteLines [][]rune, sw, sh int) ([][]int, int, error) {
	if !hasWidth {
		return nil, sw, nil
	}
	widthLines := toRunesLines(lines)
	if len(widthLines) != sh {
		return nil, 0, fmt.Errorf("width mask height differs: sprite=%d width=%d", sh, len(widthLines))
	}
	out := make([][]int, sh)
	maxWidth := 0
	for y := 0; y < sh; y++ {
		if len(widthLines[y]) > len(spriteLines[y]) {
			return nil, 0, fmt.Errorf("width mask row %d length exceeds sprite: sprite=%d width=%d", y+1, len(spriteLines[y]), len(widthLines[y]))
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
		rowWidth += (sw - len(widthLines[y]))
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	return out, maxWidth, nil
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
