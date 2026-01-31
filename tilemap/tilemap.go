package tilemap

import (
	"math"

	"github.com/dgrundel/glif/camera"
	"github.com/dgrundel/glif/render"
)

type Map struct {
	W, H    int
	TileW   int
	TileH   int
	Empty   int
	Tiles   []int
	Tileset map[int]*render.Sprite
}

func New(w, h, tileW, tileH, empty int) *Map {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	if tileW <= 0 {
		tileW = 1
	}
	if tileH <= 0 {
		tileH = 1
	}
	return &Map{
		W:       w,
		H:       h,
		TileW:   tileW,
		TileH:   tileH,
		Empty:   empty,
		Tiles:   make([]int, w*h),
		Tileset: map[int]*render.Sprite{},
	}
}

func (m *Map) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < m.W && y < m.H
}

func (m *Map) Set(x, y, id int) {
	if !m.InBounds(x, y) {
		return
	}
	m.Tiles[y*m.W+x] = id
}

func (m *Map) At(x, y int) int {
	if !m.InBounds(x, y) {
		return m.Empty
	}
	return m.Tiles[y*m.W+x]
}

func (m *Map) Draw(r *render.Renderer, worldX, worldY float64, cam *camera.Camera) {
	if m == nil || r == nil {
		return
	}
	startX, startY, endX, endY := m.visibleBounds(worldX, worldY, cam)
	for ty := startY; ty <= endY; ty++ {
		for tx := startX; tx <= endX; tx++ {
			id := m.At(tx, ty)
			if id == m.Empty {
				continue
			}
			sprite := m.Tileset[id]
			if sprite == nil {
				continue
			}
			wx := worldX + float64(tx*m.TileW)
			wy := worldY + float64(ty*m.TileH)
			if cam != nil {
				wx, wy = cam.WorldToScreen(wx, wy)
			}
			r.DrawSprite(int(math.Floor(wx)), int(math.Floor(wy)), sprite)
		}
	}
}

func (m *Map) visibleBounds(worldX, worldY float64, cam *camera.Camera) (int, int, int, int) {
	if m.W == 0 || m.H == 0 {
		return 0, 0, -1, -1
	}
	if cam == nil || cam.ViewW <= 0 || cam.ViewH <= 0 {
		return 0, 0, m.W - 1, m.H - 1
	}
	vx0 := cam.X
	vy0 := cam.Y
	vx1 := cam.X + float64(cam.ViewW)
	vy1 := cam.Y + float64(cam.ViewH)

	startX := int(math.Floor((vx0 - worldX) / float64(m.TileW)))
	startY := int(math.Floor((vy0 - worldY) / float64(m.TileH)))
	endX := int(math.Ceil((vx1-worldX)/float64(m.TileW))) - 1
	endY := int(math.Ceil((vy1-worldY)/float64(m.TileH))) - 1

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX >= m.W {
		endX = m.W - 1
	}
	if endY >= m.H {
		endY = m.H - 1
	}
	if endX < startX || endY < startY {
		return 0, 0, -1, -1
	}
	return startX, startY, endX, endY
}
