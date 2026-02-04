# TODO

## Not implemented yet

- **Camera / world-to-screen transform**
  - `camera` package exists, but it isn't wired as a standard world-to-screen projection layer.
  - Need consistent world space + camera viewport + projection/clipping across demos.

- **Render queue / draw commands**
  - Rendering is direct (ECS draw with Z sort).
  - No explicit render command queue or command batching.

## Implemented differently than initial notes

- **ECS vs scene**
  - We moved directly to a minimal ECS (`ecs` package) rather than a `scene` interface.
  - Entities are IDs with component maps; systems are lightweight functions.
