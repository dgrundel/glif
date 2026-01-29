# Architecture

This engine is a terminal-native 2D renderer built around a simple pipeline:

**World/ECS → Framebuffer → Terminal**

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
There is no camera yet. When added, it will project world coordinates into screen space and clip to the viewport.

## Rendering pipeline

Per frame:
1. `back.Clear()` (fill with spaces + default style)
2. For each renderable object:
   - Draw into `back` using render helpers (text, sprite)
3. `term.Present(back)`
   - Diff `back` vs `front`
   - Emit minimal cursor updates
   - Copy changed cells into `front`

### Diff strategy
The current implementation is a simple cell-by-cell diff:
- Compare `back` vs `front`
- For each changed cell, call `SetContent`
- Copy the changed cell into `front`

This is intentionally straightforward. A row-run diff or dirty-region tracking can be added later for fewer writes.

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
   - `DrawText`, `DrawSprite`
   - Sprite transparency support (`Ch == 0`)

4. `assets`
   - Load masked sprites from files
   - Mask/palette conventions for color and transparency

5. `ecs`
   - Minimal ECS world (positions, velocities, sprites)
   - Built-in movement system
   - Z-ordered rendering

6. `engine`
   - Main loop
   - Timing
   - Glue code

## Update loop

The `engine.Engine` owns the loop. It uses a ticker for update cadence and a select to handle input events.

```
type Game interface {
  Update(dt float64)
  Draw(r *render.Renderer)
  HandleEvent(ev tcell.Event) (quit bool)
  Resize(w, h int)
}
```

Per tick:
- compute `dt` from wall clock
- `game.Update(dt)`
- `renderer.Clear()` then `game.Draw(renderer)`
- `screen.Present(frame)`

Events:
- resize -> resize frame + notify game
- other -> `game.HandleEvent`

## Input and actions

Translate raw keys into semantic actions to keep gameplay code clean:
- `MoveUp`, `MoveDown`, `Interact`, `Quit`

This allows rebinding and future mouse support without touching game logic.

## ECS vs simple entities

Current approach uses a minimal ECS world in `ecs`:

```
type World struct {
  Positions map[Entity]*Position
  Velocities map[Entity]*Velocity
  Sprites map[Entity]*SpriteRef
}
```

Entities are IDs, components are data structs, and systems iterate over matching component sets.

## Collision and tile maps

If using tiles:
- Store tile map as its own layer
- Collisions can be grid-based + AABB for entities

The camera + render pipeline stays unchanged.

## Notes on terminal tech

`tcell` is a solid default for raw mode, resize, and input. The engine builds its own frame buffer and performs diffing in `term.Screen.Present`, so you can evolve the diff strategy independently of input handling.

## Asset conventions

Masked sprites are loaded by base path:
- `duck.sprite`
- `duck.mask`
- `duck.palette` (optional)

If `<name>.palette` is missing, `default.palette` in the same folder is used.
