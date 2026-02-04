package engine

import (
	"fmt"
	"time"

	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/term"
	"github.com/gdamore/tcell/v3"
)

type Game interface {
	Update(dt float64)
	Draw(r *render.Renderer)
	Resize(w, h int)
}

type InputAware interface {
	SetInput(state input.State)
}

type ActionAware interface {
	ActionMap() input.ActionMap
	UpdateActionState(state input.ActionState)
}

type Quitter interface {
	ShouldQuit() bool
}

type Engine struct {
	Screen   *term.Screen
	Renderer *render.Renderer
	Frame    *grid.Frame
	Tick     time.Duration
	Input    *input.Manager
	ShowFPS  bool

	fpsWindow []float64
	fpsIndex  int
	fpsCount  int
	fpsSum    float64
	fpsValue  float64
}

func New(game Game, tick time.Duration) (*Engine, error) {
	screen, err := term.NewScreen()
	if err != nil {
		return nil, err
	}
	w, h := screen.Size()
	clear := grid.Cell{Ch: ' ', Style: grid.Style{Fg: tcell.ColorReset, Bg: tcell.ColorReset}}
	frame := grid.NewFrame(w, h, clear)
	r := render.NewRenderer(frame)
	if tick <= 0 {
		tick = 16 * time.Millisecond
	}
	game.Resize(w, h)
	return &Engine{
		Screen:    screen,
		Renderer:  r,
		Frame:     frame,
		Tick:      tick,
		Input:     input.New(0.12),
		fpsWindow: make([]float64, 60),
	}, nil
}

func (e *Engine) Run(game Game) error {
	defer e.Screen.Fini()

	events := e.Screen.Events()
	ticker := time.NewTicker(e.Tick)
	defer ticker.Stop()
	last := time.Now()
	accumulator := 0.0
	step := e.Tick.Seconds()
	if step <= 0 {
		step = 1.0 / 30.0
	}
	const maxAccum = 0.25

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(last).Seconds()
			last = now

			// Drain any pending input events to avoid starving input when tick is busy.
			for {
				select {
				case ev := <-events:
					switch ev := ev.(type) {
					case *tcell.EventResize:
						e.Screen.Sync()
						w, h := e.Screen.Size()
						e.Frame.Resize(w, h)
						e.Renderer.SetFrame(e.Frame)
						e.Screen.Clear()
						game.Resize(w, h)
					default:
						e.Input.HandleEvent(ev)
					}
				default:
					goto updates
				}
			}

		updates:
			state := e.Input.Step(dt)
			if ia, ok := game.(InputAware); ok {
				ia.SetInput(state)
			}
			if aa, ok := game.(ActionAware); ok {
				actions := aa.ActionMap()
				if actions != nil {
					mapper := input.Mapper{Map: actions}
					aa.UpdateActionState(mapper.MapState(state))
				} else {
					aa.UpdateActionState(input.ActionState{})
				}
			}

			accumulator += dt
			if accumulator > maxAccum {
				accumulator = maxAccum
			}

			e.updateFPS(dt)

			for accumulator >= step {
				game.Update(step)
				accumulator -= step
			}
			if q, ok := game.(Quitter); ok && q.ShouldQuit() {
				return nil
			}

			e.Renderer.Clear()
			game.Draw(e.Renderer)
			e.drawFPSOverlay()
			e.Screen.Present(e.Renderer.Frame)
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				e.Screen.Sync()
				w, h := e.Screen.Size()
				e.Frame.Resize(w, h)
				e.Renderer.SetFrame(e.Frame)
				e.Screen.Clear()
				game.Resize(w, h)
			default:
				e.Input.HandleEvent(ev)
			}
		}
	}
}

func (e *Engine) updateFPS(dt float64) {
	if dt <= 0 {
		return
	}
	if e.fpsCount < len(e.fpsWindow) {
		e.fpsWindow[e.fpsIndex] = dt
		e.fpsSum += dt
		e.fpsCount++
	} else {
		e.fpsSum -= e.fpsWindow[e.fpsIndex]
		e.fpsWindow[e.fpsIndex] = dt
		e.fpsSum += dt
	}
	e.fpsIndex = (e.fpsIndex + 1) % len(e.fpsWindow)
	if e.fpsSum > 0 {
		e.fpsValue = float64(e.fpsCount) / e.fpsSum
	}
}

func (e *Engine) drawFPSOverlay() {
	if !e.ShowFPS {
		return
	}
	style := e.Frame.Clear.Style
	text := fmt.Sprintf("FPS: %.1f", e.fpsValue)
	e.Renderer.DrawText(0, 0, text, style)
}

func (e *Engine) FPS() float64 {
	return e.fpsValue
}
