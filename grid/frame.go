package grid

import "github.com/gdamore/tcell/v3"

type ColorKind int

const (
	ColorTCell ColorKind = iota
	ColorInherit
)

type Color struct {
	Kind       ColorKind
	TCellColor tcell.Color
}

func TCellColor(c tcell.Color) Color {
	return Color{Kind: ColorTCell, TCellColor: c}
}

func InheritColor() Color {
	return Color{Kind: ColorInherit}
}

type Style struct {
	Fg   Color
	Bg   Color
	Bold bool
}

func (s Style) Resolve(base Style) Style {
	if s.Fg.Kind == ColorInherit {
		s.Fg = base.Fg
	}
	if s.Bg.Kind == ColorInherit {
		s.Bg = base.Bg
	}
	return s
}

func (s Style) ToTCell() tcell.Style {
	fg := s.Fg
	bg := s.Bg
	if fg.Kind == ColorInherit {
		fg = TCellColor(tcell.ColorReset)
	}
	if bg.Kind == ColorInherit {
		bg = TCellColor(tcell.ColorReset)
	}
	style := tcell.StyleDefault.Foreground(fg.TCellColor).Background(bg.TCellColor)
	if s.Bold {
		style = style.Bold(true)
	}
	return style
}

type Cell struct {
	Ch    rune
	Style Style
	Skip  bool
}

type Frame struct {
	W     int
	H     int
	Cells []Cell
	Clear Cell
}

func NewFrame(w, h int, clear Cell) *Frame {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	f := &Frame{W: w, H: h, Cells: make([]Cell, w*h), Clear: clear}
	f.ClearAll()
	return f
}

func (f *Frame) Resize(w, h int) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	f.W = w
	f.H = h
	f.Cells = make([]Cell, w*h)
	f.ClearAll()
}

func (f *Frame) ClearAll() {
	for i := range f.Cells {
		f.Cells[i] = f.Clear
	}
}

func (f *Frame) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < f.W && y < f.H
}

func (f *Frame) Set(x, y int, cell Cell) {
	if !f.InBounds(x, y) {
		return
	}
	f.Cells[y*f.W+x] = cell
}

func (f *Frame) At(x, y int) Cell {
	if !f.InBounds(x, y) {
		return f.Clear
	}
	return f.Cells[y*f.W+x]
}
