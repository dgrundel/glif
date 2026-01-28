package render

import "github.com/dgrundel/glif/grid"

type Renderer struct {
	Frame *grid.Frame
}

func NewRenderer(frame *grid.Frame) *Renderer {
	return &Renderer{Frame: frame}
}

func (r *Renderer) SetFrame(frame *grid.Frame) {
	r.Frame = frame
}

func (r *Renderer) Clear() {
	r.Frame.ClearAll()
}

func (r *Renderer) DrawSprite(x, y int, sprite *Sprite) {
	for row := 0; row < sprite.H; row++ {
		for col := 0; col < sprite.W; col++ {
			cell := sprite.cellAt(col, row)
			if cell.Ch == 0 {
				continue
			}
			if sprite.Transparent != 0 && cell.Ch == sprite.Transparent {
				continue
			}
			r.Frame.Set(x+col, y+row, cell)
		}
	}
}

func (r *Renderer) DrawText(x, y int, text string, style grid.Style) {
	cx := x
	for _, ch := range text {
		if ch == '\n' {
			y++
			cx = x
			continue
		}
		r.Frame.Set(cx, y, grid.Cell{Ch: ch, Style: style})
		cx++
	}
}
