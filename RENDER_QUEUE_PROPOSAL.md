# Render Queue / Draw Commands

## Summary
Add an explicit render-queue API that records draw commands and then executes them in a single render pass. This enables batching, post-processing, and consistent ordering without requiring each system to draw directly to the frame.

## Why This Is Useful
- **Separation of concerns**: systems can emit draw intents instead of writing directly to the frame.
- **Deterministic ordering**: centralized ordering by Z-index or layer across all systems.
- **Potential performance wins**: enables command batching and future optimizations (e.g., grouping by sprite/palette).

## Feature Overview
- Introduce a `render.Command` type (or interface) and a `render.Queue` to collect commands.
- Provide convenience helpers: `Queue.DrawSprite`, `Queue.DrawText`, `Queue.Rect`, etc.
- Add `Renderer.Flush(queue)` that sorts and executes commands onto the frame.

## Fit With Existing Architecture
- **ECS**: systems can push commands to the queue instead of drawing directly.
- **Renderer**: remains the single place that writes to `grid.Frame`.
- **Engine**: stays mostly unchanged; just swaps `game.Draw(r)` to something like `game.Draw(q)` then `renderer.Flush(q)`.

## Proposed API
### render package (new)
```go
// CommandType is the kind of draw command.
type CommandType int

const (
	CmdSprite CommandType = iota
	CmdText
	CmdRect
	CmdHLine
	CmdVLine
)

// Command is a single draw intent.
type Command struct {
	Type   CommandType
	Z      int
	X, Y   int
	W, H   int
	Text   string
	Sprite *Sprite
	Style  grid.Style
	Rect   RectOptions
	Line   LineOptions
}

// Queue collects commands for a frame.
type Queue struct {
	Commands []Command
}

// NewQueue creates a new queue with optional initial capacity.
func NewQueue(capacity int) *Queue

// Clear resets the queue without releasing capacity.
func (q *Queue) Clear()

// Len returns the number of queued commands.
func (q *Queue) Len() int

// Draw helpers.
func (q *Queue) DrawSprite(x, y, z int, s *Sprite)
func (q *Queue) DrawText(x, y, z int, text string, style grid.Style)
func (q *Queue) Rect(x, y, w, h, z int, style grid.Style, opts ...RectOptions)
func (q *Queue) HLine(x, y, length, z int, style grid.Style, opts ...LineOptions)
func (q *Queue) VLine(x, y, length, z int, style grid.Style, opts ...LineOptions)
```

### render package (renderer integration)
```go
// Flush draws all commands to the current frame.
func (r *Renderer) Flush(q *Queue)
```

Behavior:
- `Flush` sorts by `Z` and stable order of insertion for ties.
- `Queue` is re-used per frame and can be cleared after flush.

## Implementation Plan
1. Add `render/queue.go` with `Command` and `Queue` types + helper methods.
2. Add `Renderer.Flush(q)` that sorts commands and executes them using existing draw calls.
3. Add a simple demo or update one demo to use the queue (optional).

## Alternatives Considered
1. **Keep direct rendering**: lowest complexity but no centralized ordering or batching.
2. **ECS render system only**: keep draw calls but add a single ECS render system; less general than a queue for non-ECS demos.
3. **Immediate-mode with Z stack**: add `Renderer` Z stack without a queue; keeps immediacy but still disperses logic.

## Risks / Tradeoffs
- Adds memory allocation per frame (commands). Can be mitigated by reusing slices.
- Slightly more complex for simple demos.

## Testing Plan
- Unit test: queue sorting and stable ordering.
- Manual: update a demo to render via queue and confirm Z layering matches current behavior.

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
	q := render.NewQueue(4)
	q.DrawSprite(2, 2, 0, g.sprite)
	q.DrawText(2, 0, 10, "queue demo", r.Frame.Clear.Style)
	r.Flush(q)
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
