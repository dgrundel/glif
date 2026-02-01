package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type ObstacleKind int

const (
	ObstacleTree ObstacleKind = iota
	ObstacleRough
)

type Gate struct {
	Y      float64
	LeftX  float64
	RightX float64
	Passed bool
}

type Obstacle struct {
	X      float64
	Y      float64
	Kind   ObstacleKind
	Hit    bool
	Sprite *render.Sprite
}

type SkiGame struct {
	state input.State
	binds input.ActionMap
	quit  bool

	score          int
	gameOver       bool
	gameOverReason string

	playerX    float64
	playerY    float64
	playerSpr  *render.Sprite
	skiDown    *render.Sprite
	skiLeft    *render.Sprite
	skiRight   *render.Sprite
	flagLeft   *render.Sprite
	flagRight  *render.Sprite
	tree       *render.Sprite
	rough      *render.Sprite
	uiStyle    grid.Style
	width      int
	height     int
	screenPX   int
	screenPY   int
	speed      float64
	targetSpd  float64
	speedPen   float64
	nextGateY  float64
	lastGateCX float64
	leftTTL    float64
	rightTTL   float64
	gates      []Gate
	obstacles  []Obstacle
	rng        *rand.Rand
}

func NewSkiGame() *SkiGame {
	load := func(name string) *render.Sprite {
		sprite, err := assets.LoadMaskedSprite("demos/ski/assets/" + name)
		if err != nil {
			log.Fatal(err)
		}
		return sprite
	}
	pal, err := palette.Load("demos/ski/assets/default.palette")
	if err != nil {
		log.Fatal(err)
	}
	uiStyle, err := pal.Style('x')
	if err != nil {
		log.Fatal(err)
	}

	g := &SkiGame{
		binds: input.ActionMap{
			"left":        "key:left",
			"right":       "key:right",
			"left_alt":    "a",
			"right_alt":   "d",
			"quit":        "key:esc",
			"quit_alt":    "key:ctrl+c",
			"restart":     "key:enter",
			"restart_alt": " ",
		},
		skiDown:   load("ski_down"),
		skiLeft:   load("ski_left"),
		skiRight:  load("ski_right"),
		flagLeft:  load("flag_left"),
		flagRight: load("flag_right"),
		tree:      load("trees"),
		rough:     load("snow_rough"),
		uiStyle:   uiStyle,
		speed:     8.4,
		targetSpd: 8.4,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.playerSpr = g.skiDown
	g.reset()
	return g
}

func (g *SkiGame) reset() {
	g.score = 0
	g.gameOver = false
	g.gameOverReason = ""
	g.playerX = 0
	g.playerY = 0
	g.speed = 8.4
	g.targetSpd = 8.4
	g.speedPen = 0
	g.nextGateY = 0
	g.lastGateCX = g.playerX
	g.gates = nil
	g.obstacles = nil
}

func (g *SkiGame) Update(dt float64) {
	if g.pressed("quit") || g.pressed("quit_alt") {
		g.quit = true
		return
	}
	if g.gameOver {
		if g.pressed("restart") || g.pressed("restart_alt") {
			g.reset()
		}
		return
	}

	g.tickSteerTTL(dt)
	left := g.steerHeld("left", "left_alt")
	right := g.steerHeld("right", "right_alt")
	moveSpd := g.speed - g.speedPen
	if moveSpd < 0 {
		moveSpd = 0
	}

	strafe := maxFloat(16.0, moveSpd*0.9)
	if left && !right {
		g.playerSpr = g.skiLeft
		g.playerX -= strafe * dt
	} else if right && !left {
		g.playerSpr = g.skiRight
		g.playerX += strafe * dt
	} else {
		g.playerSpr = g.skiDown
	}

	distance := g.playerY
	g.targetSpd = 8.4 + math.Min(7.0, distance/120.0)
	if g.speed < g.targetSpd {
		g.speed += 2.5 * dt
		if g.speed > g.targetSpd {
			g.speed = g.targetSpd
		}
	} else if g.speed > g.targetSpd {
		g.speed -= 1.0 * dt
		if g.speed < g.targetSpd {
			g.speed = g.targetSpd
		}
	}
	if g.speedPen > 0 {
		g.speedPen -= 1.0 * dt
		if g.speedPen < 0 {
			g.speedPen = 0
		}
	}
	g.playerY += moveSpd * dt

	g.generateWorld()
	g.checkGates()
	g.checkObstacles()
	g.trimWorld()
}

func (g *SkiGame) Draw(r *render.Renderer) {
	cameraX := g.playerX - float64(g.screenPX)
	cameraY := g.playerY - float64(g.screenPY)

	type drawItem struct {
		x      float64
		y      float64
		sprite *render.Sprite
	}
	items := make([]drawItem, 0, len(g.obstacles)+len(g.gates)*2)

	for _, gate := range g.gates {
		items = append(items, drawItem{x: gate.LeftX, y: gate.Y, sprite: g.flagLeft})
		items = append(items, drawItem{x: gate.RightX, y: gate.Y, sprite: g.flagRight})
	}
	for _, obs := range g.obstacles {
		items = append(items, drawItem{x: obs.X, y: obs.Y, sprite: obs.Sprite})
	}

	for _, it := range items {
		sx := int(math.Round(it.x - cameraX))
		sy := int(math.Round(it.y - cameraY))
		if sx+it.sprite.W < 0 || sx >= r.Frame.W || sy+it.sprite.H < 0 || sy >= r.Frame.H {
			continue
		}
		r.DrawSprite(sx, sy, it.sprite)
	}

	px := g.screenPX
	py := g.screenPY
	r.DrawSprite(px, py, g.playerSpr)

	scoreText := fmt.Sprintf("Score: %d", g.score)
	r.DrawText(r.Frame.W-len(scoreText)-2, 0, scoreText, g.uiStyle)

	if g.gameOver {
		msg := "Press Enter or Space to restart."
		if g.gameOverReason != "" {
			msg = fmt.Sprintf("%s. Press Enter or Space to restart.", g.gameOverReason)
		}
		x := max(0, (r.Frame.W-len(msg))/2)
		y := max(0, g.screenPY-2)
		r.DrawText(x, y, msg, g.uiStyle)
	}
}

func (g *SkiGame) Resize(w, h int) {
	g.width = w
	g.height = h
	if g.playerSpr == nil {
		return
	}
	g.screenPX = max(0, (w-g.playerSpr.W)/2)
	g.screenPY = max(0, h/5)
}

func (g *SkiGame) SetInput(state input.State) {
	g.state = state
}

func (g *SkiGame) ShouldQuit() bool {
	return g.quit
}

func (g *SkiGame) pressed(action input.Action) bool {
	key, ok := g.binds[action]
	if !ok {
		return false
	}
	return g.state.Pressed[key]
}

func (g *SkiGame) held(action input.Action) bool {
	key, ok := g.binds[action]
	if !ok {
		return false
	}
	return g.state.Held[key]
}

func (g *SkiGame) tickSteerTTL(dt float64) {
	if g.pressed("left") || g.pressed("left_alt") {
		g.leftTTL = 0.35
	}
	if g.pressed("right") || g.pressed("right_alt") {
		g.rightTTL = 0.35
	}
	if g.leftTTL > 0 {
		g.leftTTL -= dt
		if g.leftTTL < 0 {
			g.leftTTL = 0
		}
	}
	if g.rightTTL > 0 {
		g.rightTTL -= dt
		if g.rightTTL < 0 {
			g.rightTTL = 0
		}
	}
}

func (g *SkiGame) steerHeld(primary, alt input.Action) bool {
	if g.held(primary) || g.held(alt) {
		return true
	}
	if primary == "left" {
		return g.leftTTL > 0
	}
	return g.rightTTL > 0
}

func (g *SkiGame) generateWorld() {
	lookahead := float64(g.height + 10)
	if g.nextGateY == 0 {
		g.nextGateY = g.playerY + 12
		g.lastGateCX = g.playerX
	}
	for g.nextGateY < g.playerY+lookahead {
		spacing := g.gateSpacing()
		center := g.nextGateCenter(spacing)
		leftX, rightX := g.gatePositions(center)
		g.gates = append(g.gates, Gate{
			Y:      g.nextGateY,
			LeftX:  leftX,
			RightX: rightX,
		})
		g.spawnObstacles(g.nextGateY, spacing, center)
		g.lastGateCX = center
		g.nextGateY += spacing
	}
}

func (g *SkiGame) gateSpacing() float64 {
	spacing := 16.0 - g.playerY/100.0
	if spacing < 8.0 {
		spacing = 8.0
	}
	return spacing
}

func (g *SkiGame) nextGateCenter(spacing float64) float64 {
	maxShift := spacing
	shift := (g.rng.Float64()*2 - 1) * maxShift
	return g.lastGateCX + shift
}

func (g *SkiGame) gatePositions(center float64) (float64, float64) {
	gap := 6.0
	leftX := center - gap/2.0 - float64(g.flagLeft.W)
	rightX := leftX + float64(g.flagLeft.W) + gap
	return leftX, rightX
}

func (g *SkiGame) spawnObstacles(gateY, spacing, center float64) {
	startY := gateY - spacing + 2
	endY := gateY - 2
	leftWorld := g.playerX - float64(g.screenPX)
	rightWorld := leftWorld + float64(g.width)
	for y := startY; y <= endY; y += 2 {
		if math.Abs(y-gateY) <= 2 {
			continue
		}
		if y <= 12 {
			continue
		}
		if g.rng.Float64() < 0.28 {
			x := leftWorld + g.rng.Float64()*maxFloat(0, rightWorld-leftWorld-float64(g.tree.W))
			dist := math.Abs(x - g.playerX)
			scale := 0.6 + minFloat(dist/14.0, 1.2)
			if g.rng.Float64() < minFloat(scale, 1.0) {
				g.obstacles = append(g.obstacles, Obstacle{X: x, Y: y, Kind: ObstacleTree, Sprite: g.tree})
			}
		}
		if g.rng.Float64() < 0.16 {
			x := leftWorld + g.rng.Float64()*maxFloat(0, rightWorld-leftWorld-float64(g.rough.W))
			g.obstacles = append(g.obstacles, Obstacle{X: x, Y: y, Kind: ObstacleRough, Sprite: g.rough})
		}
	}
}

func (g *SkiGame) checkGates() {
	for i := range g.gates {
		gate := &g.gates[i]
		if gate.Passed {
			continue
		}
		gateBottom := gate.Y + float64(g.flagLeft.H) - 1
		playerBottom := g.playerY + float64(g.playerSpr.H) - 1
		rows := max(2, g.playerSpr.H/2)
		bandTop := playerBottom - float64(rows-1)
		if playerBottom < gateBottom || bandTop > gateBottom {
			continue
		}
		spanLeft := gate.LeftX
		spanRight := gate.RightX + float64(g.flagRight.W) - 1
		playerLeft := g.playerX
		playerRight := g.playerX + float64(g.playerSpr.W) - 1
		overlap := playerRight >= spanLeft && playerLeft <= spanRight
		if overlap {
			g.score++
			gate.Passed = true
		} else {
			g.gameOver = true
			g.gameOverReason = "Missed gate"
			return
		}
	}
}

func (g *SkiGame) checkObstacles() {
	playerRect := rect{
		x: g.playerX,
		y: g.playerY,
		w: float64(g.playerSpr.W),
		h: float64(g.playerSpr.H),
	}
	for i := range g.obstacles {
		obs := &g.obstacles[i]
		if obs.Hit {
			continue
		}
		obsRect := rect{
			x: obs.X,
			y: obs.Y,
			w: float64(obs.Sprite.W),
			h: float64(obs.Sprite.H),
		}
		if !playerRect.overlaps(obsRect) {
			continue
		}
		obs.Hit = true
		if obs.Kind == ObstacleTree {
			g.gameOver = true
			g.gameOverReason = "Hit a tree"
			return
		}
		if obs.Kind == ObstacleRough {
			g.speedPen = maxFloat(g.speedPen, 2.0)
		}
	}
}

func (g *SkiGame) trimWorld() {
	minY := g.playerY - float64(g.height) - 10
	filterGates := g.gates[:0]
	for _, gate := range g.gates {
		if gate.Y >= minY {
			filterGates = append(filterGates, gate)
		}
	}
	g.gates = filterGates

	filterObs := g.obstacles[:0]
	for _, obs := range g.obstacles {
		if obs.Y >= minY && !obs.Hit {
			filterObs = append(filterObs, obs)
		}
	}
	g.obstacles = filterObs
}

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

func (r rect) overlaps(o rect) bool {
	return r.x < o.x+o.w && r.x+r.w > o.x && r.y < o.y+o.h && r.y+r.h > o.y
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	game := NewSkiGame()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: grid.Style{Fg: tcell.ColorReset, Bg: tcell.ColorWhite}}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
