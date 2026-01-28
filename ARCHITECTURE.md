# Architecture

This engine is a terminal-native 2D renderer built around a simple pipeline:

**World → Camera → Framebuffer → Terminal**

Everything ultimately becomes a 2D grid of cells (rune + style) that is diffed and presented to the terminal.

## Core concepts

### Cells and frames
A **cell** is the smallest render unit. It stores a character and its visual style. A **frame** is a 2D array of cells.

Proposed types:

```
type Style struct {
  Fg Color
  Bg Color
  Bold bool
}

type Cell struct {
  Ch rune
  Style Style
}

type Frame struct {
  W, H int
  Cells []Cell // len = W*H
}
```

Use index `i := y*W + x`.

Maintain two frames:
- `front`: the last frame shown on screen
- `back`: the new frame you just drew

Diff `back` vs `front` and only emit terminal updates for changed cells. After presenting, swap or copy `back` → `front`.

### Grid and coordinates
Keep a small, shared grid utility layer:
- `Vec2i`, `Rect`, bounds helpers
- Screen-space vs world-space

World coordinates can be integer-based. If smooth movement is needed later, world can be float and rounded during projection.

### Sprite
Sprites are just 2D arrays of cells. Transparency can be represented by a sentinel rune (e.g. `Ch == 0`).

```
type Sprite struct {
  W, H int
  Cells []Cell
  Transparent rune // optional, or use Ch==0
}
```

### Camera
The camera defines which part of the world appears on screen.

```
type Camera struct {
  Pos Vec2i     // world top-left of view
  ViewW, ViewH int
}
```

`WorldToScreen` projects world coordinates into screen coordinates:

```
func (c Camera) WorldToScreen(p Vec2i) Vec2i {
  return Vec2i{p.X - c.Pos.X, p.Y - c.Pos.Y}
}
```

When the terminal resizes, the camera view and framebuffer resize together.

## Rendering pipeline

Per frame:
1. `back.Clear()` (fill with spaces + default style)
2. For each renderable object:
   - Convert world position to screen position via camera
   - Clip to viewport
   - Draw into `back` using render helpers (text, sprite, rect, line, etc.)
3. `term.Present(back)`
   - Diff `back` vs `front`
   - Emit minimal cursor moves + writes
   - Swap buffers

### Diff strategy
For terminal rendering, a row-based diff is effective and predictable.

Recommended algorithm:
- Scan each row left to right
- When a change is found, emit a run of changed cells
- Break runs on style changes

Optional optimizations:
- Dirty-row or dirty-rect tracking
- Per-row hashing to skip unchanged rows

Avoid LCS/Myers-style diffs; they are not aligned with 2D grids and style boundaries.

## Modules / packages

A minimal package split that scales well:

1. `term`
   - Enter/exit alt screen
   - Hide/show cursor
   - Terminal size
   - Input events
   - Efficient present/diff

2. `grid`
   - `Vec2i`, `Rect`
   - `Cell`, `Frame`

3. `render`
   - `DrawText`, `DrawSprite`, `FillRect`, `Line`
   - Optional render queue / z-layering

4. `assets`
   - Load sprites from files
   - Optional animation frames

5. `scene` (or `ecs` later)
   - Entity management
   - Update order

6. `engine`
   - Main loop
   - Timing
   - Glue code

## Update loop

Start with a fixed timestep; render can run at the same rate.

```
tick := time.Second / 60
for running {
  start := time.Now()

  events := term.PollEventsNonBlocking()
  world.HandleInput(events)

  world.Update(tick)
  renderer.Draw(world, camera, back)

  term.Present(back)
  sleepRemaining(tick - time.Since(start))
}
```

Rendering and update can be decoupled later, but not initially.

## Input and actions

Translate raw keys into semantic actions to keep gameplay code clean:
- `MoveUp`, `MoveDown`, `Interact`, `Quit`

This allows rebinding and future mouse support without touching game logic.

## ECS vs simple entities

Start with a simple interface:

```
type Entity interface {
  Update(dt time.Duration)
  Draw(r *Renderer, cam Camera)
}
```

Migrate to ECS if you start accumulating lots of entity types or complex combinations. ECS model:
- Entities are IDs
- Components are data structs
- Systems operate over entities with a set of components

## Collision and tile maps

If using tiles:
- Store tile map as its own layer
- Collisions can be grid-based + AABB for entities

The camera + render pipeline stays unchanged.

## Notes on terminal tech

`tcell` is a solid default for raw mode, resize, and input. If you later need total control over output, you can still keep `tcell` for input and write your own diff/present logic on top of its screen interface.
