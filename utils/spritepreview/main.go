package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
)

type PreviewItem struct {
	Name       string
	LabelWidth int
	Sprite     *render.Sprite
}

type SpritePreview struct {
	items      []PreviewItem
	state      input.State
	binds      input.ActionMap
	quit       bool
	panY       float64
	maxPan     float64
	width      int
	height     int
	border     grid.Style
	text       grid.Style
	background grid.Style
}

type layoutItem struct {
	item PreviewItem
	x    int
	y    int
	w    int
	h    int
}

func NewSpritePreview(items []PreviewItem) *SpritePreview {
	return &SpritePreview{
		items: items,
		binds: input.ActionMap{
			"pan_up":   "key:up",
			"pan_down": "key:down",
			"pan_up_w": "w",
			"pan_dn_s": "s",
			"quit":     "key:esc",
			"quit_alt": "key:ctrl+c",
		},
		border:     grid.Style{Fg: tcell.ColorWhite, Bg: tcell.ColorReset},
		text:       grid.Style{Fg: tcell.ColorWhite, Bg: tcell.ColorReset},
		background: grid.Style{Fg: tcell.ColorReset, Bg: tcell.ColorReset},
	}
}

func (p *SpritePreview) Update(dt float64) {
	if p.pressed("quit") || p.pressed("quit_alt") {
		p.quit = true
		return
	}

	layout := p.computeLayout(p.width)
	if layout.totalH < p.height {
		p.maxPan = 0
	} else {
		p.maxPan = float64(layout.totalH - p.height)
	}

	const panSpeed = 20.0
	delta := 0.0
	if p.pressed("pan_up") || p.pressed("pan_up_w") {
		delta -= 1
	}
	if p.pressed("pan_down") || p.pressed("pan_dn_s") {
		delta += 1
	}
	if p.held("pan_up") || p.held("pan_up_w") {
		delta -= panSpeed * dt
	}
	if p.held("pan_down") || p.held("pan_dn_s") {
		delta += panSpeed * dt
	}
	if delta == 0 {
		return
	}
	p.panY += delta
	if p.panY < 0 {
		p.panY = 0
	}
	if p.panY > p.maxPan {
		p.panY = p.maxPan
	}
}

func (p *SpritePreview) Draw(r *render.Renderer) {
	layout := p.computeLayout(r.Frame.W)
	p.maxPan = 0
	if layout.totalH > r.Frame.H {
		p.maxPan = float64(layout.totalH - r.Frame.H)
	}
	if p.panY > p.maxPan {
		p.panY = p.maxPan
	}

	offsetY := int(p.panY)
	for _, li := range layout.items {
		drawX := li.x
		drawY := li.y - offsetY
		if drawY+li.h < 0 || drawY >= r.Frame.H {
			continue
		}
		r.Rect(drawX, drawY, li.w, li.h, p.border)
		r.DrawSprite(drawX+1, drawY+1, li.item.Sprite)
		r.DrawText(drawX+1, drawY+1+li.item.Sprite.H, li.item.Name, p.text)
	}
}

func (p *SpritePreview) Resize(w, h int) {
	p.width = w
	p.height = h
}

func (p *SpritePreview) SetInput(state input.State) {
	p.state = state
}

func (p *SpritePreview) ShouldQuit() bool {
	return p.quit
}

func (p *SpritePreview) pressed(action input.Action) bool {
	key, ok := p.binds[action]
	if !ok {
		return false
	}
	return p.state.Pressed[key]
}

func (p *SpritePreview) held(action input.Action) bool {
	key, ok := p.binds[action]
	if !ok {
		return false
	}
	return p.state.Held[key]
}

type layoutResult struct {
	items  []layoutItem
	totalH int
}

func (p *SpritePreview) computeLayout(frameW int) layoutResult {
	const gapX = 2
	const gapY = 1
	x := 0
	y := 0
	rowH := 0

	out := layoutResult{items: make([]layoutItem, 0, len(p.items))}
	for _, item := range p.items {
		boxW := max(item.Sprite.W, item.LabelWidth) + 2
		boxH := item.Sprite.H + 1 + 2

		if x > 0 && x+boxW > frameW {
			x = 0
			y += rowH + gapY
			rowH = 0
		}

		out.items = append(out.items, layoutItem{
			item: item,
			x:    x,
			y:    y,
			w:    boxW,
			h:    boxH,
		})

		if boxH > rowH {
			rowH = boxH
		}
		x += boxW + gapX
	}

	out.totalH = y + rowH
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func loadItem(base, label string) PreviewItem {
	basePath := base
	if strings.HasSuffix(basePath, ".sprite") {
		basePath = strings.TrimSuffix(basePath, ".sprite")
	}
	sprite, err := assets.LoadSprite(basePath)
	if err != nil {
		log.Fatalf("load sprite %s: %v", base, err)
	}
	name := filepath.Base(label)
	return PreviewItem{
		Name:       name,
		LabelWidth: utf8.RuneCountInString(name),
		Sprite:     sprite,
	}
}

func main() {
	recursive := flag.Bool("r", false, "recursively scan folders for .sprite files")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: spritepreview [-r] path/to/sprite_or_folder [more sprites_or_folders...]")
		os.Exit(2)
	}

	items := make([]PreviewItem, 0, len(args))
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			if *recursive {
				paths := make([]string, 0)
				err := filepath.WalkDir(arg, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					if strings.HasSuffix(d.Name(), ".sprite") {
						paths = append(paths, path)
					}
					return nil
				})
				if err != nil {
					log.Fatalf("walk dir %s: %v", arg, err)
				}
				sort.Strings(paths)
				for _, path := range paths {
					base := strings.TrimSuffix(path, ".sprite")
					label := path
					items = append(items, loadItem(base, label))
				}
			} else {
				entries, err := os.ReadDir(arg)
				if err != nil {
					log.Fatalf("read dir %s: %v", arg, err)
				}
				names := make([]string, 0, len(entries))
				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					name := entry.Name()
					if strings.HasSuffix(name, ".sprite") {
						names = append(names, name)
					}
				}
				sort.Strings(names)
				for _, name := range names {
					base := filepath.Join(arg, strings.TrimSuffix(name, ".sprite"))
					items = append(items, loadItem(base, name))
				}
			}
			continue
		}
		items = append(items, loadItem(arg, filepath.Base(arg)))
	}

	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "no sprites found in provided inputs")
		os.Exit(2)
	}

	game := NewSpritePreview(items)
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: game.background}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
