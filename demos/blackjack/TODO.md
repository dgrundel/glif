This demo will be a simple blackjack game.

When the demo is first opened, the "splash" sprite is displayed on screen with a prompt to "press enter to continue". When enter is pressed, the splash disappears and the game begins.

During the game, graphical cards are displayed for the dealer and player. The current total for the dealer and player's hands are displayed as a number. The player chooses actions from a menu that can be navigated with up/down arrow keys and enter to select. This interface is shown in EXAMPLE.txt.

The rules for this blackjack game are simplified:
- Numeric cards (2 through 10) are worth face value. J, Q, and K are worth 10 points. A is worth 1 or 10 points, whichever brings the player closer to 21 without exceeding it.
- At the start of the game the dealer draws two cards for themself, and reveals on of the two. 
- Then two cards are drawn for the player and revealed.
- The player may choose to HIT, STAY, or QUIT
- If the player chooses to HIT, they draw an additional card.
  - If the additional card brings the player's total to 21, they win.
  - If the additional card brings the player's total > 21, they lose.
  - If the additional card brings the player's total < 21, they may choose again to HIT, STAY, or QUIT.
- If the player chooses to STAY, it is the dealer's turn.
- If the player chooses to QUIT, the game ends and the demo exits immediately.
- When it is the dealer's turn, the dealer must do the following:
  - If the dealer's total is < 17, they must HIT.
  - If the dealer's total is >= 18, they must STAY.
- When the dealer STAYs, the game is over and the player with the highest total wins.
- When the game is over, the player is prompted to play again or quit.