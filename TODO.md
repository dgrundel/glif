# TODO

This captures items from `chat.md` that are not implemented yet, plus areas where the implementation differs from the original architectural notes.

## Not implemented yet

- **Camera / world-to-screen transform**
  - No `Camera` type or projection; positions are treated as screen coordinates.
  - Need world space + camera viewport + projection/clipping.

- **Render primitives beyond sprites/text**
  - Only `DrawText` and `DrawSprite` exist.
  - Missing helpers like `FillRect`, `Line`, etc.

- **Render queue / draw commands**
  - Rendering is direct (ECS draw with Z sort).
  - No explicit render command queue or command batching.

- **Sprite animation / sprite sheets**
  - Assets loader supports masked sprites only.
  - No multi-frame sprites or animation system.

- **Collision / tile maps**
  - No tile map layer or collision system (grid or AABB).

- **Action mapping on input**
  - `input` package is generic key state (held/pressed).
  - No action mapping layer (e.g., move/quit/interact bindings).

- **Fixed timestep loop**
  - `engine` uses variable `dt` from wall clock.
  - No fixed-step accumulator or decoupled render/update.

- **Terminal UI extras**
  - Alt-screen/cursor hide were mentioned, but not explicitly implemented.
  - tcell handles much of this implicitly; no explicit API in `term`.

## Implemented differently than initial notes

- **ECS vs scene**
  - We moved directly to a minimal ECS (`ecs` package) rather than a `scene` interface.
  - Entities are IDs with component maps; systems are lightweight functions.

- **Diff strategy**
  - Current `term.Present` is cell-by-cell diffing.
  - Suggested row-run diffing / hashing are not implemented.
