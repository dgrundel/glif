# Tilemap Camera Integration

## Summary
Align tilemap rendering with the renderer‑centric camera model. Tilemaps should draw in world space and let the renderer apply camera transforms via `r.WithCamera(cam)`. This removes duplicated projection logic from `tilemap` and keeps the camera contract centralized in `render`.

## Why This Is Useful
- **Single source of truth** for world‑to‑screen math (the renderer + camera).
- **Consistency** with the renderer camera model (world‑space draw via `WithCamera`).
- **Simpler API** for tilemaps: no camera parameter, no projection logic.

## Feature Overview
- Remove camera/projection logic from `tilemap`.
- Tilemap draws in world space; the renderer view handles camera transforms.
- Optional: add culling helpers that take a camera to skip off‑screen tiles, but keep them separate from draw.

## Fit With Existing Architecture
- The renderer applies camera transforms via `WithCamera`.
- Tilemap remains a world‑space data structure and doesn’t own camera logic.

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
// The renderer is responsible for camera transforms.
func (m *Map) Draw(r *render.Renderer, worldX, worldY float64)
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

### render package (expected usage)
```go
// WithCamera returns a renderer view that applies the camera transform.
func (r *Renderer) WithCamera(cam camera.Camera) *Renderer
```

## Implementation Plan
1. **API change**: remove camera parameter from `Map.Draw` and keep world coords as float64.
2. **Projection**: delete camera usage from tilemap; rely on `r.WithCamera(cam)` at call sites.
3. **Optional culling**: add a separate helper if we want camera‑based bounds later.
4. **Demo updates**: update `demos/world` to use `r.WithCamera(cam)` before drawing the map.

## Alternatives Considered
1. **Keep camera parameter on tilemap**: duplicates projection logic and can drift from renderer behavior.
2. **Separate map renderer**: a dedicated tilemap renderer that owns its own camera logic; heavier and redundant.

## Risks / Tradeoffs
- API change will require updating any demo code using tilemaps.
- Tilemap culling becomes optional/explicit; without it, large maps may be slower to draw.

## Testing Plan
- Manual regression test in `demos/world` for smooth panning without column “sticking.”
- Unit test for `Map.Draw` world‑space positioning (optional).

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
		g.cam.Move(-1, 0)
	}
	if g.actions.Held[ActionRight] {
		g.cam.Move(1, 0)
	}
	if g.actions.Held[ActionUp] {
		g.cam.Move(0, -1)
	}
	if g.actions.Held[ActionDown] {
		g.cam.Move(0, 1)
	}
	if g.m != nil {
		g.cam.ClampTo(g.m.WorldBounds())
	}
}

func (g *Game) Draw(r *render.Renderer) {
	rc := r.WithCamera(g.cam)
	g.m.Draw(rc, 0, 0)
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
		cam: camera.NewBasic(),
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
