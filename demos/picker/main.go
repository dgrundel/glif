package main

import (
	"fmt"
	"log"
	"math"

	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type Picker struct {
	state    input.State
	binds    input.ActionMap
	quit     bool
	selected int
	values   [3]int
}

func NewPicker() *Picker {
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
	swatchW := 18
	swatchH := 6

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

	boxStyle := grid.Style{Fg: tcell.ColorWhite, Bg: tcell.ColorReset}
	textStyle := grid.Style{Fg: tcell.ColorWhite, Bg: tcell.ColorReset}

	for i := 0; i < 3; i++ {
		y := startY + i*(barH+rowGap)
		if i == p.selected {
			r.DrawText(selectorX, y+1, "▶", textStyle)
		}
		r.DrawText(labelX, y+1, labels[i], textStyle)

		drawBox(r, barX, y, barWidth, barH, boxStyle)
		innerW := barWidth - 2
		fill := 0
		if innerW > 0 {
			fill = (p.values[i] * innerW) / 255
		}
		barStyle := grid.Style{Fg: colors[i], Bg: tcell.ColorReset}
		for x := 0; x < fill; x++ {
			r.Frame.Set(barX+1+x, y+1, grid.Cell{Ch: '█', Style: barStyle})
		}

		valueText := fmt.Sprintf("%3d", p.values[i])
		r.DrawText(barX+barWidth+gapNum, y+1, valueText, textStyle)
	}

	swatchY := startY
	color := tcell.NewRGBColor(int32(p.values[0]), int32(p.values[1]), int32(p.values[2]))
	swatchStyle := grid.Style{Fg: color, Bg: color}
	fillRect(r, swatchX, swatchY, swatchW, swatchH, swatchStyle)

	hex := fmt.Sprintf("#%02x%02x%02x", p.values[0], p.values[1], p.values[2])
	r.DrawText(swatchX+2, swatchY+swatchH+1, hex, textStyle)
}

func (p *Picker) Resize(w, h int) {
	_ = w
	_ = h
}

func (p *Picker) SetInput(state input.State) {
	p.state = state
}

func (p *Picker) ShouldQuit() bool {
	return p.quit
}

func (p *Picker) pressed(action input.Action) bool {
	key, ok := p.binds[action]
	if !ok {
		return false
	}
	return p.state.Pressed[key]
}

func (p *Picker) held(action input.Action) bool {
	key, ok := p.binds[action]
	if !ok {
		return false
	}
	return p.state.Held[key]
}

func (p *Picker) adjustValue(dt float64) {
	delta := 0
	if p.pressed("dec") {
		delta -= 1
	}
	if p.pressed("inc") {
		delta += 1
	}

	const rate = 60.0
	step := int(math.Max(1, math.Round(rate*dt)))
	if p.held("dec") {
		delta -= step
	}
	if p.held("inc") {
		delta += step
	}

	if delta == 0 {
		return
	}
	p.values[p.selected] = clamp(p.values[p.selected]+delta, 0, 255)
}

func drawBox(r *render.Renderer, x, y, w, h int, style grid.Style) {
	if w < 2 || h < 2 {
		return
	}
	r.Frame.Set(x, y, grid.Cell{Ch: tcell.RuneULCorner, Style: style})
	r.Frame.Set(x+w-1, y, grid.Cell{Ch: tcell.RuneURCorner, Style: style})
	r.Frame.Set(x, y+h-1, grid.Cell{Ch: tcell.RuneLLCorner, Style: style})
	r.Frame.Set(x+w-1, y+h-1, grid.Cell{Ch: tcell.RuneLRCorner, Style: style})

	for i := 1; i < w-1; i++ {
		r.Frame.Set(x+i, y, grid.Cell{Ch: tcell.RuneHLine, Style: style})
		r.Frame.Set(x+i, y+h-1, grid.Cell{Ch: tcell.RuneHLine, Style: style})
	}
	for j := 1; j < h-1; j++ {
		r.Frame.Set(x, y+j, grid.Cell{Ch: tcell.RuneVLine, Style: style})
		r.Frame.Set(x+w-1, y+j, grid.Cell{Ch: tcell.RuneVLine, Style: style})
	}
}

func fillRect(r *render.Renderer, x, y, w, h int, style grid.Style) {
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			r.Frame.Set(x+col, y+row, grid.Cell{Ch: ' ', Style: style})
		}
	}
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
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: grid.Style{Fg: tcell.ColorReset, Bg: tcell.ColorReset}}
	eng.Frame.ClearAll()
	if err := eng.Run(picker); err != nil {
		log.Fatal(err)
	}
}
