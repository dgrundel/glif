# glif
2d terminal game engine

## Structure

- `engine`: main loop, timing, event dispatch
- `term`: terminal backend + framebuffer diffing (tcell)
- `grid`: core types (`Vec2i`, `Cell`, `Frame`)
- `render`: draw helpers and sprite support
- `assets`: sprite loaders (masked sprites + palettes)
- `ecs`: lightweight component/system world + z-ordered sprite rendering
- `demos/duck`: demo scene
- `utils/genmask`: sprite mask generator

## Demo

Run the duck demo:

```
go run ./demos/duck
```

## Sprite assets

Masked sprites use three files with a shared base name:

- `<name>.sprite`
- `<name>.mask`
- `<name>.palette` (optional)

If `<name>.palette` is missing, `default.palette` in the same folder is used.

Load with:

```
assets.LoadMaskedSprite("path/to/name")
```

## Mask generator

Generate a mask from a sprite:

```
go run ./utils/genmask path/to/sprite.sprite
```
