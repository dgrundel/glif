package ecs

import (
	"sort"

	"github.com/dgrundel/glif/render"
)

type Entity int

type Position struct {
	X float64
	Y float64
}

type Velocity struct {
	DX float64
	DY float64
}

type SpriteRef struct {
	Sprite *render.Sprite
	Z      int
}

type UpdateSystem func(w *World, dt float64)

type World struct {
	next       Entity
	Positions  map[Entity]*Position
	Velocities map[Entity]*Velocity
	Sprites    map[Entity]*SpriteRef

	UpdateSystems []UpdateSystem

	OnResize func(w, h int)
}

func NewWorld() *World {
	return &World{
		Positions:     make(map[Entity]*Position),
		Velocities:    make(map[Entity]*Velocity),
		Sprites:       make(map[Entity]*SpriteRef),
		UpdateSystems: []UpdateSystem{},
	}
}

func (w *World) NewEntity() Entity {
	e := w.next
	w.next++
	return e
}

func (w *World) AddPosition(e Entity, x, y float64) {
	w.Positions[e] = &Position{X: x, Y: y}
}

func (w *World) AddVelocity(e Entity, dx, dy float64) {
	w.Velocities[e] = &Velocity{DX: dx, DY: dy}
}

func (w *World) AddSprite(e Entity, sprite *render.Sprite, z int) {
	w.Sprites[e] = &SpriteRef{Sprite: sprite, Z: z}
}

func (w *World) AddSystem(sys UpdateSystem) {
	w.UpdateSystems = append(w.UpdateSystems, sys)
}

func (w *World) Update(dt float64) {
	// Movement system (built-in).
	for e, pos := range w.Positions {
		vel, ok := w.Velocities[e]
		if !ok {
			continue
		}
		pos.X += vel.DX * dt
		pos.Y += vel.DY * dt
	}

	for _, sys := range w.UpdateSystems {
		sys(w, dt)
	}
}

func (w *World) Draw(r *render.Renderer) {
	type drawItem struct {
		entity Entity
		pos    *Position
		sprite *SpriteRef
	}
	items := make([]drawItem, 0, len(w.Sprites))
	for e, spr := range w.Sprites {
		pos, ok := w.Positions[e]
		if !ok || spr.Sprite == nil {
			continue
		}
		items = append(items, drawItem{entity: e, pos: pos, sprite: spr})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].sprite.Z == items[j].sprite.Z {
			return items[i].entity < items[j].entity
		}
		return items[i].sprite.Z < items[j].sprite.Z
	})
	for _, item := range items {
		x := int(item.pos.X + 0.5)
		y := int(item.pos.Y + 0.5)
		r.DrawSprite(x, y, item.sprite.Sprite)
	}
}

func (w *World) Resize(width, height int) {
	if w.OnResize != nil {
		w.OnResize(width, height)
	}
}
