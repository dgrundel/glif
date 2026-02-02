package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/collision"
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
	ObstacleSnow
)

const (
	initialSpeed             = 8.4
	speedRampCap             = 7.0
	speedRampDistance        = 120.0
	speedAccel               = 2.5
	speedDecel               = 1.0
	speedPenaltyDecay        = 1.0
	roughSpeedPenalty        = 2.0
	strafeBaseSpeed          = 16.0
	strafeSpeedScale         = 0.9
	steerTTLSeconds          = 0.35
	gateStartOffset          = 12.0
	gateSpacingStart         = 16.0
	gateSpacingMin           = 8.0
	gateSpacingTightenFactor = 100.0
	gateGapWidth             = 6.0
	gateNearBuffer           = 2.0
	gateRowsStep             = 2.0
	gateLookaheadPadding     = 10.0
	treeSpawnChance          = 0.38
	roughSpawnChance         = 0.16
	snowSpawnChance          = 0.45
	treeSpawnMinRow          = 12.0
	treeEdgeBiasBase         = 0.6
	treeEdgeBiasScale        = 14.0
	treeEdgeBiasMaxBoost     = 1.2
	screenScorePadRight      = 2
	screenMessageOffsetY     = 2
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
	splash     *render.Sprite
	flagLeft   *render.Sprite
	flagRight  *render.Sprite
	tree       *render.Sprite
	rough      *render.Sprite
	snow       *render.Sprite
	uiStyle    grid.Style
	alertStyle grid.Style
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
	showSplash bool
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
	alertStyle, err := pal.Style('e')
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
		skiDown:    load("ski_down"),
		skiLeft:    load("ski_left"),
		skiRight:   load("ski_right"),
		splash:     load("splash"),
		flagLeft:   load("flag_left"),
		flagRight:  load("flag_right"),
		tree:       load("trees"),
		rough:      load("snow_rough"),
		snow:       load("snow"),
		uiStyle:    uiStyle,
		alertStyle: alertStyle,
		speed:      initialSpeed,
		targetSpd:  initialSpeed,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	g.playerSpr = g.skiDown
	g.showSplash = true
	g.reset()
	return g
}

func (g *SkiGame) reset() {
	g.score = 0
	g.gameOver = false
	g.gameOverReason = ""
	g.playerX = 0
	g.playerY = 0
	g.speed = initialSpeed
	g.targetSpd = initialSpeed
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
	if g.showSplash {
		if g.pressed("restart") || g.pressed("restart_alt") {
			g.showSplash = false
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

	strafe := maxFloat(strafeBaseSpeed, moveSpd*strafeSpeedScale)
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
	g.targetSpd = initialSpeed + math.Min(speedRampCap, distance/speedRampDistance)
	if g.speed < g.targetSpd {
		g.speed += speedAccel * dt
		if g.speed > g.targetSpd {
			g.speed = g.targetSpd
		}
	} else if g.speed > g.targetSpd {
		g.speed -= speedDecel * dt
		if g.speed < g.targetSpd {
			g.speed = g.targetSpd
		}
	}
	if g.speedPen > 0 {
		g.speedPen -= speedPenaltyDecay * dt
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
	if g.showSplash {
		g.drawSplash(r)
		return
	}
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
	r.DrawText(r.Frame.W-len(scoreText)-screenScorePadRight, 0, scoreText, g.uiStyle)

	if g.gameOver {
		msg := "Press Enter or Space to restart."
		if g.gameOverReason != "" {
			msg = fmt.Sprintf("%s. Press Enter or Space to restart.", g.gameOverReason)
		}
		x := max(0, (r.Frame.W-len(msg))/2)
		y := max(0, g.screenPY-screenMessageOffsetY)
		boxW := len(msg) + 2
		boxX := max(0, x-1)
		if boxX+boxW > r.Frame.W {
			boxX = max(0, r.Frame.W-boxW)
		}
		r.Rect(boxX, y-1, boxW, 3, g.alertStyle, render.RectOptions{Fill: true})
		r.DrawText(boxX+1, y, msg, g.alertStyle)
	}
}

func (g *SkiGame) drawSplash(r *render.Renderer) {
	if g.splash == nil {
		return
	}
	startY := max(1, (r.Frame.H-g.splash.H)/2-1)
	startX := max(0, (r.Frame.W-g.splash.W)/2)
	r.DrawSprite(startX, startY, g.splash)
	prompt := "Press enter or space to play"
	promptX := max(0, (r.Frame.W-len(prompt))/2)
	r.DrawText(promptX, startY+g.splash.H+1, prompt, g.uiStyle)
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
		g.leftTTL = steerTTLSeconds
	}
	if g.pressed("right") || g.pressed("right_alt") {
		g.rightTTL = steerTTLSeconds
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
	lookahead := float64(g.height) + gateLookaheadPadding
	if g.nextGateY == 0 {
		g.nextGateY = g.playerY + gateStartOffset
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
	spacing := gateSpacingStart - g.playerY/gateSpacingTightenFactor
	if spacing < gateSpacingMin {
		spacing = gateSpacingMin
	}
	return spacing
}

func (g *SkiGame) nextGateCenter(spacing float64) float64 {
	maxShift := spacing
	shift := (g.rng.Float64()*2 - 1) * maxShift
	return g.lastGateCX + shift
}

func (g *SkiGame) gatePositions(center float64) (float64, float64) {
	gap := gateGapWidth
	leftX := center - gap/2.0 - float64(g.flagLeft.W)
	rightX := leftX + float64(g.flagLeft.W) + gap
	return leftX, rightX
}

func (g *SkiGame) spawnObstacles(gateY, spacing, center float64) {
	startY := gateY - spacing + gateRowsStep
	endY := gateY - gateRowsStep
	leftWorld := g.playerX - float64(g.screenPX)
	rightWorld := leftWorld + float64(g.width)
	for y := startY; y <= endY; y += gateRowsStep {
		if math.Abs(y-gateY) <= gateNearBuffer {
			goto snowOnly
		}
		if y <= treeSpawnMinRow {
			goto snowOnly
		}
		if g.rng.Float64() < treeSpawnChance {
			x := leftWorld + g.rng.Float64()*maxFloat(0, rightWorld-leftWorld-float64(g.tree.W))
			dist := math.Abs(x - g.playerX)
			scale := treeEdgeBiasBase + minFloat(dist/treeEdgeBiasScale, treeEdgeBiasMaxBoost)
			if g.rng.Float64() < minFloat(scale, 1.0) {
				g.obstacles = append(g.obstacles, Obstacle{X: x, Y: y, Kind: ObstacleTree, Sprite: g.tree})
			}
		}
		if g.rng.Float64() < roughSpawnChance {
			x := leftWorld + g.rng.Float64()*maxFloat(0, rightWorld-leftWorld-float64(g.rough.W))
			g.obstacles = append(g.obstacles, Obstacle{X: x, Y: y, Kind: ObstacleRough, Sprite: g.rough})
		}
	snowOnly:
		if g.rng.Float64() < snowSpawnChance {
			x := leftWorld + g.rng.Float64()*maxFloat(0, rightWorld-leftWorld-float64(g.snow.W))
			g.obstacles = append(g.obstacles, Obstacle{X: x, Y: y, Kind: ObstacleSnow, Sprite: g.snow})
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
	playerX := int(math.Round(g.playerX))
	playerY := int(math.Round(g.playerY))
	for i := range g.obstacles {
		obs := &g.obstacles[i]
		if obs.Hit {
			continue
		}
		obsX := int(math.Round(obs.X))
		obsY := int(math.Round(obs.Y))
		if !collision.Overlaps(playerX, playerY, g.playerSpr, obsX, obsY, obs.Sprite) {
			continue
		}
		if obs.Kind == ObstacleTree {
			g.gameOver = true
			g.gameOverReason = "Hit a tree"
			return
		}
		if obs.Kind == ObstacleRough {
			obs.Hit = true
			g.speedPen = maxFloat(g.speedPen, roughSpeedPenalty)
		}
		if obs.Kind == ObstacleSnow {
			continue
		}
	}
}

func (g *SkiGame) trimWorld() {
	minY := g.playerY - float64(g.height) - gateLookaheadPadding
	filterGates := g.gates[:0]
	for _, gate := range g.gates {
		if gate.Y >= minY {
			filterGates = append(filterGates, gate)
		}
	}
	g.gates = filterGates

	filterObs := g.obstacles[:0]
	for _, obs := range g.obstacles {
		if obs.Y >= minY {
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
