package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
	"github.com/dgrundel/glif/spriteio"
	"github.com/gdamore/tcell/v3"
)

type Point struct {
	X int
	Y int
}

type statusKind int

const (
	statusNone statusKind = iota
	statusInfo
	statusWarn
	statusError
)

type editorMode int

const (
	modeSprite editorMode = iota
	modeWidth
	modeColor
	modeCollision
	modePreview
)

type Editor struct {
	path              string
	cells             map[Point]rune
	widthPath         string
	widthCells        map[Point]rune
	colorPath         string
	colorCells        map[Point]rune
	colorDefault      rune
	collisionPath     string
	collisionCells    map[Point]rune
	collisionDefault  rune
	pal               *palette.Palette
	mode              editorMode
	cursorX           int
	cursorY           int
	quit              bool
	status            string
	statusKind        statusKind
	pendingQuit       bool
	state             input.State
	actions           input.ActionState
	actionMap         input.ActionMap
	moveAccumX        float64
	moveAccumY        float64
	holdLeft          float64
	holdRight         float64
	holdUp            float64
	holdDown          float64
	barStyle          grid.Style
	spriteStyle       grid.Style
	areaStyle         grid.Style
	cursorStyle       grid.Style
	collisionStyle    grid.Style
	previewErrorStyle grid.Style
}

func NewEditor(path string, cells map[Point]rune, widthCells map[Point]rune, colorCells map[Point]rune, colorDefault rune, collisionCells map[Point]rune, collisionDefault rune, pal *palette.Palette, palErr string) *Editor {
	if cells == nil {
		cells = map[Point]rune{}
	}
	if widthCells == nil {
		widthCells = map[Point]rune{}
	}
	if colorCells == nil {
		colorCells = map[Point]rune{}
	}
	if collisionCells == nil {
		collisionCells = map[Point]rune{}
	}
	editor := &Editor{
		path:              path,
		cells:             cells,
		widthPath:         spriteWidthPath(path),
		widthCells:        widthCells,
		colorPath:         spriteColorPath(path),
		colorCells:        colorCells,
		colorDefault:      colorDefault,
		collisionPath:     spriteCollisionPath(path),
		collisionCells:    collisionCells,
		collisionDefault:  collisionDefault,
		pal:               pal,
		mode:              modeSprite,
		actionMap:         defaultActions(),
		barStyle:          grid.Style{Fg: grid.TCellColor(tcell.ColorBlack), Bg: grid.TCellColor(tcell.ColorWhite)},
		spriteStyle:       grid.Style{Fg: grid.TCellColor(tcell.ColorWhite), Bg: grid.TCellColor(tcell.ColorDarkGray)},
		areaStyle:         grid.Style{Fg: grid.TCellColor(tcell.ColorReset), Bg: grid.TCellColor(tcell.ColorLightGreen)},
		cursorStyle:       grid.Style{Fg: grid.TCellColor(tcell.ColorBlack), Bg: grid.TCellColor(tcell.ColorWhite)},
		collisionStyle:    grid.Style{Fg: grid.TCellColor(tcell.ColorWhite), Bg: grid.TCellColor(tcell.ColorRed)},
		previewErrorStyle: grid.Style{Fg: grid.TCellColor(tcell.ColorWhite), Bg: grid.TCellColor(tcell.ColorRed)},
	}
	if palErr != "" {
		editor.status = fmt.Sprintf("palette: %s", palErr)
		editor.statusKind = statusWarn
	}
	return editor
}

func defaultActions() input.ActionMap {
	return input.ActionMap{
		"left":      "key:left",
		"right":     "key:right",
		"up":        "key:up",
		"down":      "key:down",
		"newline":   "key:enter",
		"backspace": "key:backspace",
		"delete":    "key:delete",
		"save":      "key:ctrl+s",
		"toggle":    "key:ctrl+t",
		"trim":      "key:ctrl+k",
		"quit":      "key:ctrl+q",
		"quit_alt":  "key:esc",
	}
}

func (e *Editor) ActionMap() input.ActionMap {
	return e.actionMap
}

func (e *Editor) UpdateActionState(state input.ActionState) {
	e.actions = state
}

func (e *Editor) SetInput(state input.State) {
	e.state = state
}

func (e *Editor) ShouldQuit() bool {
	return e.quit
}

func (e *Editor) Update(dt float64) {
	quitPressed := e.pressed("quit") || e.pressed("quit_alt")
	if quitPressed {
		if e.pendingQuit {
			e.quit = true
			return
		}
		e.pendingQuit = true
		e.status = "Press Ctrl+Q or Esc again to quit"
		e.statusKind = statusWarn
		return
	}

	if e.pressed("save") {
		if e.mode == modeWidth {
			spriteW, spriteH := boundsSize(e.cells)
			if err := writeWidthMask(e.widthPath, e.widthCells, spriteW, spriteH); err != nil {
				e.status = fmt.Sprintf("save failed: %v", err)
				e.statusKind = statusError
			} else {
				e.ensureWidthCells()
				e.status = "saved"
				e.statusKind = statusInfo
			}
		} else if e.mode == modeColor {
			spriteW, spriteH := boundsSize(e.cells)
			if err := writeColorMask(e.colorPath, e.colorCells, spriteW, spriteH, e.colorDefault); err != nil {
				e.status = fmt.Sprintf("save failed: %v", err)
				e.statusKind = statusError
			} else {
				e.ensureColorCells()
				e.status = "saved"
				e.statusKind = statusInfo
			}
		} else if e.mode == modeCollision {
			spriteW, spriteH := boundsSize(e.cells)
			if err := writeCollisionMask(e.collisionPath, e.collisionCells, spriteW, spriteH, e.collisionDefault); err != nil {
				e.status = fmt.Sprintf("save failed: %v", err)
				e.statusKind = statusError
			} else {
				e.ensureCollisionCells()
				e.status = "saved"
				e.statusKind = statusInfo
			}
		} else if err := writeSprite(e.path, e.cells); err != nil {
			e.status = fmt.Sprintf("save failed: %v", err)
			e.statusKind = statusError
		} else {
			e.cells = normalizeCells(e.cells)
			e.status = "saved"
			e.statusKind = statusInfo
		}
	}
	if e.pressed("trim") {
		e.cells = trimOuterWhitespace(e.cells)
		e.status = "trimmed"
		e.statusKind = statusInfo
	}
	if e.pressed("toggle") {
		e.toggleMode()
	}

	pressedLeft := e.pressed("left")
	pressedRight := e.pressed("right")
	pressedUp := e.pressed("up")
	pressedDown := e.pressed("down")

	if pressedLeft {
		if e.cursorX > 0 {
			e.cursorX--
		}
	}
	if pressedRight {
		e.cursorX++
	}
	if pressedUp {
		if e.cursorY > 0 {
			e.cursorY--
		}
	}
	if pressedDown {
		e.cursorY++
	}

	if e.pressed("newline") {
		e.cursorY++
		e.cursorX = 0
	}

	if e.pressed("backspace") {
		if e.cursorX > 0 {
			e.cursorX--
			e.clearActiveCell(e.cursorX, e.cursorY)
		}
	}
	if e.pressed("delete") {
		e.clearActiveCell(e.cursorX, e.cursorY)
	}

	for _, r := range e.state.Typed {
		e.setActiveCell(e.cursorX, e.cursorY, r)
		e.cursorX++
	}

	e.handleHeldMovement(dt, pressedLeft || pressedRight, pressedUp || pressedDown)

	if e.pendingQuit && e.anyNonQuitPress() {
		e.pendingQuit = false
		if e.statusKind == statusWarn {
			e.status = ""
			e.statusKind = statusNone
		}
	}

	if e.cursorX < 0 {
		e.cursorX = 0
	}
	if e.cursorY < 0 {
		e.cursorY = 0
	}
}

func (e *Editor) Draw(r *render.Renderer) {
	frameW := r.Frame.W
	frameH := r.Frame.H
	if frameW <= 0 || frameH <= 0 {
		return
	}

	modeLabel := "Sprite"
	if e.mode == modeWidth {
		modeLabel = "Width"
	}
	if e.mode == modeColor {
		modeLabel = "Color"
	}
	if e.mode == modeCollision {
		modeLabel = "Collision"
	}
	if e.mode == modePreview {
		modeLabel = "Preview"
	}
	pathText := truncateToWidth(fmt.Sprintf("Mode (Ctrl+T): %s | %s", modeLabel, e.path), frameW)
	r.Rect(0, 0, frameW, 1, e.barStyle, render.RectOptions{Fill: true, FillRune: ' '})
	r.DrawText(0, 0, pathText, e.barStyle)

	w, h := e.modeBounds()
	status := fmt.Sprintf("%dx%d", w, h)
	if e.status != "" {
		status = fmt.Sprintf("%s | %s", status, e.status)
	}
	status = truncateToWidth(status, frameW)
	statusStyle := e.barStyle
	if e.statusKind == statusWarn || e.statusKind == statusError {
		statusStyle = grid.Style{Fg: grid.TCellColor(tcell.ColorWhite), Bg: grid.TCellColor(tcell.ColorRed)}
	}
	r.Rect(0, frameH-1, frameW, 1, statusStyle, render.RectOptions{Fill: true, FillRune: ' '})
	r.DrawText(0, frameH-1, status, statusStyle)

	areaW := w
	areaH := h
	if areaW < 1 {
		areaW = 1
	}
	if areaH < 1 {
		areaH = 1
	}
	maxAreaW := frameW
	maxAreaH := frameH - 2
	if maxAreaH < 0 {
		maxAreaH = 0
	}
	if areaW > maxAreaW {
		areaW = maxAreaW
	}
	if areaH > maxAreaH {
		areaH = maxAreaH
	}

	activeCells := e.activeCells()
	if e.mode == modePreview {
		sprite := e.previewSprite()
		if sprite != nil {
			r.DrawSprite(0, 1, sprite)
		}
	} else {
		for y := 0; y < areaH; y++ {
			drawY := 1 + y
			if drawY >= frameH-1 {
				break
			}
			for x := 0; x < areaW && x < frameW; x++ {
				if e.mode == modeColor {
					key, ok := activeCells[Point{X: x, Y: y}]
					if !ok {
						key = e.colorDefault
					}
					style := e.colorStyleFor(key, r.Frame.At(x, drawY).Style)
					r.Frame.Set(x, drawY, grid.Cell{Ch: key, Style: style})
					continue
				}
				if e.mode == modeCollision {
					key, ok := activeCells[Point{X: x, Y: y}]
					if !ok {
						key = e.collisionDefault
					}
					style := e.areaStyle.Resolve(r.Frame.At(x, drawY).Style)
					if isCollisionKey(key) {
						style = e.collisionStyle.Resolve(r.Frame.At(x, drawY).Style)
					}
					r.Frame.Set(x, drawY, grid.Cell{Ch: key, Style: style})
					continue
				}
				ch, ok := activeCells[Point{X: x, Y: y}]
				if !ok {
					r.Frame.Set(x, drawY, grid.Cell{Ch: ' ', Style: e.areaStyle.Resolve(r.Frame.At(x, drawY).Style)})
					continue
				}
				r.Frame.Set(x, drawY, grid.Cell{Ch: ch, Style: e.spriteStyle.Resolve(r.Frame.At(x, drawY).Style)})
			}
		}
	}

	drawX := e.cursorX
	drawY := 1 + e.cursorY
	if drawX >= 0 && drawX < frameW && drawY >= 1 && drawY < frameH-1 {
		ch, ok := activeCells[Point{X: e.cursorX, Y: e.cursorY}]
		if !ok {
			if e.mode == modeColor {
				ch = e.colorDefault
			} else if e.mode == modeCollision {
				ch = e.collisionDefault
			} else if e.mode == modePreview {
				ch = ' '
			} else {
				ch = ' '
			}
		}
		base := r.Frame.At(drawX, drawY).Style
		r.Frame.Set(drawX, drawY, grid.Cell{Ch: ch, Style: e.cursorStyle.Resolve(base)})
	}
}

func (e *Editor) Resize(w, h int) {
}

func (e *Editor) pressed(action input.Action) bool {
	if e.actions.Pressed == nil {
		return false
	}
	return e.actions.Pressed[action]
}

func (e *Editor) held(action input.Action) bool {
	if e.actions.Held == nil {
		return false
	}
	return e.actions.Held[action]
}

func (e *Editor) anyNonQuitPress() bool {
	if e.actions.Pressed != nil {
		for action, pressed := range e.actions.Pressed {
			if !pressed {
				continue
			}
			if action == "quit" || action == "quit_alt" {
				continue
			}
			return true
		}
	}
	for key := range e.state.Pressed {
		if string(key) == "key:ctrl+q" || string(key) == "key:esc" {
			continue
		}
		return true
	}
	return false
}

func (e *Editor) handleHeldMovement(dt float64, pressedX, pressedY bool) {
	const moveRate = 18.0
	const repeatDelay = 0.25

	if pressedX {
		e.holdLeft = 0
		e.holdRight = 0
	}
	if pressedY {
		e.holdUp = 0
		e.holdDown = 0
	}

	if e.held("left") {
		e.holdLeft += dt
	} else {
		e.holdLeft = 0
	}
	if e.held("right") {
		e.holdRight += dt
	} else {
		e.holdRight = 0
	}
	if e.held("up") {
		e.holdUp += dt
	} else {
		e.holdUp = 0
	}
	if e.held("down") {
		e.holdDown += dt
	} else {
		e.holdDown = 0
	}

	moveX := 0
	moveY := 0
	if e.holdLeft >= repeatDelay {
		moveX--
	}
	if e.holdRight >= repeatDelay {
		moveX++
	}
	if e.holdUp >= repeatDelay {
		moveY--
	}
	if e.holdDown >= repeatDelay {
		moveY++
	}

	if moveX == 0 {
		e.moveAccumX = 0
	} else {
		e.moveAccumX += float64(moveX) * moveRate * dt
	}
	if moveY == 0 {
		e.moveAccumY = 0
	} else {
		e.moveAccumY += float64(moveY) * moveRate * dt
	}

	for e.moveAccumX >= 1 {
		e.cursorX++
		e.moveAccumX -= 1
	}
	for e.moveAccumX <= -1 {
		if e.cursorX > 0 {
			e.cursorX--
		}
		e.moveAccumX += 1
	}
	for e.moveAccumY >= 1 {
		e.cursorY++
		e.moveAccumY -= 1
	}
	for e.moveAccumY <= -1 {
		if e.cursorY > 0 {
			e.cursorY--
		}
		e.moveAccumY += 1
	}
}

func (e *Editor) activeCells() map[Point]rune {
	if e.mode == modeWidth {
		return e.widthCells
	}
	if e.mode == modeColor {
		return e.colorCells
	}
	if e.mode == modeCollision {
		return e.collisionCells
	}
	if e.mode == modePreview {
		return e.cells
	}
	return e.cells
}

func (e *Editor) setActiveCell(x, y int, ch rune) {
	if x < 0 || y < 0 {
		return
	}
	if e.mode == modeWidth {
		e.widthCells[Point{X: x, Y: y}] = ch
		return
	}
	if e.mode == modeColor {
		e.colorCells[Point{X: x, Y: y}] = ch
		return
	}
	if e.mode == modeCollision {
		e.collisionCells[Point{X: x, Y: y}] = ch
		return
	}
	e.cells[Point{X: x, Y: y}] = ch
}

func (e *Editor) clearActiveCell(x, y int) {
	if x < 0 || y < 0 {
		return
	}
	if e.mode == modeWidth {
		e.widthCells[Point{X: x, Y: y}] = '1'
		return
	}
	if e.mode == modeColor {
		e.colorCells[Point{X: x, Y: y}] = e.colorDefault
		return
	}
	if e.mode == modeCollision {
		e.collisionCells[Point{X: x, Y: y}] = e.collisionDefault
		return
	}
	delete(e.cells, Point{X: x, Y: y})
}

func (e *Editor) toggleMode() {
	if e.mode == modeWidth {
		e.mode = modeColor
		e.ensureColorCells()
		e.status = "mode: color"
		e.statusKind = statusInfo
		return
	}
	if e.mode == modeColor {
		e.mode = modeCollision
		e.ensureCollisionCells()
		e.status = "mode: collision"
		e.statusKind = statusInfo
		return
	}
	if e.mode == modeCollision {
		e.mode = modePreview
		e.status = "mode: preview"
		e.statusKind = statusInfo
		return
	}
	if e.mode == modePreview {
		e.mode = modeSprite
		e.status = "mode: sprite"
		e.statusKind = statusInfo
		return
	}
	e.mode = modeWidth
	e.ensureWidthCells()
	e.status = "mode: width"
	e.statusKind = statusInfo
}

func (e *Editor) modeBounds() (int, int) {
	if e.mode == modeWidth {
		w, h := boundsSize(e.cells)
		if w > 0 && h > 0 {
			return w, h
		}
		return boundsSize(e.widthCells)
	}
	if e.mode == modeColor {
		w, h := boundsSize(e.cells)
		if w > 0 && h > 0 {
			return w, h
		}
		return boundsSize(e.colorCells)
	}
	if e.mode == modeCollision {
		w, h := boundsSize(e.cells)
		if w > 0 && h > 0 {
			return w, h
		}
		return boundsSize(e.collisionCells)
	}
	if e.mode == modePreview {
		return boundsSize(e.cells)
	}
	return boundsSize(e.cells)
}

func boundsSize(cells map[Point]rune) (int, int) {
	maxX := -1
	maxY := -1
	for p := range cells {
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if maxX < 0 || maxY < 0 {
		return 0, 0
	}
	return maxX + 1, maxY + 1
}

func trimOuterWhitespace(cells map[Point]rune) map[Point]rune {
	minX := -1
	maxX := -1
	minY := -1
	maxY := -1
	for p, ch := range cells {
		if ch == ' ' {
			continue
		}
		if minX == -1 || p.X < minX {
			minX = p.X
		}
		if maxX == -1 || p.X > maxX {
			maxX = p.X
		}
		if minY == -1 || p.Y < minY {
			minY = p.Y
		}
		if maxY == -1 || p.Y > maxY {
			maxY = p.Y
		}
	}
	if minX == -1 || maxX == -1 || minY == -1 || maxY == -1 {
		return map[Point]rune{}
	}
	out := make(map[Point]rune, len(cells))
	for p, ch := range cells {
		if p.X < minX || p.X > maxX || p.Y < minY || p.Y > maxY {
			continue
		}
		out[Point{X: p.X - minX, Y: p.Y - minY}] = ch
	}
	return out
}

func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == width {
			return s[:i]
		}
		count++
	}
	return s
}

func readSprite(path string) (map[Point]rune, error) {
	f, err := spriteio.LoadFile(path)
	if err != nil {
		return nil, err
	}
	rows := f.RuneRows()
	if len(rows) == 0 {
		return map[Point]rune{}, nil
	}
	out := make(map[Point]rune)
	for y, line := range rows {
		for x, ch := range line {
			out[Point{X: x, Y: y}] = ch
		}
	}
	return out, nil
}

func spriteWidthPath(spritePath string) string {
	return strings.TrimSuffix(spritePath, ".sprite") + ".width"
}

func spriteColorPath(spritePath string) string {
	return strings.TrimSuffix(spritePath, ".sprite") + ".color"
}

func spriteCollisionPath(spritePath string) string {
	return strings.TrimSuffix(spritePath, ".sprite") + ".collision"
}

func writeSprite(path string, cells map[Point]rune) error {
	w, h := boundsSize(cells)
	if w == 0 || h == 0 {
		return os.WriteFile(path, []byte(""), 0o644)
	}
	lines := make([][]rune, h)
	for y := 0; y < h; y++ {
		line := make([]rune, w)
		for x := 0; x < w; x++ {
			line[x] = ' '
		}
		lines[y] = line
	}
	for p, ch := range cells {
		if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
			continue
		}
		lines[p.Y][p.X] = ch
	}
	parts := make([]string, h)
	for y := 0; y < h; y++ {
		parts[y] = string(lines[y])
	}
	out := strings.Join(parts, "\n")
	return os.WriteFile(path, []byte(out), 0o644)
}

func normalizeCells(cells map[Point]rune) map[Point]rune {
	w, h := boundsSize(cells)
	if w == 0 || h == 0 {
		return map[Point]rune{}
	}
	out := make(map[Point]rune, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out[Point{X: x, Y: y}] = ' '
		}
	}
	for p, ch := range cells {
		if p.X < 0 || p.Y < 0 || p.X >= w || p.Y >= h {
			continue
		}
		out[p] = ch
	}
	return out
}

func generateFilledCells(w, h int, fill rune) map[Point]rune {
	if w <= 0 || h <= 0 {
		return map[Point]rune{}
	}
	out := make(map[Point]rune, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out[Point{X: x, Y: y}] = fill
		}
	}
	return out
}

func resolvePalettePath(basePath string) string {
	candidate := basePath + ".palette"
	if fileExists(candidate) {
		return candidate
	}
	dir := filepath.Dir(basePath)
	for {
		candidate = filepath.Join(dir, "default.palette")
		if fileExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return candidate
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func firstPaletteKey(path string) (rune, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return 0, false, nil
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		keyRunes := []rune(fields[0])
		if len(keyRunes) != 1 {
			return 0, false, fmt.Errorf("palette key must be single rune on line %d", i+1)
		}
		return keyRunes[0], true, nil
	}
	return 0, false, nil
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: spriteeditor path/to/file.sprite")
		os.Exit(2)
	}

	path := args[0]
	if !strings.HasSuffix(path, ".sprite") {
		fmt.Fprintln(os.Stderr, "path must end with .sprite")
		os.Exit(2)
	}

	if dir := filepath.Dir(path); dir != "." {
		if _, err := os.Stat(dir); err != nil {
			fmt.Fprintf(os.Stderr, "invalid directory: %v\n", err)
			os.Exit(1)
		}
	}

	cells := map[Point]rune{}
	if _, err := os.Stat(path); err == nil {
		loaded, err := readSprite(path)
		if err != nil {
			log.Fatal(err)
		}
		cells = loaded
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	widthCells, _, err := readWidthMask(spriteWidthPath(path))
	if err != nil {
		log.Fatal(err)
	}

	spriteW, spriteH := boundsSize(cells)
	palPath := resolvePalettePath(strings.TrimSuffix(path, ".sprite"))
	colorDefault := rune('a')
	if key, ok, err := firstPaletteKey(palPath); err != nil {
		log.Fatal(err)
	} else if ok {
		colorDefault = key
	}
	var pal *palette.Palette
	palErr := ""
	if loaded, err := palette.Load(palPath); err != nil {
		palErr = err.Error()
	} else {
		pal = loaded
	}
	colorCells, err := readColorMask(spriteColorPath(path), spriteW, spriteH, colorDefault)
	if err != nil {
		log.Fatal(err)
	}

	collisionDefault := rune('.')
	collisionCells, err := readCollisionMask(spriteCollisionPath(path), spriteW, spriteH, collisionDefault)
	if err != nil {
		log.Fatal(err)
	}

	game := NewEditor(path, cells, widthCells, colorCells, colorDefault, collisionCells, collisionDefault, pal, palErr)
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: grid.Style{Fg: grid.TCellColor(tcell.ColorReset), Bg: grid.TCellColor(tcell.ColorReset)}}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
