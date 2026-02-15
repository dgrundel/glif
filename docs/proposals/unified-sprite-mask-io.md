# Unified Sprite/Mask IO + Frame API

## Summary
Introduce a unified set of interfaces and helpers for loading sprite assets and their masks (color, width, collision) and for splitting/handling animation frames. The goal is to eliminate duplicated parsing logic, remove inconsistent width calculations, and produce actionable error messages that point to exact files and line numbers. A single, low-level **file interface** represents any sprite/mask/animation file and encapsulates width-masking semantics.

## Goals
- One consistent pipeline for `.sprite`, `.color`, `.width`, `.collision`, and `.animation` files.
- A single file interface used by *all* sprite/mask/animation files.
- Frame splitting that is reusable across color/width/animation files.
- Width expansion based on the **width mask and the actual frame rows**, not raw rune count.
- Useful error messages with file paths, frame index, row index, expected vs actual widths/heights.
- Compatible with current asset conventions and names.
- A pure parser: no implicit fixes, padding, or auto-corrections.

## Non‑Goals
- Replacing the rendering system or adding new rendering features.
- Changing file formats (only new helpers and consistent logic).
- Forcing the editor to use the same strict validation (editor can opt into tolerant mode).
- Auto-repairing malformed files.

---

## Proposed API

### 1) Low-Level File Interface (All File Types)
```go
// File is the shared abstraction for any sprite-related file.
// It does NOT care what the file "means" (sprite/color/width/collision/animation).
// It only knows: where it came from, how to split into frames, and how to apply width masks.
type File interface {
    // Path returns the original file path on disk.
    Path() string
    // Lines returns the raw lines as read from disk (no trimming/padding).
    Lines() []string
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

type ExpandedRows struct {
    Rows     [][]rune // with padding for jagged rows
    RowWidth []int    // expanded width per row
    MaxWidth int      // max expanded width across all rows
}

// LoadFile reads a file and returns a File interface for uniform access.
func LoadFile(path string) (File, error)
```
Rules:
- `RuneRows` are raw runes from file lines (no implicit padding).
- `ExpandedRows` uses **mask rows** to expand widths; mask widths are applied per rune.
- Expanded rows are padded to `MaxWidth` (using spaces) to support jagged input.
- `FrameRows` and `ExpandedRows` are the *only* way widths are computed.
- Parser never mutates or auto-fixes data; it only reports errors/warnings.

### 2) Unified Source (Sprite or Animation)
```go
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
func LoadSpriteSource(basePath string) (*SpriteSource, error)
// LoadSpriteSourceFromFile loads a sprite source from a pre-loaded File (and optional sibling masks).
func LoadSpriteSourceFromFile(f File) (*SpriteSource, error)
// LoadAnimationSource loads .animation (and optional masks) for a base path + animation name.
func LoadAnimationSource(basePath, name string) (*AnimationSource, error)
```
- `basePath` is the shared base (e.g., `demos/foo/enemy`).
- `Sprite` and `Color` required.
- `Width` and `Collision` optional.

### 3) Frame Splitting
```go
// SplitFrames splits stacked lines into frames of height frameH.
func SplitFrames(lines []string, frameH int) ([][][]rune, error)
```
Returns `frames[frameIndex][rowIndex][]rune`.

### 4) Frame View
```go
type FrameView struct {
    // SpriteRows are the per-frame sprite runes (unexpanded, no padding).
    SpriteRows    [][]rune
    // ColorRows are the per-frame color mask runes (unexpanded, no padding).
    ColorRows     [][]rune
    // WidthRows are the per-frame width mask runes (unexpanded, no padding).
    WidthRows     [][]rune // optional
    // CollisionRows are the per-frame collision mask runes (unexpanded, no padding).
    CollisionRows [][]rune // optional
}

// Frame returns the i-th animation frame (as a SpriteSource).
func (a *AnimationSource) Frame(i int) (*SpriteSource, error)
// FrameCount returns the number of frames.
func (a *AnimationSource) FrameCount() int
```

### 5) Width Expansion
```go
type WidthProfile struct {
    CellWidth int
    RowWidths []int
}

// ComputeWidthProfile computes expanded widths from sprite + width mask rows.
func ComputeWidthProfile(spriteRows, widthRows [][]rune) (WidthProfile, error)
```
Rules:
- If `widthRows` is nil: every rune is width 1.
- If `widthRows` exists: width comes from mask rune, missing mask entries default to 1.
- Expanded width is computed per row; `CellWidth` is max across rows.

### 6) Build Concrete Sprites
```go
// BuildSprite builds a render.Sprite using color/width rules and skip cells.
func BuildSprite(f *FrameView, pal *palette.Palette) (*render.Sprite, error)
```
Builds a `render.Sprite` using color/width rules and produces skip cells for wide glyphs.

### 7) Build Animations
```go
// BuildAnimation builds a render.Animation from a base sprite + animation source.
func BuildAnimation(base *render.Sprite, anim *AnimationSource, pal *palette.Palette) (*render.Animation, error)
```
Builds frames via `FrameView` and `BuildSprite`, ensuring all frames match the base expanded width.

---

## Error Messages (Actionable)
All errors include file path and row/frame indices. Examples:

```
width mask row too long
file: demos/invaders/assets/enemy.destroy.animation.width
frame: 3 row: 2
mask runes: 8
sprite runes: 7
```

```
animation width mismatch
file: demos/invaders/assets/enemy.destroy.animation
frame: 4 row: 1
expanded width: 8 (expected 9)
```

```
color size mismatch
file: demos/invaders/assets/enemy.destroy.animation.color
expected height: 24
got: 23
```

---

## How This Fixes Current Issues
- Avoids raw rune-width comparisons when width masks exist.
- Always uses the correct frame rows when computing widths.
- One place to debug/modify width or frame rules.
- All file handling goes through `File`, so mask behavior is consistent.
- Tools can still render with partial data while surfacing parse errors.

---

## Migration Plan (Incremental)
1. Add `SpriteSource`, `AnimationSource`, and `SplitFrames` in a new package (e.g., `spriteio`).
2. Add `LoadFile` + `File` implementation and migrate all file reads to it.
3. Update `assets.LoadSprite` to use the new pipeline.
4. Update animation loading to split frames and pass each through the sprite loading path.
5. Update tools (spritepreview, spriteeditor) to use tolerant mode or direct `spriteio` helpers.
6. Consolidate width-mask logic into the shared pipeline and delete duplicated parsers.

---

## Alternatives Considered
1) Patch existing parsing in-place
- Lower effort, but duplicated logic remains across assets/render/tools.

2) Force tools to use assets package
- Simplifies API surface but prevents editors from loading broken assets.

---

## Appendix: Example Usage
```go
spriteFile, _ := LoadFile("demos/invaders/assets/enemy.sprite")
widthFile, _ := LoadFile("demos/invaders/assets/enemy.width")
expanded, _ := spriteFile.ExpandedRows(widthFile)
_ = expanded.MaxWidth

sprite, _ := LoadSpriteSource("demos/invaders/assets/enemy")
profile, _ := ComputeWidthProfile(sprite.Sprite, sprite.Width)
_ = profile.CellWidth

animSrc, _ := LoadAnimationSource(sprite.BasePath, "destroy")
frames := animSrc.FrameCount()
for i := 0; i < frames; i++ {
    f, _ := animSrc.Frame(i)
    _ = f
}
```
