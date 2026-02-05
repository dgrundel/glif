# Expanded Camera APIs (World Demo Simplification)

## Summary
Provide higher‑level camera helpers so demos don’t manipulate `camera.Basic` internals directly. The goal is to move camera math (clamping, panning, viewport bounds) into the camera package so game/demo code reads as “intent” rather than implementation detail.

## Problem in World Demo
The world demo currently:
- Mutates camera fields directly (`cam.X`, `cam.Y`).
- Implements clamp logic using tilemap dimensions.
- Converts between view size and world size itself.

This couples the demo to camera internals and duplicates logic that will likely be used in other demos.

## Goals
- Keep game/demo code declarative: “pan left”, “clamp to map”, “set viewport”.
- Preserve the game‑owned camera model.
- Avoid introducing a heavyweight scene system.

## Proposed API (exhaustive)

### camera package
```go
// Bounds represents a world-space rectangle.
type Bounds struct {
	X float64
	Y float64
	W float64
	H float64
}

// Camera is unchanged, but we add helpers around it.
type Camera interface {
	WorldToScreen(x, y float64) (sx, sy float64)
	ScreenToWorld(x, y float64) (wx, wy float64)
	Visible(x, y float64, w, h int) bool
	SetViewport(w, h int)
}

// Basic remains the concrete implementation.
type Basic struct {
	X float64
	Y float64
	W int
	H int
}

func NewBasic(x, y float64, w, h int) *Basic
func (c *Basic) Set(x, y float64)
func (c *Basic) Move(dx, dy float64)
func (c *Basic) SetViewport(w, h int)
func (c *Basic) WorldToScreen(x, y float64) (sx, sy float64)
func (c *Basic) ScreenToWorld(x, y float64) (wx, wy float64)
func (c *Basic) Visible(x, y float64, w, h int) bool

// New helpers:

// SetCenter positions the camera so that (cx, cy) is the center of the viewport.
func (c *Basic) SetCenter(cx, cy float64)

// Center returns the current center of the camera viewport.
func (c *Basic) Center() (cx, cy float64)

// ClampTo bounds the camera to a world rect (e.g., tilemap extents).
// If the world is smaller than the viewport, the camera is pinned to 0,0.
func (c *Basic) ClampTo(b Bounds)

// Pan applies a directional pan with speed (units/sec) and dt.
func (c *Basic) Pan(dx, dy, speed, dt float64)
```

### tilemap package (optional helper)
```go
// WorldBounds returns the world-space bounds of the map in tile units.
func (m *Map) WorldBounds() camera.Bounds
```

## How This Simplifies the World Demo
Current behavior:
- pan with arrow keys
- clamp to tilemap bounds
- set viewport size on resize

With the new API, demo code becomes:
```go
func (d *Demo) Update(dt float64) {
	d.cam.Pan(panX, panY, 12, dt)
	d.cam.ClampTo(d.tile.WorldBounds())
}

func (d *Demo) Resize(w, h int) {
	d.cam.SetViewport(w, h)
}
```

The demo no longer touches camera fields or computes clamp values.

## Alternatives Considered
1. **Keep manual camera math in demos**: simplest but repetitive and error‑prone.
2. **Add a full camera controller system**: more power but heavier than needed right now.
3. **Tie camera to renderer/engine**: breaks the game‑owned camera model you prefer.

## Risks / Tradeoffs
- Adds more API surface to `camera`.
- Some helpers (like `Pan`) encode movement conventions (units/sec) which should be documented.

## Open Questions
- Should `ClampTo` pin to `(0,0)` or center the camera when the world is smaller than the viewport?
- Should `Pan` accept a normalized direction or raw dx/dy?
- Do we want a separate `CameraController` type (e.g., for smoothing/follow), or keep these on `Basic`?
