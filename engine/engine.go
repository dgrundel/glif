package engine

import (
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
	HandleEvent(ev tcell.Event) (quit bool)
	Resize(w, h int)
}

type InputAware interface {
	SetInput(state input.State)
}

type Engine struct {
	Screen   *term.Screen
	Renderer *render.Renderer
	Frame    *grid.Frame
	Tick     time.Duration
	Input    *input.Manager
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
		tick = 33 * time.Millisecond
	}
	game.Resize(w, h)
	return &Engine{
		Screen:   screen,
		Renderer: r,
		Frame:    frame,
		Tick:     tick,
		Input:    input.New(0.12),
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
						game.Resize(w, h)
					default:
						e.Input.HandleEvent(ev)
						if game.HandleEvent(ev) {
							return nil
						}
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

			accumulator += dt
			if accumulator > maxAccum {
				accumulator = maxAccum
			}

			for accumulator >= step {
				game.Update(step)
				accumulator -= step
			}

			e.Renderer.Clear()
			game.Draw(e.Renderer)
			e.Screen.Present(e.Renderer.Frame)
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				e.Screen.Sync()
				w, h := e.Screen.Size()
				e.Frame.Resize(w, h)
				e.Renderer.SetFrame(e.Frame)
				game.Resize(w, h)
			default:
				e.Input.HandleEvent(ev)
				if game.HandleEvent(ev) {
					return nil
				}
			}
		}
	}
}
