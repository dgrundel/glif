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

func (m *Map) Draw(r *render.Renderer, worldX, worldY float64, cam camera.Camera) {
	if m == nil || r == nil {
		return
	}
	for ty := 0; ty < m.H; ty++ {
		for tx := 0; tx < m.W; tx++ {
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
				if !cam.Visible(wx, wy, m.TileW, m.TileH) {
					continue
				}
				wx, wy = cam.WorldToScreen(wx, wy)
			}
			r.DrawSprite(int(math.Floor(wx)), int(math.Floor(wy)), sprite)
		}
	}
}
