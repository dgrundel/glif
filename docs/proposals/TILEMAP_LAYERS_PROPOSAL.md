# Tilemap Layers

## Summary
Add support for multiple tilemap layers with simple ordering and optional per‑layer camera settings. This enables backgrounds, mid‑ground details, and overlays like fog or UI‑style grids while keeping tilemaps easy to work with.

## Why This Is Useful
- **Visual richness**: layered worlds (background terrain + props + foreground).
- **Parallax**: slow‑moving background layers for depth.
- **UI reuse**: tilemap rendering can be used for menus, inventories, or minimaps.

## Feature Overview
- Introduce a `Layer` type that wraps a `tilemap.Map` plus layer settings.
- Provide a `LayeredMap` (or `Stack`) that draws layers in order.
- Allow per‑layer camera overrides (optional) to support parallax.

## Fit With Existing Architecture
- The camera proposal provides a clean interface for world‑to‑screen conversion.
- The overlay proposal uses screen‑space renderers; a layer can opt to render in screen space if desired.

## Proposed API (exhaustive)
### tilemap package
```go
// Layer describes a tilemap and how it should be rendered.
type Layer struct {
	Map       *Map
	Z         int
	OffsetX   int
	OffsetY   int
	ParallaxX float64
	ParallaxY float64
	Camera    camera.Camera // optional override
	Screen    bool          // if true, ignore camera and draw in screen space
}

// LayeredMap draws multiple layers in order.
type LayeredMap struct {
	Layers []Layer
}

func NewLayeredMap(capacity int) *LayeredMap
func (lm *LayeredMap) Add(layer Layer)
func (lm *LayeredMap) Clear()
func (lm *LayeredMap) Draw(r *render.Renderer, cam camera.Camera)
```

### Behavior details
- Layers are drawn in `Z` order (stable for equal Z).
- `OffsetX/OffsetY` shift the layer in world space.
- `ParallaxX/ParallaxY` scale camera movement for the layer (1.0 = normal, 0.5 = slower).
- If `Screen` is true, the layer ignores `cam` and draws in screen space (useful for UI grids).
- If `Camera` is set on the layer, it overrides the shared `cam` argument.

## Implementation Plan
1. Add `Layer` + `LayeredMap` types in `tilemap`.
2. Implement ordered draw with parallax offsets:
   - compute layer camera offset = base camera origin * parallax
   - apply layer offsets
3. Use `Map.Draw` for each layer, passing either the shared camera or the layer override.
4. Update or add a demo (world) to show background + water + foreground.

## Alternatives Considered
1. **Manual layering in demos**: simple but repetitive and error‑prone.
2. **Render queue only**: you can push tilemap draw calls into a queue; still need a layer abstraction for parallax and ordering.
3. **Multiple maps in game code**: similar to #1, but without a helper API for ordering and parallax.

## Risks / Tradeoffs
- Adds another abstraction layer to tilemaps, which may be overkill for simple games.
- Parallax math introduces floats (can be kept internal to avoid leaking into game code).

## Testing Plan
- Manual: add a background layer moving at 0.5x speed and verify parallax.
- Manual: use a screen‑space layer for a grid overlay to ensure it ignores camera movement.

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
)

type Game struct {
	layers  *tilemap.LayeredMap
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
}

func (g *Game) Draw(r *render.Renderer) {
	g.layers.Draw(r, g.cam)
}

func (g *Game) Resize(w, h int) { g.cam.SetViewport(w, h) }

func (g *Game) ActionMap() input.ActionMap {
	return input.ActionMap{
		ActionLeft:  "key:left",
		ActionRight: "key:right",
	}
}

func (g *Game) UpdateActionState(state input.ActionState) { g.actions = state }

func main() {
	bg, err := tilemap.LoadFromFiles("path/to/bg.map", "path/to/bg.tiles")
	if err != nil {
		log.Fatal(err)
	}
	fg, err := tilemap.LoadFromFiles("path/to/fg.map", "path/to/fg.tiles")
	if err != nil {
		log.Fatal(err)
	}

	layers := tilemap.NewLayeredMap(2)
	layers.Add(tilemap.Layer{
		Map:       bg,
		Z:         0,
		ParallaxX: 0.5,
		ParallaxY: 0.5,
	})
	layers.Add(tilemap.Layer{
		Map: fg,
		Z:   10,
	})

	game := &Game{
		layers: layers,
		cam:    camera.NewBasic(0, 0, 0, 0),
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
