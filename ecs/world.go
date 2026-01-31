package ecs

import (
	"sort"

	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/tilemap"
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

type TileMapRef struct {
	Map *tilemap.Map
	Z   int
}

type UpdateSystem func(w *World, dt float64)

type World struct {
	next       Entity
	Positions  map[Entity]*Position
	Velocities map[Entity]*Velocity
	Sprites    map[Entity]*SpriteRef
	TileMaps   map[Entity]*TileMapRef
	Camera     *camera.Camera

	UpdateSystems []UpdateSystem

	OnResize func(w, h int)
}

func NewWorld() *World {
	return &World{
		Positions:     make(map[Entity]*Position),
		Velocities:    make(map[Entity]*Velocity),
		Sprites:       make(map[Entity]*SpriteRef),
		TileMaps:      make(map[Entity]*TileMapRef),
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

func (w *World) AddTileMap(e Entity, m *tilemap.Map, z int) {
	w.TileMaps[e] = &TileMapRef{Map: m, Z: z}
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
		tile   *TileMapRef
		z      int
	}
	items := make([]drawItem, 0, len(w.Sprites)+len(w.TileMaps))
	for e, spr := range w.Sprites {
		pos, ok := w.Positions[e]
		if !ok || spr.Sprite == nil {
			continue
		}
		items = append(items, drawItem{entity: e, pos: pos, sprite: spr, z: spr.Z})
	}
	for e, tm := range w.TileMaps {
		pos, ok := w.Positions[e]
		if !ok || tm.Map == nil {
			continue
		}
		items = append(items, drawItem{entity: e, pos: pos, tile: tm, z: tm.Z})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].z == items[j].z {
			return items[i].entity < items[j].entity
		}
		return items[i].z < items[j].z
	})
	for _, item := range items {
		if item.sprite != nil {
			x := item.pos.X
			y := item.pos.Y
			if w.Camera != nil {
				if !w.Camera.InView(x, y, item.sprite.Sprite.W, item.sprite.Sprite.H) {
					continue
				}
				x, y = w.Camera.WorldToScreen(x, y)
			}
			sx := int(x + 0.5)
			sy := int(y + 0.5)
			r.DrawSprite(sx, sy, item.sprite.Sprite)
			continue
		}
		if item.tile != nil {
			x := item.pos.X
			y := item.pos.Y
			item.tile.Map.Draw(r, x, y, w.Camera)
		}
	}
}

func (w *World) Resize(width, height int) {
	if w.OnResize != nil {
		w.OnResize(width, height)
	}
}
