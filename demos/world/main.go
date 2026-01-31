package main

import (
	"log"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/ecs"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/tilemap"
)

type Demo struct {
	world  *ecs.World
	cam    *camera.Camera
	tile   *tilemap.Map
	state  input.State
	binds  input.ActionMap
	player ecs.Entity
	quit   bool
	bg     grid.Style
}

func NewDemo() *Demo {
	pal, err := palette.Load("demos/world/assets/ui.palette")
	if err != nil {
		log.Fatal(err)
	}
	bg, err := pal.Style('b')
	if err != nil {
		log.Fatal(err)
	}
	world := ecs.NewWorld()
	cam := &camera.Camera{}
	world.Camera = cam

	duck, err := assets.LoadMaskedSprite("demos/world/assets/duck")
	if err != nil {
		log.Fatal(err)
	}

	player := world.NewEntity()
	world.AddPosition(player, 0, 0)
	world.AddVelocity(player, 0, 0)
	world.AddSprite(player, duck, 0)

	tm, err := tilemap.LoadFromFiles(
		"demos/world/assets/world.map",
		"demos/world/assets/world.tiles",
	)
	if err != nil {
		log.Fatal(err)
	}
	tmEntity := world.NewEntity()
	world.AddPosition(tmEntity, 0, 0)
	world.AddTileMap(tmEntity, tm, -1)

	world.OnResize = func(w, h int) {
		cam.ViewW = w
		cam.ViewH = h
		pos := world.Positions[player]
		if pos != nil && pos.X == 0 && pos.Y == 0 {
			pos.X = float64((w / 2) - 1)
			pos.Y = float64(h / 2)
		}
	}

	return &Demo{
		world:  world,
		cam:    cam,
		tile:   tm,
		player: player,
		binds: input.ActionMap{
			"move_up":    "w",
			"move_down":  "s",
			"move_left":  "a",
			"move_right": "d",
			"stop":       " ",
			"pan_up":     "key:up",
			"pan_down":   "key:down",
			"pan_left":   "key:left",
			"pan_right":  "key:right",
			"quit":       "key:esc",
			"quit_alt":   "key:ctrl+c",
		},
		bg: bg,
	}
}

func (d *Demo) Update(dt float64) {
	d.applyMovement()
	d.updateCamera(dt)
	d.world.Update(dt)
	if d.pressed("quit") || d.pressed("quit_alt") {
		d.quit = true
	}
}

func (d *Demo) updateCamera(dt float64) {
	speed := 12.0
	if d.held("pan_left") {
		d.cam.X -= speed * dt
	}
	if d.held("pan_right") {
		d.cam.X += speed * dt
	}
	if d.held("pan_up") {
		d.cam.Y -= speed * dt
	}
	if d.held("pan_down") {
		d.cam.Y += speed * dt
	}
	d.clampCamera()
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

func (d *Demo) held(action input.Action) bool {
	key, ok := d.binds[action]
	if !ok {
		return false
	}
	return d.state.Held[key]
}

func (d *Demo) pressed(action input.Action) bool {
	key, ok := d.binds[action]
	if !ok {
		return false
	}
	return d.state.Pressed[key]
}

func (d *Demo) applyMovement() {
	speed := 10.0
	dx := 0.0
	dy := 0.0
	if d.held("move_left") || d.pressed("move_left") {
		dx -= 1
	}
	if d.held("move_right") || d.pressed("move_right") {
		dx += 1
	}
	if d.held("move_up") || d.pressed("move_up") {
		dy -= 1
	}
	if d.held("move_down") || d.pressed("move_down") {
		dy += 1
	}
	if d.pressed("stop") {
		dx = 0
		dy = 0
	}
	if dx != 0 && dy != 0 {
		scale := 0.70710678
		dx *= scale
		dy *= scale
	}
	vel := d.world.Velocities[d.player]
	if vel != nil {
		vel.DX = dx * speed
		vel.DY = dy * speed
	}
}

func (d *Demo) clampCamera() {
	if d.tile == nil || d.cam.ViewW <= 0 || d.cam.ViewH <= 0 {
		return
	}
	maxX := float64(d.tile.W*d.tile.TileW - d.cam.ViewW)
	maxY := float64(d.tile.H*d.tile.TileH - d.cam.ViewH)
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	if d.cam.X < 0 {
		d.cam.X = 0
	}
	if d.cam.Y < 0 {
		d.cam.Y = 0
	}
	if d.cam.X > maxX {
		d.cam.X = maxX
	}
	if d.cam.Y > maxY {
		d.cam.Y = maxY
	}
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
