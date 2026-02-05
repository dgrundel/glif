# Camera World-to-Screen Transform

## Summary
Introduce a consistent, engine-level world-to-screen transform so demos and systems can operate in world space while rendering and input operate in screen space. The `camera` package already exists; this proposal wires it into the render and input paths in a minimal, opt-in way that can be expanded later. The **game/scene owns the camera** and moves it; the **renderer only asks the camera to project** during draw calls.

## Why This Is Useful
- **Consistency across demos**: right now each demo handles positioning differently, which leads to duplicated math and bugs.
- **World space**: enables large maps and panning without rewriting every draw call.
- **Separation of concerns**: game logic stays in world coordinates; rendering and input are projected by the camera.

## Feature Overview
- Provide a single camera transform that converts world coordinates to screen coordinates for render operations.
- The camera defines a viewport (screen size) and a world-space origin (top-left).
- Add lightweight clipping so that render operations outside the viewport are ignored early.
- Camera **ownership stays with the game/scene**; renderer receives a reference per frame or via a view.

## Fit With Existing Architecture
- **Engine** owns the screen size and resize events; it can pass viewport dimensions into the camera.
- **Renderer** already writes to a `grid.Frame`; this is the right place to apply the transform.
- **Camera** package provides the transformation math and holds camera state (position, viewport size).
- **Game/scene** owns and updates the camera; renderer only consumes it.

## Proposed API
### camera package (new/updated)

```go
// Camera is the minimal interface used by render.
type Camera interface {
	WorldToScreen(x, y int) (sx, sy int)
	ScreenToWorld(x, y int) (wx, wy int)
	Visible(x, y, w, h int) bool
	SetViewport(w, h int)
}

// Basic is the default camera implementation.
// X,Y are the world-space coordinates of the top-left screen pixel.
type Basic struct {
	X int
	Y int
	W int
	H int
}

// NewBasic creates a camera with the given top-left origin and viewport size.
func NewBasic() *Basic

// Set sets camera origin directly.
func (c *Basic) Set(x, y int)

// Move offsets the camera origin by dx, dy.
func (c *Basic) Move(dx, dy int)

// SetViewport updates the viewport size.
func (c *Basic) SetViewport(w, h int)

// WorldToScreen converts world coords to screen coords.
func (c *Basic) WorldToScreen(x, y int) (sx, sy int)

// ScreenToWorld converts screen coords to world coords.
func (c *Basic) ScreenToWorld(x, y int) (wx, wy int)

// Visible reports whether a world-space rect intersects the viewport.
func (c *Basic) Visible(x, y, w, h int) bool
```

### render package (minimal, opt-in)
```go
func (r *Renderer) SetCamera(cam camera.Camera)
func (r *Renderer) Camera() camera.Camera

// WithCamera returns a renderer view that uses the given camera for projection.
// The camera remains owned by the caller; the renderer does not manage its lifetime.
func (r *Renderer) WithCamera(cam camera.Camera) *Renderer
```

Behavior:
- If no camera is set, rendering behaves exactly as today (screen space).
- If a camera is set, all `Renderer` draw methods convert coordinates using `WorldToScreen` and skip if not visible.
- `WithCamera` is the preferred usage when the game/scene owns the camera and wants to avoid “renderer ownership.”

### engine integration (behavior only)
- On resize: if a camera is set on the renderer, call `cam.SetViewport(w, h)`.
  (This is done in `engine.Run` alongside the existing resize handling.)

## Implementation Plan
1. **Camera API**: ensure `camera` package exposes `WorldToScreen`, `Visible`, and `SetViewport`.
2. **Renderer hookup**: add optional camera field to `render.Renderer`, update draw methods to transform coordinates.
3. **Clipping**: in each draw operation, check `Visible` to avoid unnecessary writes.
4. **Demo adoption**: update demos that pan or scroll (e.g. world, wasd) to use the camera.

## Alternatives Considered
1. **Do nothing / per-demo transforms**: simplest but keeps duplication and inconsistencies.
2. **Scene graph approach**: add a full scene graph and attach a camera node. More powerful but heavier than needed right now.
3. **Transform in game logic**: each game pre-transforms positions before draw calls. Minimal changes but repeats math in every demo.

## Risks / Tradeoffs
- Adds an extra conditional to each draw call (camera vs no camera).
- Clipping logic needs to be careful with sprites and lines to avoid partial draws being skipped incorrectly.

## Testing Plan
- Add a small unit test for camera transforms (world->screen) and visibility.
- Manual verification in `demos/world` and `demos/wasd` for smooth panning and clipping.

## Appendix: Minimal Demo
Simple example showing a camera-aware world draw plus an optional overlay.

```go
package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
)

const (
	ActionLeft  input.Action = "left"
	ActionRight input.Action = "right"
	ActionUp    input.Action = "up"
	ActionDown  input.Action = "down"
)

type Game struct {
	sprite  *render.Sprite
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
	rc := r.WithCamera(g.cam)       // game owns camera, renderer just projects
	rc.DrawSprite(10, 5, g.sprite)  // world space
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
	sprite := assets.MustLoadSprite("path/to/sprite")
	game := &Game{
		sprite: sprite,
		cam:    camera.NewBasic(),
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
