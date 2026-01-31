package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/ecs"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

type Game struct {
	world   *ecs.World
	player  ecs.Entity
	screenW int
	screenH int
	state   input.State
	binds   input.ActionMap
	bg      grid.Style
	quit    bool
}

func NewGame() *Game {
	pal, err := palette.Load("demos/wasd/assets/default.palette")
	if err != nil {
		log.Fatal(err)
	}
	bg, err := pal.Style('b')
	if err != nil {
		log.Fatal(err)
	}
	world := ecs.NewWorld()

	duck, err := assets.LoadMaskedSprite("demos/wasd/assets/duck")
	if err != nil {
		log.Fatal(err)
	}

	player := world.NewEntity()
	world.AddPosition(player, 0, 0)
	world.AddVelocity(player, 0, 0)
	world.AddSprite(player, duck, 0)

	return &Game{
		world:  world,
		player: player,
		binds: input.ActionMap{
			"move_up":    "w",
			"move_down":  "s",
			"move_left":  "a",
			"move_right": "d",
			"stop":       " ",
			"quit":       "key:esc",
			"quit_alt":   "key:ctrl+c",
		},
		bg: bg,
	}
}

func (g *Game) Update(dt float64) {
	g.applyMovement()
	g.world.Update(dt)
	if g.pressed("quit") || g.pressed("quit_alt") {
		g.quit = true
	}
}

func (g *Game) Draw(r *render.Renderer) {
	g.world.Draw(r)
}

func (g *Game) Resize(w, h int) {
	g.screenW = w
	g.screenH = h
	pos := g.world.Positions[g.player]
	if pos != nil && pos.X == 0 && pos.Y == 0 {
		pos.X = float64((w / 2) - 1)
		pos.Y = float64(h / 2)
	}
}

func (g *Game) setVelocity(dx, dy float64) {
	vel := g.world.Velocities[g.player]
	if vel == nil {
		return
	}
	vel.DX = dx
	vel.DY = dy
}

func (g *Game) SetInput(state input.State) {
	g.state = state
}

func (g *Game) applyMovement() {
	speed := 10.0
	dx := 0.0
	dy := 0.0
	if g.held("move_left") || g.pressed("move_left") {
		dx -= 1
	}
	if g.held("move_right") || g.pressed("move_right") {
		dx += 1
	}
	if g.held("move_up") || g.pressed("move_up") {
		dy -= 1
	}
	if g.held("move_down") || g.pressed("move_down") {
		dy += 1
	}
	if g.pressed("stop") {
		dx = 0
		dy = 0
	}
	if dx != 0 && dy != 0 {
		scale := 0.70710678
		dx *= scale
		dy *= scale
	}
	g.setVelocity(dx*speed, dy*speed)
}

func (g *Game) ShouldQuit() bool {
	return g.quit
}

func (g *Game) held(action input.Action) bool {
	key, ok := g.binds[action]
	if !ok {
		return false
	}
	return g.state.Held[key]
}

func (g *Game) pressed(action input.Action) bool {
	key, ok := g.binds[action]
	if !ok {
		return false
	}
	return g.state.Pressed[key]
}

func main() {
	game := NewGame()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: game.bg}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
