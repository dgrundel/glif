package render

import (
	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/grid"
)

type Renderer struct {
	Frame  *grid.Frame
	camera camera.Camera
}

func NewRenderer(frame *grid.Frame) *Renderer {
	return &Renderer{Frame: frame}
}

func (r *Renderer) SetFrame(frame *grid.Frame) {
	r.Frame = frame
}

func (r *Renderer) SetCamera(cam camera.Camera) {
	r.camera = cam
}

func (r *Renderer) Camera() camera.Camera {
	return r.camera
}

func (r *Renderer) WithCamera(cam camera.Camera) *Renderer {
	if r == nil {
		return nil
	}
	return &Renderer{Frame: r.Frame, camera: cam}
}

func (r *Renderer) Clear() {
	r.Frame.ClearAll()
}

func (r *Renderer) DrawSprite(x, y int, sprite *Sprite) {
	if r == nil || r.Frame == nil || sprite == nil {
		return
	}
	if r.camera != nil {
		if !r.camera.Visible(x, y, sprite.W, sprite.H) {
			return
		}
		x, y = r.camera.WorldToScreen(x, y)
	}
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
	if r == nil || r.Frame == nil {
		return
	}
	if r.camera != nil {
		x, y = r.camera.WorldToScreen(x, y)
	}
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
