# Ski Demo

Build a simple 2D downhill skiing game. The player controls a skier moving downhill on a continuous slope. The skier constantly moves forward, and the player can steer left and right. The course is defined by pairs of gates (like a slalom course) that the player must pass through in order. Missing a gate counts as a failure or penalty. Gates are generated continuously as the skier moves downhill, with increasing difficulty over time (offset gates and faster speed). The goal is to complete as many gates as possible without missing any.

In addition to gates, there are obstacles in the form of trees. Hitting a tree ends the game. There are also clumps of rough snow. When the player hits the rough snow, it reduces the player's speed (but never below zero). The player's score should be displayed at the top right of the screen. The score is simply the number of gates through which the player has passed.

## Sprites

There are several sprites in the assets folder for you to use. ski_down, ski_left, and ski_right are for the player character. Use ski_down when the character is moving straight down. Use ski_left and ski_right when the player is pressing the left or right arrow keys to move left or right. flag_left and flag_right should be used for the gates. Use flag_left for the left side of the gate and flag_right for the right side. trees and snow_rough sprites are also provided for the obstacles and speed reduction as previously described.

## Implementation Notes

The game is shown from a somewhat top-down perspective, so on screen the player character appears to be moving downward. The camera follows the character.

I think it's probably simplest to implement the player character in a fixed position on screen and to simply move the obstacles, gates, etc upward from the bottom of the screen. 

### Rules for gates, obstacles

- The vertical distance between gates must be at least 4 vertical spaces (rows). At game start, the distance should be larger, perhaps 12 rows. Over time, the distance should reduce to 4.
- The horizontal distance between gates should be attainable for the player. Probably no worse than a 45 degree angle between gates.
- Rough snow and trees should appear randomly. The rate (likelihood) at which each appears should be adjustable in code.