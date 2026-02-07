package main

import (
	"fmt"
	"log"
	"math"

	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type Picker struct {
	binds    input.ActionMap
	quit     bool
	actions  input.ActionState
	selected int
	values   [3]int
	incHeld  float64
	decHeld  float64
	incSpeed float64
	decSpeed float64
	uiStyle  grid.Style
}

func NewPicker() *Picker {
	pal, err := palette.Load("demos/picker/assets/default.palette")
	if err != nil {
		log.Fatal(err)
	}
	uiStyle, err := pal.Style('t')
	if err != nil {
		log.Fatal(err)
	}
	return &Picker{
		values: [3]int{100, 255, 105},
		binds: input.ActionMap{
			"select_up":   "key:up",
			"select_down": "key:down",
			"dec":         "key:left",
			"inc":         "key:right",
			"quit":        "key:esc",
			"quit_alt":    "key:ctrl+c",
		},
		uiStyle: uiStyle,
	}
}

func (p *Picker) Update(dt float64) {
	if p.pressed("select_up") {
		p.selected = (p.selected + 2) % 3
	}
	if p.pressed("select_down") {
		p.selected = (p.selected + 1) % 3
	}

	p.adjustValue(dt)

	if p.pressed("quit") || p.pressed("quit_alt") {
		p.quit = true
	}
}

func (p *Picker) Draw(r *render.Renderer) {
	w := r.Frame.W
	swatchW := 16
	swatchH := 8

	labelX := 2
	selectorX := 0
	barX := 10
	numWidth := 3
	gapNum := 2
	gapSwatch := 3
	swatchX := w - swatchW - 2
	barWidth := swatchX - gapSwatch - (barX + numWidth + gapNum)
	if barWidth < 10 {
		barWidth = 10
		swatchX = barX + barWidth + numWidth + gapNum + gapSwatch
	}
	if swatchX+swatchW+1 > w {
		swatchW = max(6, w-swatchX-2)
	}

	startY := 1
	rowGap := 1
	barH := 3

	labels := []string{"Red", "Green", "Blue"}
	colors := []tcell.Color{
		tcell.NewRGBColor(int32(p.values[0]), 0, 0),
		tcell.NewRGBColor(0, int32(p.values[1]), 0),
		tcell.NewRGBColor(0, 0, int32(p.values[2])),
	}

	boxStyle := grid.Style{Fg: grid.TCellColor(tcell.ColorWhite), Bg: grid.TCellColor(tcell.ColorReset)}
	textStyle := p.uiStyle

	for i := 0; i < 3; i++ {
		y := startY + i*(barH+rowGap)
		if i == p.selected {
			r.DrawText(selectorX, y+1, "▶", textStyle)
		}
		r.DrawText(labelX, y+1, labels[i], textStyle)

		r.Rect(barX, y, barWidth, barH, boxStyle)
		innerW := barWidth - 2
		fill := 0
		if innerW > 0 {
			fill = (p.values[i] * innerW) / 255
		}
		barStyle := grid.Style{Fg: grid.TCellColor(colors[i]), Bg: grid.TCellColor(tcell.ColorReset)}
		for x := 0; x < fill; x++ {
			r.Frame.Set(barX+1+x, y+1, grid.Cell{Ch: '█', Style: barStyle})
		}

		valueText := fmt.Sprintf("%3d", p.values[i])
		r.DrawText(barX+barWidth+gapNum, y+1, valueText, textStyle)
	}

	swatchY := startY
	color := tcell.NewRGBColor(int32(p.values[0]), int32(p.values[1]), int32(p.values[2]))
	swatchStyle := grid.Style{Fg: grid.TCellColor(color), Bg: grid.TCellColor(color)}
	r.Rect(swatchX, swatchY, swatchW, swatchH, swatchStyle, render.RectOptions{Fill: true})
	r.Rect(swatchX, swatchY, swatchW, swatchH, boxStyle)

	hex := fmt.Sprintf("#%02x%02x%02x", p.values[0], p.values[1], p.values[2])
	r.DrawText(swatchX+2, swatchY+swatchH+1, hex, textStyle)
}

func (p *Picker) Resize(w, h int) {
	_ = w
	_ = h
}

func (p *Picker) ActionMap() input.ActionMap {
	return p.binds
}

func (p *Picker) UpdateActionState(state input.ActionState) {
	p.actions = state
}

func (p *Picker) ShouldQuit() bool {
	return p.quit
}

func (p *Picker) pressed(action input.Action) bool {
	return p.actions.Pressed[action]
}

func (p *Picker) held(action input.Action) bool {
	return p.actions.Held[action]
}

func (p *Picker) adjustValue(dt float64) {
	delta := 0
	if p.pressed("dec") {
		delta -= 1
		p.decHeld = 0
	}
	if p.pressed("inc") {
		delta += 1
		p.incHeld = 0
	}

	const repeatDelay = 0.25
	const repeatRate = 18.0
	if p.held("dec") {
		p.decHeld += dt
		if p.decHeld >= repeatDelay {
			p.decSpeed = math.Min(p.decSpeed+dt*24.0, 80.0)
			steps := int(math.Floor((p.decHeld - repeatDelay) * p.decSpeed))
			if steps > 0 {
				delta -= steps
				p.decHeld -= float64(steps) / p.decSpeed
			}
		}
	} else {
		p.decHeld = 0
		p.decSpeed = repeatRate
	}
	if p.held("inc") {
		p.incHeld += dt
		if p.incHeld >= repeatDelay {
			p.incSpeed = math.Min(p.incSpeed+dt*24.0, 80.0)
			steps := int(math.Floor((p.incHeld - repeatDelay) * p.incSpeed))
			if steps > 0 {
				delta += steps
				p.incHeld -= float64(steps) / p.incSpeed
			}
		}
	} else {
		p.incHeld = 0
		p.incSpeed = repeatRate
	}

	if delta == 0 {
		return
	}
	p.values[p.selected] = clamp(p.values[p.selected]+delta, 0, 255)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	picker := NewPicker()
	eng, err := engine.New(picker, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err := eng.Run(picker); err != nil {
		log.Fatal(err)
	}
}
