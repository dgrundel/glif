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
	duck, err := assets.LoadMaskedSprite("demos/duck/assets/duck")
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

type Whale struct {
	sprite *render.Sprite
	posX   float64
	posY   int
	dir    int
	speed  float64
	w      int
	h      int
}

func NewWhale(y int) *Whale {
	sprite, err := assets.LoadMaskedSprite("demos/duck/assets/whale")
	if err != nil {
		log.Fatal(err)
	}
	return &Whale{
		sprite: sprite,
		posX:   1,
		posY:   y,
		dir:    -1,
		speed:  4,
	}
}

func (w *Whale) Update(dt float64) {
	maxX := w.w - w.sprite.W
	if maxX < 0 {
		maxX = 0
	}

	w.posX += float64(w.dir) * w.speed * dt
	if w.posX <= 0 {
		w.posX = 0
		w.dir = 1
	}
	if w.posX >= float64(maxX) {
		w.posX = float64(maxX)
		w.dir = -1
	}
}

func (w *Whale) Draw(r *render.Renderer) {
	x := int(w.posX + 0.5)
	r.DrawSprite(x, w.posY, w.sprite)
}

func (w *Whale) Resize(width, height int) {
	w.w = width
	w.h = height
	if w.posY >= height {
		w.posY = height - 1
	}
	if w.posY < 0 {
		w.posY = 0
	}
}

func main() {
	world := scene.New()
	duck := NewDuck()
	whaleY := duck.posY + duck.duck.H + 1
	world.Add(duck)
	world.Add(NewWhale(whaleY))
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
