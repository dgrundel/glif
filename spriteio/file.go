package spriteio

import (
	"fmt"
	"os"
	"strings"
)

// File is the shared abstraction for any sprite-related file.
// It does NOT care what the file "means" (sprite/color/width/collision/animation).
// It only knows: where it came from, how to split into frames, and how to apply width masks.
type File interface {
	// Path returns the original file path on disk.
	Path() string
	// RuneRows returns each line as []rune (no implicit padding).
	RuneRows() [][]rune
	// FrameCount returns the number of frames given a frame height.
	FrameCount(frameH int) (int, error)
	// FrameRows returns the rune rows for a single frame (no padding).
	FrameRows(frameH, frameIndex int) ([][]rune, error)
	// ExpandedRows returns rows with width mask applied.
	// When mask is nil, this is a no-op that treats all runes as width 1.
	ExpandedRows(mask File) (ExpandedRows, error)
}

// ExpandedRows is the width-expanded view of a file's rows.
type ExpandedRows struct {
	Rows     [][]rune // with padding for jagged rows
	RowWidth []int    // expanded width per row
	MaxWidth int      // max expanded width across all rows
}

type fileData struct {
	path string
	rows [][]rune
}

// LoadFile reads a file and returns a File interface for uniform access.
func LoadFile(path string) (File, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	return &fileData{
		path: path,
		rows: toRunesLines(lines),
	}, nil
}

func (f *fileData) Path() string {
	return f.path
}

func (f *fileData) RuneRows() [][]rune {
	return f.rows
}

func (f *fileData) FrameCount(frameH int) (int, error) {
	if frameH <= 0 {
		return 0, fmt.Errorf("frame height must be > 0")
	}
	total := len(f.rows)
	if total%frameH != 0 {
		return 0, fmt.Errorf("frame height mismatch: total=%d frameH=%d", total, frameH)
	}
	return total / frameH, nil
}

func (f *fileData) FrameRows(frameH, frameIndex int) ([][]rune, error) {
	count, err := f.FrameCount(frameH)
	if err != nil {
		return nil, err
	}
	if frameIndex < 0 || frameIndex >= count {
		return nil, fmt.Errorf("frame index out of range: %d (count=%d)", frameIndex, count)
	}
	start := frameIndex * frameH
	return f.rows[start : start+frameH], nil
}

func (f *fileData) ExpandedRows(mask File) (ExpandedRows, error) {
	var maskRows [][]rune
	if mask != nil {
		maskRows = mask.RuneRows()
	}
	return expandRows(f.rows, maskRows)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
