# Tilemap Camera Integration

## Summary
Refactor `tilemap.Map.Draw` to use the proposed `camera.Camera` interface and integer world coordinates. This removes duplicated projection math, makes tilemaps first‑class camera citizens, and ensures consistent culling across the engine.

## Why This Is Useful
- **Single source of truth** for world‑to‑screen math (the camera).
- **Consistency** with the renderer and camera proposal (same interface, same semantics).
- **Simpler API** for game code: pass world coords as ints; no float math.

## Feature Overview
- Change the tilemap draw signature to accept a `camera.Camera` interface (not a concrete type).
- Move visible bounds calculations to the camera where possible.
- Keep tilemap rendering in world space, with camera handling projection and clipping.

## Fit With Existing Architecture
- The camera proposal adds a `camera.Camera` interface with `WorldToScreen` and `Visible`.
- The renderer will optionally apply the camera transform; tilemap should follow the same contract.

## Proposed API (exhaustive)
### tilemap package
```go
// Map represents a single tilemap.
type Map struct {
	W, H    int
	TileW   int
	TileH   int
	Empty   int
	Tiles   []int
	Tileset map[int]*render.Sprite
}

// New creates a map with the given dimensions and tile size.
func New(w, h, tileW, tileH, empty int) *Map

// InBounds reports whether tile coords are within bounds.
func (m *Map) InBounds(x, y int) bool

// Set sets the tile ID at (x, y) if in bounds.
func (m *Map) Set(x, y, id int)

// At returns the tile ID at (x, y) or Empty if out of bounds.
func (m *Map) At(x, y int) int

// Draw renders the map in world space at the given origin.
// If cam is nil, coordinates are treated as screen space.
func (m *Map) Draw(r *render.Renderer, worldX, worldY int, cam camera.Camera)

// VisibleBounds returns the inclusive tile bounds visible in the camera viewport.
// If cam is nil, returns the full map bounds.
func (m *Map) VisibleBounds(worldX, worldY int, cam camera.Camera) (startX, startY, endX, endY int)
```

### tilemap package (loading)
```go
type TilesetMapping struct {
	IDs map[rune]int
	Map *Map
}

// LoadFromFiles loads a map and tileset file.
func LoadFromFiles(mapPath, tilesPath string) (*Map, error)
```

### camera package (expected usage)
```go
// Visible checks whether a world-space rect intersects the viewport.
Visible(x, y, w, h int) bool

// WorldToScreen converts world-space coords to screen-space coords.
WorldToScreen(x, y int) (sx, sy int)
```

## Implementation Plan
1. **API change**: update `Map.Draw` signature to use `int` world coords and `camera.Camera` interface.
2. **Projection**: compute tile world coords as ints; use `cam.WorldToScreen` if provided.
3. **Culling**: replace internal float‑based `visibleBounds` with camera‑derived bounds or simpler checks.
4. **Demo updates**: update `demos/world` to use the new signature.

## Alternatives Considered
1. **Keep internal float math**: minimal change but duplicates camera logic and risks inconsistencies.
2. **Let renderer handle camera only**: draw in world coords and rely on renderer to project. This is viable if the renderer applies camera transforms uniformly.
3. **Separate map renderer**: a dedicated tilemap renderer that owns its own camera logic; heavier and redundant.

## Risks / Tradeoffs
- API change will require updating any demo code using tilemaps.
- Camera visibility math must be correct for tiles; off‑by‑one issues can cause edge tiles to flicker.

## Testing Plan
- Add a manual regression test in `demos/world` for smooth panning without column “sticking.”
- Unit test for visible bounds of tiles at various camera offsets (optional).

## Appendix: Minimal Demo
```go
package main

import (
	"log"

	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/tilemap"
)

const (
	ActionLeft  input.Action = "left"
	ActionRight input.Action = "right"
	ActionUp    input.Action = "up"
	ActionDown  input.Action = "down"
)

type Game struct {
	m       *tilemap.Map
	cam     *camera.Basic
	actions input.ActionState
}

func (g *Game) Update(dt float64) {
	if g.actions.Held[ActionLeft] {
		g.cam.X--
	}
	if g.actions.Held[ActionRight] {
		g.cam.X++
	}
	if g.actions.Held[ActionUp] {
		g.cam.Y--
	}
	if g.actions.Held[ActionDown] {
		g.cam.Y++
	}
}

func (g *Game) Draw(r *render.Renderer) {
	g.m.Draw(r, 0, 0, g.cam)
}

func (g *Game) Resize(w, h int) { g.cam.SetViewport(w, h) }

func (g *Game) ActionMap() input.ActionMap {
	return input.ActionMap{
		ActionLeft:  "key:left",
		ActionRight: "key:right",
		ActionUp:    "key:up",
		ActionDown:  "key:down",
	}
}

func (g *Game) UpdateActionState(state input.ActionState) { g.actions = state }

func main() {
	m, err := tilemap.LoadFromFiles("path/to/world.map", "path/to/world.tiles")
	if err != nil {
		log.Fatal(err)
	}
	game := &Game{
		m:   m,
		cam: camera.NewBasic(0, 0, 0, 0),
	}
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
```
