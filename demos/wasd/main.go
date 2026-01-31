package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/ecs"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type Game struct {
	world   *ecs.World
	player  ecs.Entity
	screenW int
	screenH int
	input   *input.Manager
}

func NewGame() *Game {
	world := ecs.NewWorld()

	duck, err := assets.LoadMaskedSprite("demos/duck/assets/duck")
	if err != nil {
		log.Fatal(err)
	}

	player := world.NewEntity()
	world.AddPosition(player, 0, 0)
	world.AddVelocity(player, 0, 0)
	world.AddSprite(player, duck, 0)

	return &Game{world: world, player: player, input: input.New(0.12)}
}

func (g *Game) Update(dt float64) {
	state := g.input.Step(dt)
	g.applyMovement(state)
	g.world.Update(dt)
}

func (g *Game) Draw(r *render.Renderer) {
	g.world.Draw(r)
}

func (g *Game) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		keyID := input.KeyID(ev)
		if keyID == "key:esc" || keyID == "key:ctrl+c" {
			return true
		}
		g.input.HandleEvent(ev)
	}
	return false
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

func (g *Game) applyMovement(state input.State) {
	speed := 10.0
	dx := 0.0
	dy := 0.0
	if state.Held["a"] || state.Pressed["a"] {
		dx -= 1
	}
	if state.Held["d"] || state.Pressed["d"] {
		dx += 1
	}
	if state.Held["w"] || state.Pressed["w"] {
		dy -= 1
	}
	if state.Held["s"] || state.Pressed["s"] {
		dy += 1
	}
	if state.Pressed[" "] {
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

func main() {
	game := NewGame()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{
		Ch: ' ',
		Style: grid.Style{
			Fg: tcell.ColorReset,
			Bg: tcell.ColorBlue,
		},
	}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
