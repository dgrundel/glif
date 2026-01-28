package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/scene"
	"github.com/gdamore/tcell/v3"
)

type Duck struct {
	duck  *render.Sprite
	posX  float64
	posY  int
	dir   int
	speed float64
	w     int
	h     int
}

func NewDuck() *Duck {
	duck, err := assets.LoadMaskedSprite(
		"demos/duck/assets/duck.sprite",
		"demos/duck/assets/duck.mask",
		"",
	)
	if err != nil {
		log.Fatal(err)
	}
	return &Duck{
		duck:  duck,
		posX:  1,
		posY:  2,
		dir:   1,
		speed: 8,
	}
}

func (d *Duck) Update(dt float64) {
	maxX := d.w - d.duck.W
	if maxX < 0 {
		maxX = 0
	}

	d.posX += float64(d.dir) * d.speed * dt
	if d.posX <= 0 {
		d.posX = 0
		d.dir = 1
	}
	if d.posX >= float64(maxX) {
		d.posX = float64(maxX)
		d.dir = -1
	}
}

func (d *Duck) Draw(r *render.Renderer) {
	x := int(d.posX + 0.5)
	r.DrawSprite(x, d.posY, d.duck)
}

func (d *Duck) Resize(w, h int) {
	d.w = w
	d.h = h
	_ = d.h
	if d.posY >= h {
		d.posY = h - 1
	}
	if d.posY < 0 {
		d.posY = 0
	}
}

func main() {
	world := scene.New()
	world.Add(NewDuck())
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
			Bg: tcell.ColorBlue,
		},
	}
	eng.Frame.ClearAll()
	if err := eng.Run(world); err != nil {
		log.Fatal(err)
	}
}
