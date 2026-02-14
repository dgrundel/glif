package term

import (
	"github.com/dgrundel/glif/grid"
	"github.com/gdamore/tcell/v3"
)

type Screen struct {
	screen tcell.Screen
	front  *grid.Frame
}

func NewScreen() (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	w, h := s.Size()
	clear := grid.Cell{Ch: ' ', Style: grid.Style{Fg: grid.TCellColor(tcell.ColorReset), Bg: grid.TCellColor(tcell.ColorReset)}}
	front := grid.NewFrame(w, h, clear)

	return &Screen{screen: s, front: front}, nil
}

func (s *Screen) Fini() {
	s.screen.Fini()
}

func (s *Screen) Size() (int, int) {
	return s.screen.Size()
}

func (s *Screen) SetTitle(title string) {
	s.screen.SetTitle(title)
}

func (s *Screen) Events() <-chan tcell.Event {
	return s.screen.EventQ()
}

func (s *Screen) Sync() {
	s.screen.Sync()
}

func (s *Screen) EnsureSize(w, h int) {
	if s.front.W == w && s.front.H == h {
		return
	}
	clear := s.front.Clear
	s.front = grid.NewFrame(w, h, clear)
}

func (s *Screen) Clear() {
	s.screen.Clear()
	s.front.ClearAll()
}

func (s *Screen) Present(back *grid.Frame) {
	if back == nil {
		return
	}
	if back.W != s.front.W || back.H != s.front.H {
		s.EnsureSize(back.W, back.H)
	}
	for i := range back.Cells {
		b := back.Cells[i]
		if b.Skip {
			x := i % back.W
			y := i / back.W
			ch := b.Ch
			if ch == 0 {
				ch = ' '
			}
			s.screen.SetContent(x, y, ch, nil, b.Style.ToTCell())
			s.front.Cells[i] = b
			continue
		}
		if b == s.front.Cells[i] {
			continue
		}
		x := i % back.W
		y := i / back.W
		s.screen.SetContent(x, y, b.Ch, nil, b.Style.ToTCell())
		s.front.Cells[i] = b
	}
	s.screen.Show()
}
