package engine

import (
	"time"

	"github.com/dgrundel/glif/grid"
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

type Engine struct {
	Screen   *term.Screen
	Renderer *render.Renderer
	Frame    *grid.Frame
	Tick     time.Duration
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
	return &Engine{Screen: screen, Renderer: r, Frame: frame, Tick: tick}, nil
}

func (e *Engine) Run(game Game) error {
	defer e.Screen.Fini()

	events := e.Screen.Events()
	ticker := time.NewTicker(e.Tick)
	defer ticker.Stop()
	last := time.Now()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(last).Seconds()
			last = now
			if dt > 0.1 {
				dt = 0.1
			}

			game.Update(dt)
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
				if game.HandleEvent(ev) {
					return nil
				}
			}
		}
	}
}
