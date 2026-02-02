package collision

import "github.com/dgrundel/glif/render"

// Overlaps returns true when two sprites' collision masks overlap in world space.
// If either sprite has no collision mask, Overlaps returns false.
func Overlaps(ax, ay int, a *render.Sprite, bx, by int, b *render.Sprite) bool {
	if a == nil || b == nil || a.Collision == nil || b.Collision == nil {
		return false
	}

	left := max(ax, bx)
	top := max(ay, by)
	right := min(ax+a.Collision.W, bx+b.Collision.W)
	bottom := min(ay+a.Collision.H, by+b.Collision.H)
	if right <= left || bottom <= top {
		return false
	}

	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			if a.Collision.At(x-ax, y-ay) && b.Collision.At(x-bx, y-by) {
				return true
			}
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
