package grid

type Vec2i struct {
	X int
	Y int
}

type Rect struct {
	X int
	Y int
	W int
	H int
}

func (r Rect) Contains(p Vec2i) bool {
	return p.X >= r.X && p.Y >= r.Y && p.X < r.X+r.W && p.Y < r.Y+r.H
}

func (r Rect) Empty() bool {
	return r.W <= 0 || r.H <= 0
}
