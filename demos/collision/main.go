package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/collision"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

type CollisionDemo struct {
	binds     input.ActionMap
	quit      bool
	playerX   float64
	playerY   float64
	player    *render.Sprite
	block     *render.Sprite
	textStyle grid.Style
	okStyle   grid.Style
	noStyle   grid.Style
	bgStyle   grid.Style
	actions   input.ActionState
}

func NewCollisionDemo() *CollisionDemo {
	pal := palette.MustLoad("demos/collision/assets/default.palette")
	return &CollisionDemo{
		binds: input.ActionMap{
			"up":        "key:up",
			"down":      "key:down",
			"left":      "key:left",
			"right":     "key:right",
			"up_alt":    "w",
			"down_alt":  "s",
			"left_alt":  "a",
			"right_alt": "d",
			"quit":      "key:esc",
			"quit_alt":  "key:ctrl+c",
		},
		player:    mustSprite("demos/collision/assets/player"),
		block:     mustSprite("demos/collision/assets/block"),
		textStyle: pal.MustStyle('x'),
		okStyle:   pal.MustStyle('g'),
		noStyle:   pal.MustStyle('r'),
		bgStyle:   pal.MustStyle('b'),
		playerX:   4,
		playerY:   6,
	}
}

func (d *CollisionDemo) Update(dt float64) {
	if d.actions.Pressed["quit"] || d.actions.Pressed["quit_alt"] {
		d.quit = true
		return
	}
	const moveSpeed = 12.0
	if d.actions.Held["left"] || d.actions.Held["left_alt"] {
		d.playerX -= moveSpeed * dt
	}
	if d.actions.Held["right"] || d.actions.Held["right_alt"] {
		d.playerX += moveSpeed * dt
	}
	if d.actions.Held["up"] || d.actions.Held["up_alt"] {
		d.playerY -= moveSpeed * dt
	}
	if d.actions.Held["down"] || d.actions.Held["down_alt"] {
		d.playerY += moveSpeed * dt
	}
}

func (d *CollisionDemo) Draw(r *render.Renderer) {
	blockX := max(0, (r.Frame.W-d.block.W)/2)
	blockY := max(0, (r.Frame.H-d.block.H)/2)
	r.DrawSprite(blockX, blockY, d.block)
	playerX := int(d.playerX + 0.5)
	playerY := int(d.playerY + 0.5)
	r.DrawSprite(playerX, playerY, d.player)

	hit := collision.Overlaps(playerX, playerY, d.player, blockX, blockY, d.block)
	r.DrawText(1, 1, "Collision:", d.textStyle)
	if hit {
		r.DrawText(12, 1, "YES", d.okStyle)
	} else {
		r.DrawText(12, 1, "NO", d.noStyle)
	}
}

func (d *CollisionDemo) Resize(w, h int) {
	_ = w
	_ = h
}

func (d *CollisionDemo) ActionMap() input.ActionMap {
	return d.binds
}

func (d *CollisionDemo) UpdateActionState(state input.ActionState) {
	d.actions = state
}

func (d *CollisionDemo) ShouldQuit() bool {
	return d.quit
}

func mustSprite(base string) *render.Sprite {
	sprite, err := assets.LoadSprite(base)
	if err != nil {
		log.Fatal(err)
	}
	return sprite
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	demo := NewCollisionDemo()
	eng, err := engine.New(demo, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: demo.bgStyle}
	eng.Frame.ClearAll()
	if err := eng.Run(demo); err != nil {
		log.Fatal(err)
	}
}
