package render

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/spriteio"
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
	animSrc, err := spriteio.LoadAnimationSource(base, name)
	if err != nil {
		return nil, err
	}

	pal, err := palette.Load(resolvePalettePath(base))
	if err != nil {
		return nil, err
	}

	frames := make([]*Sprite, 0, len(animSrc.Frames)+1)
	frames = append(frames, s)
	for i, frameSrc := range animSrc.Frames {
		frame, err := buildSpriteFromSource(frameSrc, pal)
		if err != nil {
			return nil, fmt.Errorf("frame %d: %w", i+1, err)
		}
		if frame.W != s.W || frame.H != s.H {
			return nil, fmt.Errorf("animation frame %d size differs: sprite=%dx%d frame=%dx%d", i+1, s.W, s.H, frame.W, frame.H)
		}
		frames = append(frames, frame)
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

func buildSpriteFromSource(src *spriteio.SpriteSource, pal *palette.Palette) (*Sprite, error) {
	if src == nil {
		return nil, fmt.Errorf("sprite source is nil")
	}
	if pal == nil {
		return nil, fmt.Errorf("palette is nil")
	}
	spriteLines := src.Sprite
	maskLines := src.Color
	sw, sh := dims(spriteLines)
	mw, mh := dims(maskLines)
	if sw != mw || sh != mh {
		return nil, fmt.Errorf("sprite and color sizes differ: sprite=%dx%d color=%dx%d", sw, sh, mw, mh)
	}

	widthMask, cellW, err := spriteio.ParseWidthMask(spriteLines, src.Width)
	if err != nil {
		return nil, err
	}

	collisionMask, err := buildCollisionMask(spriteLines, src.Collision, widthMask, cellW, sw, sh)
	if err != nil {
		return nil, err
	}

	cells := make([]grid.Cell, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		for x := 0; x < sw; x++ {
			spr := runeAt(spriteLines[y], x)
			mask := runeAt(maskLines[y], x)
			width := 1
			if widthMask != nil && x < len(widthMask[y]) {
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
				idx := y*cellW + col
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
		for col < cellW {
			col++
		}
	}

	return &Sprite{W: cellW, H: sh, Cells: cells, Collision: collisionMask, Source: src.BasePath}, nil
}

func buildCollisionMask(spriteLines, collisionLines [][]rune, widthMask [][]int, cellW, sw, sh int) (*CollisionMask, error) {
	if collisionLines == nil {
		return nil, nil
	}
	cw, ch := dims(collisionLines)
	if cw != sw || ch != sh {
		return nil, fmt.Errorf("sprite and collision sizes differ: sprite=%dx%d collision=%dx%d", sw, sh, cw, ch)
	}
	cells := make([]bool, cellW*sh)
	for y := 0; y < sh; y++ {
		col := 0
		for x := 0; x < sw; x++ {
			ch := runeAt(collisionLines[y], x)
			width := 1
			if widthMask != nil && x < len(widthMask[y]) {
				width = widthMask[y][x]
			}
			collides := ch != ' ' && ch != '.'
			for i := 0; i < width; i++ {
				idx := y*cellW + col
				col++
				if collides {
					cells[idx] = true
				}
			}
		}
		for col < cellW {
			col++
		}
	}
	return &CollisionMask{W: cellW, H: sh, Cells: cells}, nil
}
