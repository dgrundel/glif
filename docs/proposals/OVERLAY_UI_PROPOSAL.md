# Overlay / Menu Rendering System

## Summary
Add a lightweight overlay/UI layer that renders in **screen space**, independent of the camera transform. This enables HUDs, menus, and modal overlays (inventory, settings) to draw on top of the world without fighting camera math.

## Why This Is Useful
- **World + UI separation**: the world uses camera transforms; UI should not.
- **Ease of use**: menus and HUDs become simple to render, no manual unprojection.
- **Layering**: UI reliably appears above world content.

## Feature Overview
- A **screen-space renderer** that bypasses camera transforms.
- A small **overlay stack** for draw order (HUD, modal, tooltips).
- Optional **UI draw callbacks** separate from world draw callbacks.

## Fit With Existing Architecture
- **Renderer** already owns drawing to a `grid.Frame`.
- **Camera proposal** applies to world-space draw operations; overlays should opt out.
- **Engine** can call both world and overlay draw passes, keeping game code clean.

## Proposed API
### render package (new renderer view)
```go
// Screen returns a renderer view that ignores the camera.
// It shares the same frame as the parent renderer.
func (r *Renderer) Screen() *Renderer
```

### engine package (optional interfaces)
```go
type OverlayAware interface {
	DrawOverlay(r *render.Renderer)
}

type MenuAware interface {
	DrawMenu(r *render.Renderer)
}
```

Engine behavior (conceptual):
```go
r.Clear()

game.Draw(r) // world-space, camera-aware

if o, ok := game.(OverlayAware); ok {
	o.DrawOverlay(r.Screen())
}
if m, ok := game.(MenuAware); ok {
	m.DrawMenu(r.Screen())
}
```

### render package (optional overlay stack helpers)
```go
type Overlay struct {
	Z    int
	Draw func(r *render.Renderer)
}

type OverlayStack struct {
	Overlays []Overlay
}

func NewOverlayStack(capacity int) *OverlayStack
func (s *OverlayStack) Clear()
func (s *OverlayStack) Add(z int, draw func(r *render.Renderer))
func (s *OverlayStack) DrawAll(r *render.Renderer)
```

This lets games register multiple overlays (HUD, dialog, tooltip) with simple z ordering.

## Developer Experience
- **World draw stays the same** (uses camera).
- **Overlay draw is simple**: `DrawOverlay(r.Screen())` to draw at fixed screen positions.
- **Menus** can use render primitives (`Rect`, `Text`, `Line`) without any transform math.

## Alternatives Considered
1. **UI always uses camera coordinates**: forces developers to convert, error-prone.
2. **Separate screen frame**: render UI to a second frame and merge; more complex.
3. **Full UI toolkit**: out of scope; heavier than needed right now.

## Risks / Tradeoffs
- Adds another draw pass; minimal overhead in practice.
- Requires clear conventions so devs know which draw pass to use.

## Testing Plan
- Manual test: add a HUD to `demos/world` showing camera coordinates.
- Manual test: add a modal box to `demos/ski` that remains fixed while the world scrolls.

## Appendix: Minimal Demo
```go
package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
)

const ActionQuit input.Action = "quit"

type Game struct {
	sprite  *render.Sprite
	actions input.ActionState
}

func (g *Game) Update(dt float64) {}

func (g *Game) Draw(r *render.Renderer) {
	r.DrawSprite(2, 2, g.sprite) // world space
}

func (g *Game) DrawOverlay(r *render.Renderer) {
	r.Rect(0, 0, 20, 3, r.Frame.Clear.Style, render.RectOptions{Fill: true})
	r.DrawText(1, 1, "HUD: 100%", r.Frame.Clear.Style)
}

func (g *Game) Resize(w, h int) {}

func (g *Game) ActionMap() input.ActionMap {
	return input.ActionMap{ActionQuit: "key:esc"}
}

func (g *Game) UpdateActionState(state input.ActionState) { g.actions = state }

func main() {
	sprite := assets.MustLoadSprite("path/to/sprite")
	game := &Game{sprite: sprite}
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
```
