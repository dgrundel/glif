package spriteio

import "fmt"

// WidthProfile describes expanded widths for a set of rows.
type WidthProfile struct {
	CellWidth int
	RowWidths []int
}

// ComputeWidthProfile computes expanded widths from sprite + width mask rows.
func ComputeWidthProfile(spriteRows, widthRows [][]rune) (WidthProfile, error) {
	if widthRows != nil && len(widthRows) != len(spriteRows) {
		return WidthProfile{}, fmt.Errorf("width mask height differs: sprite=%d width=%d", len(spriteRows), len(widthRows))
	}

	rowWidths := make([]int, len(spriteRows))
	maxWidth := 0
	for y, row := range spriteRows {
		var maskRow []rune
		if widthRows != nil {
			maskRow = widthRows[y]
			if len(maskRow) > len(row) {
				return WidthProfile{}, fmt.Errorf("width mask row %d length exceeds sprite: sprite=%d width=%d", y+1, len(row), len(maskRow))
			}
		}
		rowWidth := 0
		for x := range row {
			width := 1
			if maskRow != nil && x < len(maskRow) {
				parsed, err := parseWidth(maskRow[x])
				if err != nil {
					return WidthProfile{}, fmt.Errorf("width mask row %d col %d: %v", y+1, x+1, err)
				}
				width = parsed
			}
			rowWidth += width
		}
		rowWidths[y] = rowWidth
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}

	return WidthProfile{CellWidth: maxWidth, RowWidths: rowWidths}, nil
}

// ParseWidthMask returns width-per-rune values and the max expanded width.
func ParseWidthMask(spriteRows, widthRows [][]rune) ([][]int, int, error) {
	profile, err := ComputeWidthProfile(spriteRows, widthRows)
	if err != nil {
		return nil, 0, err
	}
	if widthRows == nil {
		return nil, profile.CellWidth, nil
	}

	out := make([][]int, len(widthRows))
	for y, row := range widthRows {
		out[y] = make([]int, len(row))
		for x, ch := range row {
			w, err := parseWidth(ch)
			if err != nil {
				return nil, 0, fmt.Errorf("width mask row %d col %d: %v", y+1, x+1, err)
			}
			out[y][x] = w
		}
	}
	return out, profile.CellWidth, nil
}

func expandRows(spriteRows, widthRows [][]rune) (ExpandedRows, error) {
	profile, err := ComputeWidthProfile(spriteRows, widthRows)
	if err != nil {
		return ExpandedRows{}, err
	}

	out := make([][]rune, len(spriteRows))
	for y, row := range spriteRows {
		var maskRow []rune
		if widthRows != nil {
			maskRow = widthRows[y]
		}
		expanded := make([]rune, 0, profile.CellWidth)
		for x, ch := range row {
			width := 1
			if maskRow != nil && x < len(maskRow) {
				parsed, err := parseWidth(maskRow[x])
				if err != nil {
					return ExpandedRows{}, fmt.Errorf("width mask row %d col %d: %v", y+1, x+1, err)
				}
				width = parsed
			}
			expanded = append(expanded, ch)
			for i := 1; i < width; i++ {
				expanded = append(expanded, ' ')
			}
		}
		for len(expanded) < profile.CellWidth {
			expanded = append(expanded, ' ')
		}
		out[y] = expanded
	}

	return ExpandedRows{Rows: out, RowWidth: profile.RowWidths, MaxWidth: profile.CellWidth}, nil
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
