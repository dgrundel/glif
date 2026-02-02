package render

type CollisionMask struct {
	W     int
	H     int
	Cells []bool
}

func (m *CollisionMask) At(x, y int) bool {
	if m == nil || x < 0 || y < 0 || x >= m.W || y >= m.H {
		return false
	}
	return m.Cells[y*m.W+x]
}
