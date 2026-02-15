package spriteio

import (
	"errors"
	"fmt"
	"strings"
)

// One source type for static sprite data.
type SpriteSource struct {
	BasePath  string
	Sprite    [][]rune
	Color     [][]rune
	Width     [][]rune // optional
	Collision [][]rune // optional
}

// AnimationSource is a named list of sprite frames.
type AnimationSource struct {
	Name   string
	Frames []*SpriteSource
}

// LoadSpriteSource loads .sprite and its optional masks from a base path.
func LoadSpriteSource(basePath string) (*SpriteSource, error) {
	spriteFile, err := LoadFile(basePath + ".sprite")
	if err != nil {
		return nil, err
	}
	return loadSpriteSourceFromFile(spriteFile, basePath)
}

// LoadSpriteSourceFromFile loads a sprite source from a pre-loaded File (and optional sibling masks).
func LoadSpriteSourceFromFile(f File) (*SpriteSource, error) {
	if f == nil {
		return nil, fmt.Errorf("sprite file is nil")
	}
	basePath := strings.TrimSuffix(f.Path(), ".sprite")
	if basePath == f.Path() {
		return nil, fmt.Errorf("sprite file path must end with .sprite")
	}
	return loadSpriteSourceFromFile(f, basePath)
}

func loadSpriteSourceFromFile(spriteFile File, basePath string) (*SpriteSource, error) {
	errs := []error{}
	src := &SpriteSource{
		BasePath: basePath,
		Sprite:   spriteFile.RuneRows(),
	}

	colorFile, err := LoadFile(basePath + ".color")
	if err != nil {
		errs = append(errs, err)
	} else {
		src.Color = colorFile.RuneRows()
		if err := validateMaskRows("color", src.Sprite, src.Color); err != nil {
			errs = append(errs, err)
		}
	}

	if fileExists(basePath + ".width") {
		widthFile, err := LoadFile(basePath + ".width")
		if err != nil {
			errs = append(errs, err)
		} else {
			src.Width = widthFile.RuneRows()
			if _, err := ComputeWidthProfile(src.Sprite, src.Width); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if fileExists(basePath + ".collision") {
		collisionFile, err := LoadFile(basePath + ".collision")
		if err != nil {
			errs = append(errs, err)
		} else {
			src.Collision = collisionFile.RuneRows()
			if err := validateMaskRows("collision", src.Sprite, src.Collision); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return src, errors.Join(errs...)
}

// LoadAnimationSource loads .animation (and optional masks) for a base path + animation name.
func LoadAnimationSource(basePath, name string) (*AnimationSource, error) {
	errs := []error{}

	baseSrc, err := LoadSpriteSource(basePath)
	if err != nil {
		errs = append(errs, err)
	}
	if baseSrc == nil || len(baseSrc.Sprite) == 0 {
		return nil, errors.Join(append(errs, fmt.Errorf("base sprite is empty"))...)
	}

	animFile, err := LoadFile(basePath + "." + name + ".animation")
	if err != nil {
		return nil, errors.Join(append(errs, err)...)
	}

	frameH := len(baseSrc.Sprite)
	frameCount, err := animFile.FrameCount(frameH)
	if err != nil {
		return nil, errors.Join(append(errs, err)...)
	}

	var animColor File
	if fileExists(basePath + "." + name + ".animation.color") {
		animColor, err = LoadFile(basePath + "." + name + ".animation.color")
		if err != nil {
			errs = append(errs, err)
		}
	}
	var animWidth File
	if fileExists(basePath + "." + name + ".animation.width") {
		animWidth, err = LoadFile(basePath + "." + name + ".animation.width")
		if err != nil {
			errs = append(errs, err)
		}
	}
	var animCollision File
	if fileExists(basePath + "." + name + ".animation.collision") {
		animCollision, err = LoadFile(basePath + "." + name + ".animation.collision")
		if err != nil {
			errs = append(errs, err)
		}
	}

	frames := make([]*SpriteSource, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		spriteRows, err := animFile.FrameRows(frameH, i)
		if err != nil {
			errs = append(errs, err)
			break
		}

		colorRows := baseSrc.Color
		if animColor != nil {
			if rows, err := animColor.FrameRows(frameH, i); err != nil {
				errs = append(errs, err)
			} else {
				colorRows = rows
			}
		}

		widthRows := baseSrc.Width
		if animWidth != nil {
			if rows, err := animWidth.FrameRows(frameH, i); err != nil {
				errs = append(errs, err)
			} else {
				widthRows = rows
			}
		}

		collisionRows := baseSrc.Collision
		if animCollision != nil {
			if rows, err := animCollision.FrameRows(frameH, i); err != nil {
				errs = append(errs, err)
			} else {
				collisionRows = rows
			}
		}

		frame := &SpriteSource{
			BasePath:  basePath,
			Sprite:    spriteRows,
			Color:     colorRows,
			Width:     widthRows,
			Collision: collisionRows,
		}

		if err := validateMaskRows("color", frame.Sprite, frame.Color); err != nil {
			errs = append(errs, err)
		}
		if frame.Width != nil {
			if _, err := ComputeWidthProfile(frame.Sprite, frame.Width); err != nil {
				errs = append(errs, err)
			}
		}
		if frame.Collision != nil {
			if err := validateMaskRows("collision", frame.Sprite, frame.Collision); err != nil {
				errs = append(errs, err)
			}
		}

		frames = append(frames, frame)
	}

	return &AnimationSource{Name: name, Frames: frames}, errors.Join(errs...)
}

func validateMaskRows(name string, spriteRows, maskRows [][]rune) error {
	if maskRows == nil {
		return fmt.Errorf("%s mask missing", name)
	}
	if len(maskRows) != len(spriteRows) {
		return fmt.Errorf("%s mask height differs: sprite=%d %s=%d", name, len(spriteRows), name, len(maskRows))
	}
	maxLen := 0
	for _, row := range spriteRows {
		if len(row) > maxLen {
			maxLen = len(row)
		}
	}
	for y := 0; y < len(spriteRows); y++ {
		if len(maskRows[y]) > maxLen {
			return fmt.Errorf("%s mask row %d length exceeds sprite: sprite=%d %s=%d", name, y+1, maxLen, name, len(maskRows[y]))
		}
	}
	return nil
}
