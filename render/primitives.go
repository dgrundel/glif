package render

import (
	"github.com/dgrundel/glif/grid"
	"github.com/gdamore/tcell/v3"
)

type RectOptions struct {
	VLine    rune
	HLine    rune
	TLCorner rune
	TRCorner rune
	BLCorner rune
	BRCorner rune
	Fill     bool
	FillRune rune
}

type LineOptions struct {
	Rune rune
}

// Rect draws a rectangle. When Fill is true, the rectangle area is filled and no border is drawn.
// When Fill is false, only the border is drawn.
func (r *Renderer) Rect(x, y, w, h int, style grid.Style, opts ...RectOptions) {
	if r == nil || r.Frame == nil || w <= 0 || h <= 0 {
		return
	}
	opt := rectDefaults()
	if len(opts) > 0 {
		opt = mergeRectOptions(opt, opts[0])
	}

	if opt.Fill {
		fillRune := opt.FillRune
		if fillRune == 0 {
			fillRune = ' '
		}
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				r.Frame.Set(x+col, y+row, grid.Cell{Ch: fillRune, Style: style})
			}
		}
		return
	}

	if w < 2 || h < 2 {
		return
	}
	r.Frame.Set(x, y, grid.Cell{Ch: opt.TLCorner, Style: style})
	r.Frame.Set(x+w-1, y, grid.Cell{Ch: opt.TRCorner, Style: style})
	r.Frame.Set(x, y+h-1, grid.Cell{Ch: opt.BLCorner, Style: style})
	r.Frame.Set(x+w-1, y+h-1, grid.Cell{Ch: opt.BRCorner, Style: style})
	for i := 1; i < w-1; i++ {
		r.Frame.Set(x+i, y, grid.Cell{Ch: opt.HLine, Style: style})
		r.Frame.Set(x+i, y+h-1, grid.Cell{Ch: opt.HLine, Style: style})
	}
	for j := 1; j < h-1; j++ {
		r.Frame.Set(x, y+j, grid.Cell{Ch: opt.VLine, Style: style})
		r.Frame.Set(x+w-1, y+j, grid.Cell{Ch: opt.VLine, Style: style})
	}
}

// HLine draws a horizontal line.
func (r *Renderer) HLine(x, y, length int, style grid.Style, opts ...LineOptions) {
	if r == nil || r.Frame == nil || length <= 0 {
		return
	}
	ch := tcell.RuneHLine
	if len(opts) > 0 && opts[0].Rune != 0 {
		ch = opts[0].Rune
	}
	for i := 0; i < length; i++ {
		r.Frame.Set(x+i, y, grid.Cell{Ch: ch, Style: style})
	}
}

// VLine draws a vertical line.
func (r *Renderer) VLine(x, y, length int, style grid.Style, opts ...LineOptions) {
	if r == nil || r.Frame == nil || length <= 0 {
		return
	}
	ch := tcell.RuneVLine
	if len(opts) > 0 && opts[0].Rune != 0 {
		ch = opts[0].Rune
	}
	for i := 0; i < length; i++ {
		r.Frame.Set(x, y+i, grid.Cell{Ch: ch, Style: style})
	}
}

func rectDefaults() RectOptions {
	return RectOptions{
		VLine:    tcell.RuneVLine,
		HLine:    tcell.RuneHLine,
		TLCorner: tcell.RuneULCorner,
		TRCorner: tcell.RuneURCorner,
		BLCorner: tcell.RuneLLCorner,
		BRCorner: tcell.RuneLRCorner,
		FillRune: ' ',
	}
}

func mergeRectOptions(base, in RectOptions) RectOptions {
	if in.VLine != 0 {
		base.VLine = in.VLine
	}
	if in.HLine != 0 {
		base.HLine = in.HLine
	}
	if in.TLCorner != 0 {
		base.TLCorner = in.TLCorner
	}
	if in.TRCorner != 0 {
		base.TRCorner = in.TRCorner
	}
	if in.BLCorner != 0 {
		base.BLCorner = in.BLCorner
	}
	if in.BRCorner != 0 {
		base.BRCorner = in.BRCorner
	}
	if in.FillRune != 0 {
		base.FillRune = in.FillRune
	}
	if in.Fill {
		base.Fill = true
	}
	return base
}
