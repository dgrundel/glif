package spriteio

import (
	"errors"
	"path/filepath"
	"testing"
)

func unwrapJoin(err error) []error {
	if err == nil {
		return nil
	}
	type unwrapper interface {
		Unwrap() []error
	}
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return []error{err}
}

// TestFileFrameCountAndRows validates frame splitting via the File interface.
func TestFileFrameCountAndRows(t *testing.T) {
	path := filepath.Join("testdata", "a.sprite")

	f, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	count, err := f.FrameCount(2)
	if err != nil {
		t.Fatalf("FrameCount: %v", err)
	}
	if count != 2 {
		t.Fatalf("FrameCount=%d want=2", count)
	}

	rows, err := f.FrameRows(2, 1)
	if err != nil {
		t.Fatalf("FrameRows: %v", err)
	}
	if string(rows[0]) != "ef" {
		t.Fatalf("rows[0]=%q", string(rows[0]))
	}
}

// TestComputeWidthProfile validates width expansion with a width mask.
func TestComputeWidthProfile(t *testing.T) {
	sprite := [][]rune{
		[]rune("ab"),
		[]rune("c"),
	}
	width := [][]rune{
		[]rune("21"),
		[]rune("2"),
	}
	profile, err := ComputeWidthProfile(sprite, width)
	if err != nil {
		t.Fatalf("ComputeWidthProfile: %v", err)
	}
	if profile.CellWidth != 3 {
		t.Fatalf("CellWidth=%d want=3", profile.CellWidth)
	}
	if profile.RowWidths[0] != 3 || profile.RowWidths[1] != 2 {
		t.Fatalf("RowWidths=%v", profile.RowWidths)
	}
}

// TestExpandedRowsPadding verifies expanded rows are padded to max width.
func TestExpandedRowsPadding(t *testing.T) {
	sprite := [][]rune{
		[]rune("ab"),
		[]rune("c"),
	}
	width := [][]rune{
		[]rune("21"),
		[]rune("2"),
	}
	exp, err := expandRows(sprite, width)
	if err != nil {
		t.Fatalf("expandRows: %v", err)
	}
	if exp.MaxWidth != 3 {
		t.Fatalf("MaxWidth=%d want=3", exp.MaxWidth)
	}
	if string(exp.Rows[0]) != "a b" {
		t.Fatalf("row0=%q", string(exp.Rows[0]))
	}
	if string(exp.Rows[1]) != "c  " {
		t.Fatalf("row1=%q", string(exp.Rows[1]))
	}
}

// TestExpandedRowsWideRuneUsesMask ensures width comes from mask, not rune width.
func TestExpandedRowsWideRuneUsesMask(t *testing.T) {
	sprite := [][]rune{
		[]rune("界"),
	}
	width := [][]rune{
		[]rune("2"),
	}
	exp, err := expandRows(sprite, width)
	if err != nil {
		t.Fatalf("expandRows: %v", err)
	}
	if exp.MaxWidth != 2 {
		t.Fatalf("MaxWidth=%d want=2", exp.MaxWidth)
	}
	if string(exp.Rows[0]) != "界 " {
		t.Fatalf("row0=%q", string(exp.Rows[0]))
	}
}

// TestComputeWidthProfileWideRune ensures rune width does not affect expanded width.
func TestComputeWidthProfileWideRune(t *testing.T) {
	sprite := [][]rune{
		[]rune("界a"),
	}
	width := [][]rune{
		[]rune("21"),
	}
	profile, err := ComputeWidthProfile(sprite, width)
	if err != nil {
		t.Fatalf("ComputeWidthProfile: %v", err)
	}
	if profile.CellWidth != 3 {
		t.Fatalf("CellWidth=%d want=3", profile.CellWidth)
	}
	if profile.RowWidths[0] != 3 {
		t.Fatalf("RowWidths=%v", profile.RowWidths)
	}
}

// TestLoadSpriteSourceErrorsJoin ensures multiple parse errors are joined.
func TestLoadSpriteSourceErrorsJoin(t *testing.T) {
	base := filepath.Join("testdata", "bad")

	_, err := LoadSpriteSource(base)
	if err == nil {
		t.Fatalf("expected error")
	}
	errs := unwrapJoin(err)
	if len(errs) < 2 {
		t.Fatalf("expected multiple errors, got %d", len(errs))
	}
}

// TestLoadAnimationSourceFrames verifies animation frames load with per-frame masks.
func TestLoadAnimationSourceFrames(t *testing.T) {
	base := filepath.Join("testdata", "anim")

	anim, err := LoadAnimationSource(base, "walk")
	if err != nil {
		t.Fatalf("LoadAnimationSource: %v", err)
	}
	if len(anim.Frames) != 2 {
		t.Fatalf("FrameCount=%d want=2", len(anim.Frames))
	}
	if string(anim.Frames[0].Sprite[0]) != "cc" {
		t.Fatalf("frame0 sprite row0=%q", string(anim.Frames[0].Sprite[0]))
	}
	if string(anim.Frames[1].Color[0]) != "33" {
		t.Fatalf("frame1 color row0=%q", string(anim.Frames[1].Color[0]))
	}
}

// TestLoadAnimationSourceJoinErrors ensures animation parse errors are joined.
func TestLoadAnimationSourceJoinErrors(t *testing.T) {
	base := filepath.Join("testdata", "animerr")

	_, err := LoadAnimationSource(base, "bad")
	if err == nil {
		t.Fatalf("expected error")
	}
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("expected joined error")
	}
}
