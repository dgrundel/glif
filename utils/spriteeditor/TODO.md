# Sprite Editor

## Overview

Make a sprite editor tool used to create and edit sprite files and their corresponding masks and palettes. The editor is designed to operate on a single sprite at once. It is run via command line, e.g.:

```
go run ./utils/spriteeditor ./path/to/spritefile.sprite
```

Based on the sprite path, it also loads any width, color, palette, collision, or other associate files as needed for the modes described below.

## Editor Modes

### Sprite Mode

A free-form text editor for creating and editing .sprite files.

### Width Mode

An editor for the sprite's width mask. If no width mask file is present, generate a width mask consisting of a `1` for every character in the sprite file.

The editor should not save width masks to disk if they consist of only `1`s, as this is the default and creating the file is unnecessary.

### Color Mode

An editor for the sprite's color mask. If no color mask file is present, generate one.

### Palette Mode

An editor for either the sprite-specific palette or the default palette, whichever exists. if neither exists, create a sprite-specific palette file.

### Collision Mode

An editor for the sprite's collision mask.

The editor should not save collision masks to disk if they do not have any collision areas, as this is the default behavior and creating the file is unnecessary.

## Additional features

### Resize

This tool will allow the user to resize a sprite to a desired width and height. The user may specify and anchor point (top left, top right, bottom left, bottom right, center, etc.). If the new dimensions are smaller than the existing dimensions, the sprite is clipped. Else the sprite is padded with spaces.

## UI, Controls, Layout

The UI has a "top bar" with an alternate background color to distinguish it. There is also a "bottom bar" which also has an alternate background color. The current mode is displayed in the top bar as `Mode (Ctrl+M): [mode name]`. The user uses Ctrl+M to cycle between modes. The user uses Ctrl+S to save. The bottom bar shows contextual information such as error messages or sprite dimensions.

Note: The shortcut keys above may conflict with MacOS or Linux shortcut keys. Please suggest alternatives.

## Implementation Notes

We should consider modifying the existing sprite loading code so that we can use the same code in both the engine and the editor. Of course, the editor should allow for invalid sprite files to be opened so that they can be fixed in the editor.