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

	colorFile := loadRequiredFile(basePath+".color", &errs)
	if colorFile != nil {
		src.Color = colorFile.RuneRows()
		if err := validateMaskRows("color", src.Sprite, src.Color); err != nil {
			errs = append(errs, err)
		}
	}

	widthFile := loadOptionalFile(basePath+".width", &errs)
	if widthFile != nil {
		src.Width = widthFile.RuneRows()
		if _, err := ComputeWidthProfile(src.Sprite, src.Width); err != nil {
			errs = append(errs, err)
		}
	}

	collisionFile := loadOptionalFile(basePath+".collision", &errs)
	if collisionFile != nil {
		src.Collision = collisionFile.RuneRows()
		if err := validateMaskRows("collision", src.Sprite, src.Collision); err != nil {
			errs = append(errs, err)
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

	animColor := loadOptionalFile(basePath+"."+name+".animation.color", &errs)
	animWidth := loadOptionalFile(basePath+"."+name+".animation.width", &errs)
	animCollision := loadOptionalFile(basePath+"."+name+".animation.collision", &errs)

	frames := make([]*SpriteSource, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		spriteRows, err := animFile.FrameRows(frameH, i)
		if err != nil {
			errs = append(errs, err)
			break
		}

		colorRows := frameRowsOrBase(animColor, frameH, i, baseSrc.Color, &errs)
		widthRows := frameRowsOrBase(animWidth, frameH, i, baseSrc.Width, &errs)
		collisionRows := frameRowsOrBase(animCollision, frameH, i, baseSrc.Collision, &errs)

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

func loadOptionalFile(path string, errs *[]error) File {
	if !fileExists(path) {
		return nil
	}
	f, err := LoadFile(path)
	if err != nil {
		*errs = append(*errs, err)
		return nil
	}
	return f
}

func loadRequiredFile(path string, errs *[]error) File {
	f, err := LoadFile(path)
	if err != nil {
		*errs = append(*errs, err)
		return nil
	}
	return f
}

func frameRowsOrBase(f File, frameH, frameIndex int, base [][]rune, errs *[]error) [][]rune {
	if f == nil {
		return base
	}
	rows, err := f.FrameRows(frameH, frameIndex)
	if err != nil {
		*errs = append(*errs, err)
		return base
	}
	return rows
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
