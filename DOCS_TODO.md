# Docs Todo

## 1) Home
- Overview: what glif is and the design goals
- Quickstart snippet (load sprite, engine loop, quit handling)
- Key concepts summary (sprite, palette, animation, collision)

## 2) Getting Started
- Installation / requirements
- Running demos (ski, blackjack, animation, collision)
- Project structure overview

## 3) Core Engine
- Game interface (Update/Draw/Resize)
- InputAware / Quitter interfaces
- Timing loop (fixed step, default 60 fps)
- Frame clear behavior
- FPS overlay (ShowFPS)

## 4) Rendering
- Renderer basics (DrawSprite, DrawText)
- Primitives (Rect/HLine/VLine + options)

## 5) Assets & Palettes
- Sprite files: .sprite, .color, .palette, .collision
- Color mask rules (space/`.` is transparent)
- Palette file format and `//` comments
- Color formats (hex and named)
- `palette.MustLoad` / `MustStyle`
- `assets.LoadSprite` / `assets.MustLoadSprite`

## 6) Animation
- File naming: `<base>.<name>.animation`
- Size constraints (width match, height multiple)
- `.animation.color` behavior
- API: `sprite.LoadAnimation(name)`, `anim.Play(fps)`, `player.Update(dt)`

## 7) Collision
- `.collision` format (non-space and not `.` = collidable)
- `collision.Overlaps(ax, ay, a, bx, by, b)`

## 8) Tilemaps & Camera
- `.map` / `.tiles` formats
- `tilemap.LoadFromFiles`
- Camera basics

## 9) Tools
- `genmask` usage (color/collision flags, mapping, animation output)
- `spritepreview` usage (files/folders, controls)

## 10) Demos
- Ski
- Blackjack
- Animation
- Collision
- Duck / WASD / World (if still relevant)

## 11) Troubleshooting
- Missing palette key
- Sprite/color size mismatch
- Collision size mismatch
- Animation size mismatch

## 12) Contributing
- Code style (gofmt)
- Where to add assets/demos
- Build/check steps
