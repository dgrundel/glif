# Scene API

## Summary
Introduce a first‑class Scene API to manage game states (menu, gameplay, pause, etc.) with clear lifecycle hooks and a scene stack. This gives a consistent way to switch between modes without scattering state management across demos.

## Why This Is Useful
- **State management**: cleanly separate menu, gameplay, pause, and overlays.
- **Lifecycle hooks**: consistent initialization and teardown when scenes switch.
- **Stack support**: push a pause/menu scene without losing the underlying game.

## Feature Overview
- A `Scene` interface with lifecycle methods.
- A `SceneStack` that owns update/draw routing.
- Optional hooks for resize and input/action state.

## Fit With Existing Architecture
- The engine already drives `Update`, `Draw`, and `Resize`; a SceneStack can implement the same interfaces and become the engine’s `Game`.
- Action mapping fits naturally: the active scene provides its action map and receives action state each frame.
- Renderer default is screen‑space; scenes can opt into world‑space via `r.WithCamera(cam)`.

## Proposed API (exhaustive)

### scene package
```go
// Scene is a game state (menu, gameplay, pause, etc.).
type Scene interface {
	Update(dt float64)
	Draw(r *render.Renderer)
}

// Optional hooks (reuse engine interfaces).
// - engine.Resizer
// - engine.ActionAware
// - engine.InputAware

// Lifecycle hooks.
// OnEnter receives the owning stack so scenes can push/pop/replace directly.
type Enterer interface { OnEnter(stack *Stack) }
type Exiter interface { OnExit() }
type Suspender interface { OnSuspend() }
type Resumer interface { OnResume() }

// SceneStack manages scene routing.
type Stack struct {
	scenes []Scene
}

func NewStack() *Stack
func (s *Stack) Push(scene Scene)
func (s *Stack) Pop() Scene
func (s *Stack) Replace(scene Scene)
func (s *Stack) Top() Scene
func (s *Stack) Len() int

// Engine‑compatible methods.
func (s *Stack) Update(dt float64)
func (s *Stack) Draw(r *render.Renderer)
func (s *Stack) Resize(w, h int)

// Input/action routing helpers (optional):
func (s *Stack) ActionMap() input.ActionMap
func (s *Stack) UpdateActionState(state input.ActionState)
func (s *Stack) SetInput(state input.State)
```

### Behavior
- **Update/Draw**: routed to the top scene only.
- **Resize**: forwarded to all scenes that implement `Resizer` (or only top; see alternatives).
- **Push**: calls `OnSuspend` on the current top scene (if implemented), then `OnEnter` on the new scene.
- **Pop**: calls `OnExit` on the popped scene, then `OnResume` on the new top (if present).
- **Replace**: calls `OnExit` on the current top, then `OnEnter` on the new scene.

### Lifecycle Summary
- **OnEnter**: scene becomes active for the first time (or after replace).
- **OnExit**: scene is removed permanently.
- **OnSuspend**: scene is covered by another scene but remains on the stack.
- **OnResume**: scene becomes active again after being suspended.

## Implementation Plan
1. Add `scene/stack.go` with the `Stack` implementation.
2. Keep existing `scene.Scene` (or replace with interfaces above).
3. Update engine usage in demos to pass a `scene.Stack` as the `Game`.

## Alternatives Considered
1. **Single Scene struct (current)**: simpler but lacks lifecycle and stack semantics.
2. **ECS only**: manage states as ECS systems; flexible but heavier for small demos.
3. **Scene graph**: full tree of nodes; overkill for simple state switching.

## Risks / Tradeoffs
- Scene stacks add indirection; debugging needs good logging.
- Decisions about resize routing (top‑only vs all) can impact hidden scenes.

## Testing Plan
- Manual: push/pull pause scene without breaking gameplay state.
- Unit: ensure lifecycle hooks fire in the correct order.

## Appendix: Sample Code
```go
package main

import (
	"log"

	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/scene"
)

type MenuScene struct{}
func (m *MenuScene) Update(dt float64) {}
func (m *MenuScene) Draw(r *render.Renderer) { r.DrawText(2, 2, "MENU", r.Frame.Clear.Style) }

func (m *MenuScene) OnEnter(stack *scene.Stack) {
	_ = stack
}

func (m *MenuScene) ActionMap() input.ActionMap { return input.ActionMap{"start": "key:enter"} }
func (m *MenuScene) UpdateActionState(state input.ActionState) {}

func main() {
	stack := scene.NewStack()
	stack.Push(&MenuScene{})

	eng, err := engine.New(stack, 0)
	if err != nil { log.Fatal(err) }
	if err := eng.Run(stack); err != nil { log.Fatal(err) }
}
```
