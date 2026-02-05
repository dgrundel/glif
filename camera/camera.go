package camera

import "math"

// Camera projects world-space coordinates into screen space.
type Camera interface {
	WorldToScreen(x, y float64) (sx, sy float64)
	ScreenToWorld(x, y float64) (wx, wy float64)
	Visible(x, y float64, w, h int) bool
	SetViewport(w, h int)
}

// Basic is a simple top-left anchored camera.
// X,Y are the world-space coordinates of the top-left screen cell.
type Basic struct {
	X float64
	Y float64
	W int
	H int
}

func NewBasic(x, y float64, w, h int) *Basic {
	return &Basic{X: x, Y: y, W: w, H: h}
}

func (c *Basic) Set(x, y float64) {
	c.X = x
	c.Y = y
}

func (c *Basic) Move(dx, dy float64) {
	c.X += dx
	c.Y += dy
}

func (c *Basic) SetViewport(w, h int) {
	c.W = w
	c.H = h
}

func (c *Basic) WorldToScreen(x, y float64) (sx, sy float64) {
	return x - c.X, y - c.Y
}

func (c *Basic) ScreenToWorld(x, y float64) (wx, wy float64) {
	return x + c.X, y + c.Y
}

func (c *Basic) Visible(x, y float64, w, h int) bool {
	if w <= 0 || h <= 0 || c.W <= 0 || c.H <= 0 {
		return false
	}
	left := int(math.Floor(x))
	top := int(math.Floor(y))
	right := left + w
	bottom := top + h
	cx := int(math.Floor(c.X))
	cy := int(math.Floor(c.Y))

	if right <= cx || bottom <= cy {
		return false
	}
	if left >= cx+c.W || top >= cy+c.H {
		return false
	}
	return true
}
