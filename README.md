# glif
2d terminal game engine

## Structure

- `engine`: main loop, timing, event dispatch
- `term`: terminal backend + framebuffer diffing (tcell)
- `grid`: core types (`Vec2i`, `Cell`, `Frame`)
- `render`: draw helpers and sprite support
- `assets`: sprite loaders (masked sprites + palettes)
- `ecs`: lightweight component/system world + z-ordered sprite rendering
- `input`: generic key-state tracking (held/pressed) with TTL
- `demos/duck`: demo scene
- `demos/wasd`: demo with keyboard movement
- `demos/world`: demo with camera panning and tilemap background
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

Palette colors support:
- Hex RGB (`#RRGGBB` or `#RGB`)
- Named colors supported by tcell

Example:

`duck.sprite`:
```
>o)
(_>
```

`duck.mask`:
```
gww
wwg
```

`duck.palette`:
```
# key fg bg [bold] [transparent]
w #ffffff #0000ff
g #ffd700 #0000ff
. reset reset transparent
```

## Tile maps

Tile maps can be loaded from a text map and a tiles file:

```
tilemap.LoadFromFiles("path/to/level.map", "path/to/level.tiles")
```

Example `.tiles`:
```
~ water
. empty
```

Example `.map`:
```
~~~~..~~
~..~~..~
```

Sprite names in `.tiles` are base paths (same as `assets.LoadMaskedSprite`).

## Mask generator

Generate a mask from a sprite:

```
go run ./utils/genmask path/to/sprite.sprite
```
