package camera

// Camera projects world-space coordinates into screen space.
type Camera interface {
	WorldToScreen(x, y int) (sx, sy int)
	ScreenToWorld(x, y int) (wx, wy int)
	Visible(x, y, w, h int) bool
	SetViewport(w, h int)
}

// Basic is a simple top-left anchored camera.
// X,Y are the world-space coordinates of the top-left screen cell.
type Basic struct {
	X int
	Y int
	W int
	H int
}

func NewBasic(x, y, w, h int) *Basic {
	return &Basic{X: x, Y: y, W: w, H: h}
}

func (c *Basic) Set(x, y int) {
	c.X = x
	c.Y = y
}

func (c *Basic) Move(dx, dy int) {
	c.X += dx
	c.Y += dy
}

func (c *Basic) SetViewport(w, h int) {
	c.W = w
	c.H = h
}

func (c *Basic) WorldToScreen(x, y int) (sx, sy int) {
	return x - c.X, y - c.Y
}

func (c *Basic) ScreenToWorld(x, y int) (wx, wy int) {
	return x + c.X, y + c.Y
}

func (c *Basic) Visible(x, y, w, h int) bool {
	if w <= 0 || h <= 0 || c.W <= 0 || c.H <= 0 {
		return false
	}
	left := x
	top := y
	right := x + w
	bottom := y + h

	if right <= c.X || bottom <= c.Y {
		return false
	}
	if left >= c.X+c.W || top >= c.Y+c.H {
		return false
	}
	return true
}
