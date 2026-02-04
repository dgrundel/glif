# TODO

This captures items from `chat.md` that are not implemented yet, plus areas where the implementation differs from the original architectural notes.

## Not implemented yet

- **Camera / world-to-screen transform**
  - `camera` package exists, but it isn't wired as a standard world-to-screen projection layer.
  - Need consistent world space + camera viewport + projection/clipping across demos.

- **Render queue / draw commands**
  - Rendering is direct (ECS draw with Z sort).
  - No explicit render command queue or command batching.

- **Tile map collision layers**
  - Tile maps exist, but there is no collision/grid layer support yet.

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
