package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/collision"
	"github.com/dgrundel/glif/ecs"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

const (
	shipSpeed    = 30.0
	enemySpeed   = 6.0
	bulletSpeed  = 45.0
	fireCooldown = 0.35
	explodeFPS   = 12.0
	flyInSpeed   = 28.0
	enemy2Chance = 0.1
	enemyRows    = 2
	enemyGapX    = 2
	enemyGapY    = 2
	enemyStartY  = 2
)

type Game struct {
	world   *ecs.World
	ship    ecs.Entity
	enemies []ecs.Entity
	bullets []ecs.Entity

	shipSprite    *render.Sprite
	enemySprite   *render.Sprite
	enemy2Sprite  *render.Sprite
	bulletSprite  *render.Sprite
	enemyDestroy  *render.Animation
	enemy2Destroy *render.Animation

	screenW   int
	screenH   int
	enemyMaxW int
	enemyMaxH int

	enemyDir      float64
	fireTimer     float64
	shipPlaced    bool
	enemiesPlaced bool
	level         int

	binds      input.ActionMap
	actions    input.ActionState
	bg         grid.Style
	levelStyle grid.Style
	quit       bool

	explosions   []explosion
	enemyAnims   map[ecs.Entity]*render.Animation
	enemyTargets map[ecs.Entity]float64
	rng          *rand.Rand
}

type explosion struct {
	entity  ecs.Entity
	frames  []*render.Sprite
	elapsed float64
}

func NewGame() *Game {
	pal, err := palette.Load("demos/invaders/assets/default.palette")
	if err != nil {
		log.Fatal(err)
	}
	bg, err := pal.Style('k')
	if err != nil {
		log.Fatal(err)
	}
	levelStyle, err := pal.Style('s')
	if err != nil {
		log.Fatal(err)
	}

	world := ecs.NewWorld()
	shipSprite := assets.MustLoadSprite("demos/invaders/assets/ship")
	enemySprite := assets.MustLoadSprite("demos/invaders/assets/enemy")
	enemy2Sprite := assets.MustLoadSprite("demos/invaders/assets/enemy2")
	bulletSprite := assets.MustLoadSprite("demos/invaders/assets/bullet")
	enemyDestroy, err := enemySprite.LoadAnimation("destroy")
	if err != nil {
		log.Printf("load destroy animation: %v", err)
	}
	enemy2Destroy, err := enemy2Sprite.LoadAnimation("destroy")
	if err != nil {
		log.Printf("load enemy2 destroy animation: %v", err)
	}
	enemyMaxW := enemySprite.W
	enemyMaxH := enemySprite.H
	if enemy2Sprite.W > enemyMaxW {
		enemyMaxW = enemy2Sprite.W
	}
	if enemy2Sprite.H > enemyMaxH {
		enemyMaxH = enemy2Sprite.H
	}

	ship := world.NewEntity()
	world.AddPosition(ship, 0, 0)
	world.AddVelocity(ship, 0, 0)
	world.AddSprite(ship, shipSprite, 1)

	return &Game{
		world:         world,
		ship:          ship,
		shipSprite:    shipSprite,
		enemySprite:   enemySprite,
		enemy2Sprite:  enemy2Sprite,
		bulletSprite:  bulletSprite,
		enemyDestroy:  enemyDestroy,
		enemy2Destroy: enemy2Destroy,
		enemyMaxW:     enemyMaxW,
		enemyMaxH:     enemyMaxH,
		enemyDir:      1,
		level:         1,
		binds: input.ActionMap{
			"move_left":  "key:left",
			"move_right": "key:right",
			"shoot":      " ",
			"quit":       "key:esc",
			"quit_alt":   "key:ctrl+c",
		},
		bg:           bg,
		levelStyle:   levelStyle,
		enemyAnims:   make(map[ecs.Entity]*render.Animation),
		enemyTargets: make(map[ecs.Entity]float64),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Game) Update(dt float64) {
	if g.pressed("quit") || g.pressed("quit_alt") {
		g.quit = true
		return
	}

	if g.fireTimer > 0 {
		g.fireTimer -= dt
		if g.fireTimer < 0 {
			g.fireTimer = 0
		}
	}

	g.applyShipMovement()
	g.updateEnemyVelocities()

	if g.pressed("shoot") && g.fireTimer == 0 {
		g.spawnBullet()
		g.fireTimer = fireCooldown
	}

	g.world.Update(dt)
	g.clampShip()
	g.resolveHits()
	g.updateExplosions(dt)
	g.cleanupBullets()
	g.checkNextLevel()
}

func (g *Game) Draw(r *render.Renderer) {
	g.world.Draw(r)
	r.DrawText(0, 0, fmt.Sprintf("Level: %d", g.level), g.levelStyle)
}

func (g *Game) Resize(w, h int) {
	g.screenW = w
	g.screenH = h

	if !g.shipPlaced && g.shipSprite != nil {
		pos := g.world.Positions[g.ship]
		if pos != nil {
			pos.X = float64((w - g.shipSprite.W) / 2)
			pos.Y = float64(h - g.shipSprite.H - 1)
			g.shipPlaced = true
		}
	}

	if !g.enemiesPlaced {
		g.layoutEnemies()
	}
}

func (g *Game) ShouldQuit() bool {
	return g.quit
}

func (g *Game) ClearStyle() grid.Style {
	return g.bg
}

func (g *Game) ActionMap() input.ActionMap {
	return g.binds
}

func (g *Game) UpdateActionState(state input.ActionState) {
	g.actions = state
}

func (g *Game) applyShipMovement() {
	vel := g.world.Velocities[g.ship]
	if vel == nil {
		return
	}
	dx := 0.0
	if g.held("move_left") || g.pressed("move_left") {
		dx -= 1
	}
	if g.held("move_right") || g.pressed("move_right") {
		dx += 1
	}
	vel.DX = dx * shipSpeed
	vel.DY = 0
}

func (g *Game) updateEnemyVelocities() {
	if g.screenW <= 0 || len(g.enemies) == 0 {
		return
	}
	if g.updateEnemyFlyIn() {
		return
	}
	minX := math.MaxFloat64
	maxX := -math.MaxFloat64
	for _, e := range g.enemies {
		pos := g.world.Positions[e]
		ref := g.world.Sprites[e]
		if pos == nil || ref == nil || ref.Sprite == nil {
			continue
		}
		left := pos.X
		right := pos.X + float64(ref.Sprite.W)
		if left < minX {
			minX = left
		}
		if right > maxX {
			maxX = right
		}
	}
	if maxX >= float64(g.screenW-1) {
		g.enemyDir = -1
	}
	if minX <= 0 {
		g.enemyDir = 1
	}

	for _, e := range g.enemies {
		vel := g.world.Velocities[e]
		if vel == nil {
			continue
		}
		vel.DX = g.enemyDir * enemySpeed
		vel.DY = 0
	}
}

func (g *Game) updateEnemyFlyIn() bool {
	if len(g.enemyTargets) == 0 {
		return false
	}
	flying := false
	for _, e := range g.enemies {
		targetX, ok := g.enemyTargets[e]
		if !ok {
			continue
		}
		pos := g.world.Positions[e]
		vel := g.world.Velocities[e]
		if pos == nil || vel == nil {
			continue
		}
		delta := targetX - pos.X
		if math.Abs(delta) <= 0.5 {
			pos.X = targetX
			vel.DX = 0
			delete(g.enemyTargets, e)
			continue
		}
		flying = true
		if delta < 0 {
			vel.DX = -flyInSpeed
		} else {
			vel.DX = flyInSpeed
		}
		vel.DY = 0
	}
	if flying {
		for _, e := range g.enemies {
			if _, ok := g.enemyTargets[e]; ok {
				continue
			}
			vel := g.world.Velocities[e]
			if vel != nil {
				vel.DX = 0
				vel.DY = 0
			}
		}
	}
	return flying
}

func (g *Game) spawnBullet() {
	shipPos := g.world.Positions[g.ship]
	if shipPos == nil {
		return
	}
	bullet := g.world.NewEntity()
	bx := shipPos.X + float64(g.shipSprite.W/2)
	by := shipPos.Y - 1
	g.world.AddPosition(bullet, bx, by)
	g.world.AddVelocity(bullet, 0, -bulletSpeed)
	g.world.AddSprite(bullet, g.bulletSprite, 2)
	g.bullets = append(g.bullets, bullet)
}

func (g *Game) cleanupBullets() {
	if len(g.bullets) == 0 {
		return
	}
	remaining := g.bullets[:0]
	for _, b := range g.bullets {
		pos := g.world.Positions[b]
		if pos == nil {
			continue
		}
		if pos.Y < -1 {
			g.removeEntity(b)
			continue
		}
		remaining = append(remaining, b)
	}
	g.bullets = remaining
}

func (g *Game) resolveHits() {
	if len(g.bullets) == 0 || len(g.enemies) == 0 {
		return
	}
	remainingBullets := g.bullets[:0]
	remainingEnemies := g.enemies[:0]

	enemyHit := make(map[ecs.Entity]bool, len(g.enemies))
	for _, b := range g.bullets {
		bpos := g.world.Positions[b]
		if bpos == nil {
			continue
		}
		bulletRemoved := false
		for _, e := range g.enemies {
			if enemyHit[e] {
				continue
			}
			epos := g.world.Positions[e]
			eref := g.world.Sprites[e]
			if epos == nil || eref == nil || eref.Sprite == nil {
				continue
			}
			if collision.Overlaps(int(math.Floor(bpos.X)), int(math.Floor(bpos.Y)), g.bulletSprite, int(math.Floor(epos.X)), int(math.Floor(epos.Y)), eref.Sprite) {
				enemyHit[e] = true
				bulletRemoved = true
				break
			}
		}
		if !bulletRemoved {
			remainingBullets = append(remainingBullets, b)
		}
	}

	for _, e := range g.enemies {
		if enemyHit[e] {
			g.onEnemyHit(e)
			continue
		}
		remainingEnemies = append(remainingEnemies, e)
	}
	g.enemies = remainingEnemies

	for _, b := range g.bullets {
		if g.world.Positions[b] == nil {
			continue
		}
		if containsEntity(remainingBullets, b) {
			continue
		}
		g.removeEntity(b)
	}
	g.bullets = remainingBullets
}

func (g *Game) onEnemyHit(e ecs.Entity) {
	anim := g.enemyAnims[e]
	if anim == nil || len(anim.Frames) == 0 {
		g.removeEntity(e)
		return
	}

	frames := explosionFrames(anim)
	if len(frames) == 0 {
		g.removeEntity(e)
		return
	}

	ref := g.world.Sprites[e]
	if ref == nil {
		g.world.AddSprite(e, frames[0], 1)
	} else {
		ref.Sprite = frames[0]
	}
	if vel := g.world.Velocities[e]; vel != nil {
		vel.DX = 0
		vel.DY = 0
	}
	g.explosions = append(g.explosions, explosion{
		entity: e,
		frames: frames,
	})
}

func (g *Game) updateExplosions(dt float64) {
	if len(g.explosions) == 0 {
		return
	}
	remaining := g.explosions[:0]
	for _, ex := range g.explosions {
		ex.elapsed += dt
		frame := int(ex.elapsed * explodeFPS)
		if frame >= len(ex.frames) {
			g.removeEntity(ex.entity)
			continue
		}
		ref := g.world.Sprites[ex.entity]
		if ref != nil {
			ref.Sprite = ex.frames[frame]
		}
		remaining = append(remaining, ex)
	}
	g.explosions = remaining
}

func explosionFrames(anim *render.Animation) []*render.Sprite {
	if anim == nil {
		return nil
	}
	if len(anim.Frames) <= 1 {
		return anim.Frames
	}
	return anim.Frames[1:]
}

func (g *Game) checkNextLevel() {
	if len(g.enemies) > 0 || len(g.explosions) > 0 {
		return
	}
	g.level++
	g.enemiesPlaced = false
	g.layoutEnemies()
}

func containsEntity(list []ecs.Entity, target ecs.Entity) bool {
	for _, e := range list {
		if e == target {
			return true
		}
	}
	return false
}

func (g *Game) clampShip() {
	if g.screenW <= 0 {
		return
	}
	pos := g.world.Positions[g.ship]
	if pos == nil {
		return
	}
	maxX := float64(g.screenW - g.shipSprite.W)
	if pos.X < 0 {
		pos.X = 0
	}
	if pos.X > maxX {
		pos.X = maxX
	}
}

func (g *Game) layoutEnemies() {
	if g.screenW == 0 || g.enemySprite == nil {
		return
	}
	g.enemyDir = 1
	for e := range g.enemyTargets {
		delete(g.enemyTargets, e)
	}
	cols := 8
	maxCols := (g.screenW + enemyGapX) / (g.enemyMaxW + enemyGapX)
	if maxCols < 1 {
		maxCols = 1
	}
	if cols > maxCols {
		cols = maxCols
	}
	totalW := cols*g.enemyMaxW + (cols-1)*enemyGapX
	startX := (g.screenW - totalW) / 2
	startY := enemyStartY
	rowGap := g.enemyMaxH + enemyGapY

	for row := 0; row < enemyRows; row++ {
		y := startY + row*rowGap
		for col := 0; col < cols; col++ {
			x := startX + col*(g.enemyMaxW+enemyGapX)
			targetX := float64(x)
			enemy := g.world.NewEntity()
			sprite := g.enemySprite
			anim := g.enemyDestroy
			if g.rng != nil && g.rng.Float64() < enemy2Chance {
				sprite = g.enemy2Sprite
				anim = g.enemy2Destroy
			}
			flyOffset := float64(g.screenW + g.enemyMaxW)
			startX := targetX - flyOffset
			if row%2 == 1 {
				startX = targetX + flyOffset
			}
			g.world.AddPosition(enemy, startX, float64(y))
			g.world.AddVelocity(enemy, 0, 0)
			g.world.AddSprite(enemy, sprite, 1)
			g.enemies = append(g.enemies, enemy)
			g.enemyAnims[enemy] = anim
			g.enemyTargets[enemy] = targetX
		}
	}
	g.enemiesPlaced = true
}

func (g *Game) removeEntity(e ecs.Entity) {
	delete(g.world.Positions, e)
	delete(g.world.Velocities, e)
	delete(g.world.Sprites, e)
	delete(g.world.TileMaps, e)
	delete(g.enemyAnims, e)
}

func (g *Game) pressed(action input.Action) bool {
	return g.actions.Pressed[action]
}

func (g *Game) held(action input.Action) bool {
	return g.actions.Held[action]
}

func main() {
	game := NewGame()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
