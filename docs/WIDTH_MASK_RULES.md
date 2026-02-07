# Width Mask Rules

This document describes how optional sprite width masks should work.

## Files
A width mask is an optional file alongside a sprite:

- `<name>.sprite`
- `<name>.color`
- `<name>.width` (optional)

If `.width` is missing, all glyphs are treated as width 1.

## Shape Rules
- `.sprite` and `.width` are aligned by **row and column position**.
- Both files may have **ragged lines** (different rune counts per row).
- The width mask should mirror the shape of the sprite (i.e., provide a width value for each rune in that row).

## Width Values
- Each width mask cell is a single digit:
  - `1` = normal width (one terminal cell)
  - `2` = wide (two terminal cells)

(We can extend this in the future, but only 1/2 are required now.)

## Expanded Width Rule (Critical)
For each row:

- Compute the **expanded cell width** as the sum of the width values for the runes in that row.
- **All rows must expand to the same total cell width.**
- If any row expands to a different width, sprite loading should error.

This allows ragged input while still producing a consistent rectangular sprite size.

## Rendering Semantics
- For width `1`, render the glyph in one cell.
- For width `2`, render the glyph in the first cell and treat the second cell as a **continuation** cell (so nothing else overwrites it).
- Color mask and collision mask should apply to **both cells** for a width‑2 rune.


## Rendering Behavior (Detailed)

### 1) Load-time expansion into cells
Width masks are applied when the sprite is loaded. Each rune is expanded into one or more **terminal cells** based on its width value.

- Width `1` → one cell containing the rune
- Width `2` → two cells:
  - **Cell A**: the rune
  - **Cell B**: a **continuation** placeholder

The expanded width is the sum of widths per row; all rows must match.

### 2) Continuation cells
A continuation cell is a real cell in the sprite grid, but it is **not rendered as a glyph**. It exists to reserve space so other rendering does not overwrite the second half of a wide character.

### 3) Color and collision duplication
For width‑2 runes:
- **Color** is applied to both cells
- **Collision** is applied to both cells

This keeps visual and collision behavior consistent with what the player sees.

### 4) Draw-time behavior
When rendering the sprite:
- Normal cells draw their rune and style.
- Continuation cells are **skipped** (no glyph drawn), but they still **occupy** a cell in the sprite grid.

“Occupy” means other draw calls should not overwrite that cell, preserving wide glyph alignment.

### 5) Layout and positioning
Because width expansion happens at load time:
- `Sprite.W` is the **expanded cell width** (not rune count).
- All alignment and positioning math should use cell width.

### 6) Text rendering is separate
The normal text renderer still assumes 1 cell per rune. If you want wide glyphs in text, a separate width‑aware text renderer is needed.
