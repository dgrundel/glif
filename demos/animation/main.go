package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

// AnimationDemo shows a single looping animation and exits on Esc/Ctrl+C.
// It keeps the current frame in player and advances it via playerAn.Update.
type AnimationDemo struct {
	state    input.State             // current input snapshot from the engine
	binds    input.ActionMap         // action -> key bindings
	quit     bool                    // set true to exit the demo
	player   *render.Sprite          // current frame to draw
	playerAn *render.AnimationPlayer // animation playback state
	text     grid.Style              // UI text style (palette-driven)
	bg       grid.Style              // background style (palette-driven)
}

// NewAnimationDemo loads assets, creates the animation player, and seeds styles.
func NewAnimationDemo() *AnimationDemo {
	// Palette for UI and background styling.
	pal := palette.MustLoad("demos/animation/assets/default.palette")

	// Base sprite provides size + color + optional collision data.
	sprite, err := assets.LoadSprite("demos/animation/assets/explosion")
	if err != nil {
		log.Fatal(err)
	}

	// Load the named animation from explosion.explode.animation (and optional .animation.color).
	anim, err := sprite.LoadAnimation("explode")
	if err != nil {
		log.Fatal(err)
	}

	return &AnimationDemo{
		// Key bindings: quit only.
		binds: input.ActionMap{
			"quit":     "key:esc",
			"quit_alt": "key:ctrl+c",
		},
		player:   sprite,        // initial frame
		playerAn: anim.Play(12), // 12 fps playback
		text:     pal.MustStyle('w'),
		bg:       pal.MustStyle('y'),
	}
}

// Update advances animation time and handles exit input.
func (d *AnimationDemo) Update(dt float64) {
	// Exit on Esc/Ctrl+C.
	if d.pressed("quit") || d.pressed("quit_alt") {
		d.quit = true
		return
	}
	// Advance the animation and grab the current frame.
	if d.playerAn != nil {
		d.playerAn.Update(dt)
		d.player = d.playerAn.Sprite()
	}
}

// Draw centers the current frame and draws a small label.
func (d *AnimationDemo) Draw(r *render.Renderer) {
	// Center sprite on screen.
	x := max(0, (r.Frame.W-d.player.W)/2)
	y := max(0, (r.Frame.H-d.player.H)/2)
	r.DrawSprite(x, y, d.player)

	// Simple on-screen hint.
	r.DrawText(2, 1, "Explosion animation (press Esc to quit)", d.text)
}

// Resize is required by the engine; this demo doesn't need it.
func (d *AnimationDemo) Resize(w, h int) {
	_ = w
	_ = h
}

// SetInput receives the per-frame input state from the engine.
func (d *AnimationDemo) SetInput(state input.State) {
	d.state = state
}

// ShouldQuit tells the engine when to exit.
func (d *AnimationDemo) ShouldQuit() bool {
	return d.quit
}

// pressed checks for a one-frame key press.
func (d *AnimationDemo) pressed(action input.Action) bool {
	key, ok := d.binds[action]
	if !ok {
		return false
	}
	return d.state.Pressed[key]
}

// max is a tiny helper for centering.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// main wires the demo into the engine and starts the loop.
func main() {
	demo := NewAnimationDemo()
	eng, err := engine.New(demo, 0)
	if err != nil {
		log.Fatal(err)
	}
	// Clear the frame with the demo background style.
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: demo.bg}
	eng.Frame.ClearAll()
	if err := eng.Run(demo); err != nil {
		log.Fatal(err)
	}
}
