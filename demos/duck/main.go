package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/ecs"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func main() {
	world := ecs.NewWorld()

	duck, err := assets.LoadMaskedSprite("demos/duck/assets/duck")
	if err != nil {
		log.Fatal(err)
	}
	whale, err := assets.LoadMaskedSprite("demos/duck/assets/whale")
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

	world.OnEvent = func(ev tcell.Event) bool {
		if key, ok := ev.(*tcell.EventKey); ok {
			return key.Key() == tcell.KeyEscape || key.Key() == tcell.KeyCtrlC
		}
		return false
	}

	eng, err := engine.New(world, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{
		Ch: ' ',
		Style: grid.Style{
			Fg: tcell.ColorReset,
			Bg: color.Blue,
		},
	}
	eng.Frame.ClearAll()
	if err := eng.Run(world); err != nil {
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
