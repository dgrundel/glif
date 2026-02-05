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
	x float64
	y float64
	w int
	h int
}

func NewBasic(x, y float64, w, h int) *Basic {
	return &Basic{x: x, y: y, w: w, h: h}
}

func (c *Basic) Set(x, y float64) {
	c.x = x
	c.y = y
}

func (c *Basic) SetCenter(cx, cy float64) {
	c.x = cx - float64(c.w)/2
	c.y = cy - float64(c.h)/2
}

func (c *Basic) Center() (float64, float64) {
	return c.x + float64(c.w)/2, c.y + float64(c.h)/2
}

func (c *Basic) Move(dx, dy float64) {
	c.x += dx
	c.y += dy
}

func (c *Basic) Pan(dx, dy, speed, dt float64) {
	c.x += dx * speed * dt
	c.y += dy * speed * dt
}

func (c *Basic) SetViewport(w, h int) {
	c.w = w
	c.h = h
}

func (c *Basic) WorldToScreen(x, y float64) (sx, sy float64) {
	return x - c.x, y - c.y
}

func (c *Basic) ScreenToWorld(x, y float64) (wx, wy float64) {
	return x + c.x, y + c.y
}

func (c *Basic) Visible(x, y float64, w, h int) bool {
	if w <= 0 || h <= 0 || c.w <= 0 || c.h <= 0 {
		return false
	}
	left := int(math.Floor(x))
	top := int(math.Floor(y))
	right := left + w
	bottom := top + h
	cx := int(math.Floor(c.x))
	cy := int(math.Floor(c.y))

	if right <= cx || bottom <= cy {
		return false
	}
	if left >= cx+c.w || top >= cy+c.h {
		return false
	}
	return true
}

type Bounds struct {
	X float64
	Y float64
	W float64
	H float64
}

func (c *Basic) ClampTo(b Bounds) {
	if c == nil {
		return
	}
	if b.W <= 0 || b.H <= 0 || c.w <= 0 || c.h <= 0 {
		return
	}
	maxX := b.X + b.W - float64(c.w)
	maxY := b.Y + b.H - float64(c.h)
	if maxX < b.X {
		maxX = b.X
	}
	if maxY < b.Y {
		maxY = b.Y
	}
	if c.x < b.X {
		c.x = b.X
	}
	if c.y < b.Y {
		c.y = b.Y
	}
	if c.x > maxX {
		c.x = maxX
	}
	if c.y > maxY {
		c.y = maxY
	}
	if b.W < float64(c.w) {
		c.x = b.X + (b.W-float64(c.w))/2
	}
	if b.H < float64(c.h) {
		c.y = b.Y + (b.H-float64(c.h))/2
	}
}

func (c *Basic) Position() (float64, float64) {
	return c.x, c.y
}

func (c *Basic) Viewport() (int, int) {
	return c.w, c.h
}
