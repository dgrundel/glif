package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/palette"
)

type Animation struct {
	Base   *Sprite
	Frames []*Sprite
}

type AnimationPlayer struct {
	anim   *Animation
	fps    float64
	accum  float64
	index  int
	locked bool
}

func (s *Sprite) LoadAnimation(name string) (*Animation, error) {
	if s == nil {
		return nil, fmt.Errorf("sprite is nil")
	}
	if s.Source == "" {
		return nil, fmt.Errorf("sprite has no source path")
	}
	if name == "" {
		return nil, fmt.Errorf("animation name is empty")
	}

	base := s.Source
	animPath := base + "." + name + ".animation"
	animLinesRaw, err := readLines(animPath)
	if err != nil {
		return nil, err
	}
	animLines := toRunesLines(animLinesRaw)
	aw, ah := dims(animLines)

	spriteLinesRaw, err := readLines(base + ".sprite")
	if err != nil {
		return nil, err
	}
	spriteLines := toRunesLines(spriteLinesRaw)
	sw, sh := dims(spriteLines)
	if sh != s.H {
		return nil, fmt.Errorf("sprite height differs: sprite=%d file=%d", s.H, sh)
	}
	if aw != sw {
		return nil, fmt.Errorf("animation width differs: sprite=%d animation=%d", sw, aw)
	}

	widthLinesRaw, hasWidth, err := readOptionalLines(base + ".width")
	if err != nil {
		return nil, err
	}
	widthMask, cellW, err := parseWidthMask(widthLinesRaw, hasWidth, spriteLines, sh)
	if err != nil {
		return nil, err
	}
	if cellW != s.W {
		return nil, fmt.Errorf("animation width differs: sprite=%d animation=%d", s.W, cellW)
	}
	if ah == 0 || ah%s.H != 0 {
		return nil, fmt.Errorf("animation height must be multiple of sprite height: sprite=%d animation=%d", s.H, ah)
	}
	frameCount := ah / s.H

	colorLines, err := loadAnimationColors(base, name, s.H, frameCount)
	if err != nil {
		return nil, err
	}
	cw, _ := dims(colorLines)
	if cw != sw {
		return nil, fmt.Errorf("animation color width differs: expected=%d got=%d", sw, cw)
	}

	pal, err := palette.Load(resolvePalettePath(base))
	if err != nil {
		return nil, err
	}

	frames := make([]*Sprite, 0, frameCount+1)
	frames = append(frames, s)
	for frame := 0; frame < frameCount; frame++ {
		startY := frame * s.H
		cells := make([]grid.Cell, s.W*s.H)
		for y := 0; y < s.H; y++ {
			col := 0
			rowLen := len(spriteLines[y])
			for x := 0; x < rowLen; x++ {
				spr := runeAt(animLines[startY+y], x)
				mask := runeAt(colorLines[startY+y], x)
				width := 1
				if hasWidth && x < len(widthMask[y]) {
					width = widthMask[y][x]
				}
				visible := mask != ' ' && mask != '.'
				var entry palette.Entry
				if visible {
					entry, err = pal.Entry(mask)
					if err != nil {
						return nil, err
					}
					if entry.Transparent {
						visible = false
					}
				}
				for i := 0; i < width; i++ {
					idx := y*s.W + col
					col++
					if !visible {
						continue
					}
					if i == 0 {
						cells[idx] = grid.Cell{Ch: spr, Style: entry.Style}
					} else {
						cells[idx] = grid.SkipCell(entry.Style)
					}
				}
			}
			for col < s.W {
				col++
			}
		}
		frames = append(frames, &Sprite{
			W:         s.W,
			H:         s.H,
			Cells:     cells,
			Collision: s.Collision,
			Source:    s.Source,
		})
	}

	return &Animation{Base: s, Frames: frames}, nil
}

func (a *Animation) Play(fps float64) *AnimationPlayer {
	if fps <= 0 {
		fps = 8
	}
	return &AnimationPlayer{anim: a, fps: fps}
}

func (p *AnimationPlayer) Update(dt float64) {
	if p == nil || p.anim == nil || p.locked {
		return
	}
	if p.fps <= 0 {
		p.fps = 8
	}
	p.accum += dt
	step := 1.0 / p.fps
	for p.accum >= step {
		p.accum -= step
		p.index = (p.index + 1) % len(p.anim.Frames)
	}
}

func (p *AnimationPlayer) Sprite() *Sprite {
	if p == nil || p.anim == nil || len(p.anim.Frames) == 0 {
		return nil
	}
	return p.anim.Frames[p.index]
}

func loadAnimationColors(base, name string, frameH, frameCount int) ([][]rune, error) {
	colorPath := base + "." + name + ".animation.color"
	if fileExists(colorPath) {
		linesRaw, err := readLines(colorPath)
		if err != nil {
			return nil, err
		}
		lines := toRunesLines(linesRaw)
		_, h := dims(lines)
		if h != frameH*frameCount {
			return nil, fmt.Errorf("animation color height differs: expected=%d got=%d", frameH*frameCount, h)
		}
		return lines, nil
	}

	baseColorPath := base + ".color"
	linesRaw, err := readLines(baseColorPath)
	if err != nil {
		return nil, err
	}
	lines := toRunesLines(linesRaw)
	_, h := dims(lines)
	if h != frameH {
		return nil, fmt.Errorf("base color height differs: expected=%d got=%d", frameH, h)
	}
	out := make([][]rune, 0, frameH*frameCount)
	for i := 0; i < frameCount; i++ {
		for _, line := range lines {
			out = append(out, line)
		}
	}
	return out, nil
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

func parseWidthMask(lines []string, hasWidth bool, spriteLines [][]rune, sh int) ([][]int, int, error) {
	if !hasWidth {
		maxW := 0
		for _, row := range spriteLines {
			if len(row) > maxW {
				maxW = len(row)
			}
		}
		return nil, maxW, nil
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
		rowWidth += (len(spriteLines[y]) - len(widthLines[y]))
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
