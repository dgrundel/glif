package camera

type Camera struct {
	X     float64
	Y     float64
	ViewW int
	ViewH int
}

func (c Camera) WorldToScreen(x, y float64) (float64, float64) {
	return x - c.X, y - c.Y
}

func (c Camera) InView(x, y float64, w, h int) bool {
	if w <= 0 || h <= 0 || c.ViewW <= 0 || c.ViewH <= 0 {
		return false
	}
	left := x
	top := y
	right := x + float64(w)
	bottom := y + float64(h)

	if right <= 0 || bottom <= 0 {
		return false
	}
	if left >= float64(c.ViewW) || top >= float64(c.ViewH) {
		return false
	}
	return true
}
