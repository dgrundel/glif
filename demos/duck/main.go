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

type Demo struct {
	world *ecs.World
	state input.State
	binds input.ActionMap
	quit  bool
	bg    grid.Style
}

func NewDemo() *Demo {
	pal, err := palette.Load("demos/duck/assets/default.palette")
	if err != nil {
		log.Fatal(err)
	}
	bg, err := pal.Style('b')
	if err != nil {
		log.Fatal(err)
	}
	world := ecs.NewWorld()

	duck, err := assets.LoadSprite("demos/duck/assets/duck")
	if err != nil {
		log.Fatal(err)
	}
	whale, err := assets.LoadSprite("demos/duck/assets/whale")
	if err != nil {
		log.Fatal(err)
	}

	duckEntity := world.NewEntity()
	world.AddPosition(duckEntity, 1, 2)
	world.AddVelocity(duckEntity, 8, 0)
	world.AddSprite(duckEntity, duck, 1)

	whaleEntity := world.NewEntity()
	whaleY := float64(2 + duck.H + 1)
	world.AddPosition(whaleEntity, 1, whaleY)
	world.AddVelocity(whaleEntity, 4, 0)
	world.AddSprite(whaleEntity, whale, 0)

	screenW := 0
	world.OnResize = func(w, h int) {
		screenW = w
	}

	world.AddSystem(func(w *ecs.World, dt float64) {
		_ = dt
		if screenW <= 0 {
			return
		}
		applyBounce(w, screenW, duckEntity)
		applyBounce(w, screenW, whaleEntity)
	})

	return &Demo{
		world: world,
		binds: input.ActionMap{
			"quit":     "key:esc",
			"quit_alt": "key:ctrl+c",
		},
		bg: bg,
	}
}

func (d *Demo) Update(dt float64) {
	d.world.Update(dt)
	if d.pressed("quit") || d.pressed("quit_alt") {
		d.quit = true
	}
}

func (d *Demo) Draw(r *render.Renderer) {
	d.world.Draw(r)
}

func (d *Demo) Resize(w, h int) {
	d.world.Resize(w, h)
}

func (d *Demo) SetInput(state input.State) {
	d.state = state
}

func (d *Demo) ShouldQuit() bool {
	return d.quit
}

func (d *Demo) pressed(action input.Action) bool {
	key, ok := d.binds[action]
	if !ok {
		return false
	}
	return d.state.Pressed[key]
}

func main() {
	demo := NewDemo()
	eng, err := engine.New(demo, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: demo.bg}
	eng.Frame.ClearAll()
	if err := eng.Run(demo); err != nil {
		log.Fatal(err)
	}
}

func applyBounce(w *ecs.World, screenW int, e ecs.Entity) {
	pos, ok := w.Positions[e]
	if !ok {
		return
	}
	vel, ok := w.Velocities[e]
	if !ok {
		return
	}
	spriteRef, ok := w.Sprites[e]
	if !ok || spriteRef.Sprite == nil {
		return
	}
	maxX := float64(screenW - spriteRef.Sprite.W)
	if maxX < 0 {
		maxX = 0
	}
	if pos.X <= 0 {
		pos.X = 0
		vel.DX = abs(vel.DX)
	}
	if pos.X >= maxX {
		pos.X = maxX
		vel.DX = -abs(vel.DX)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
