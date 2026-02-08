Make a sprite editor. 

The sprite editor is a basic text editor for manipulating (or creating) sprite files. 

It has a "top bar" which spans the first row of the screen. The top bar has a white background with black text. It displays the current sprite file path. The path is truncated if it is too long to fit.

It also has a "status bar" at the bottom of the screen. The status bar is also white background/black text. The status bar displays the dimensions of the current sprite. Any kind of error should be displayed in the status bar as white text on a red background.

The "area" of the sprite should have a different background color (perhaps a dark gray instead of the default black background). As the user types or adds lines to the sprite, the sprite area highlight should adapt.

## Controls 
- Ctrl+S saves the file. A note in the status bar tells the user the file was saved.
- Ctrl+Q (twice) or Esc (twice) exits without saving. The first time the user presses Ctrl+Q or Esc, a warning is displayed in the status bar and the status bar turns red with white text.
- Arrow keys can be used to move the cursur anywhere, even beyond the bounds of the sprite.
- Ctrl+K trims off space from all four sides of the sprite