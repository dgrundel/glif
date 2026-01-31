package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/palette"
	"github.com/dgrundel/glif/render"
)

type Phase int

const (
	PhaseSplash Phase = iota
	PhasePlayer
	PhaseDealer
	PhaseOver
)

type Blackjack struct {
	state input.State
	binds input.ActionMap
	quit  bool

	phase        Phase
	menuIndex    int
	endMenuIndex int
	message      string

	deck   []Card
	player []Card
	dealer []Card

	splash *render.Sprite
	cards  map[rune]*render.Sprite
	back   *render.Sprite
	uiText grid.Style
	uiRed  grid.Style
	bg     grid.Style
	width  int
	height int
}

type Card struct {
	Rank string
	Suit rune
}

func NewBlackjack() *Blackjack {
	splash, err := assets.LoadMaskedSprite("demos/blackjack/assets/splash")
	if err != nil {
		log.Fatal(err)
	}
	pal, err := palette.Load("demos/blackjack/assets/ui.palette")
	if err != nil {
		log.Fatal(err)
	}
	uiText, err := pal.Style('w')
	if err != nil {
		log.Fatal(err)
	}
	uiRed, err := pal.Style('r')
	if err != nil {
		log.Fatal(err)
	}
	bg, err := pal.Style('b')
	if err != nil {
		log.Fatal(err)
	}
	b := &Blackjack{
		binds: input.ActionMap{
			"up":       "key:up",
			"down":     "key:down",
			"left":     "key:left",
			"right":    "key:right",
			"enter":    "key:enter",
			"quit":     "key:esc",
			"quit_alt": "key:ctrl+c",
		},
		splash: splash,
		cards:  loadCardSprites(),
		back:   loadSprite("card_back"),
		uiText: uiText,
		uiRed:  uiRed,
		bg:     bg,
	}
	b.resetDeck()
	return b
}

func (b *Blackjack) Update(dt float64) {
	_ = dt
	if b.pressed("quit") || b.pressed("quit_alt") {
		b.quit = true
		return
	}

	switch b.phase {
	case PhaseSplash:
		if b.pressed("enter") {
			b.startGame()
		}
	case PhasePlayer:
		b.updateMenu(&b.menuIndex, 3)
		if b.pressed("enter") {
			switch b.menuIndex {
			case 0:
				b.playerHit()
			case 1:
				b.phase = PhaseDealer
			case 2:
				b.quit = true
			}
		}
	case PhaseDealer:
		b.runDealer()
	case PhaseOver:
		b.updateMenu(&b.endMenuIndex, 2)
		if b.pressed("enter") {
			if b.endMenuIndex == 0 {
				b.startGame()
			} else {
				b.quit = true
			}
		}
	}
}

func (b *Blackjack) Draw(r *render.Renderer) {
	textStyle := b.uiText
	redStyle := b.uiRed

	switch b.phase {
	case PhaseSplash:
		b.drawSplash(r, textStyle)
		return
	}

	playerTotal := handTotal(b.player)
	dealerTotal := handTotal(b.dealer)
	if b.phase == PhasePlayer {
		dealerTotal = handTotal([]Card{b.dealer[0]})
	}

	x0, y0 := b.layoutOffsets(r, dealerTotal, playerTotal)
	r.DrawText(x0, y0, fmt.Sprintf("Dealer hand (%d):", dealerTotal), textStyle)
	drawHand(r, x0, y0+1, b.dealer, b.phase != PhasePlayer, textStyle, redStyle, b.cards, b.back)

	r.DrawText(x0, y0+7, fmt.Sprintf("Your hand (%d):", playerTotal), textStyle)
	drawHand(r, x0, y0+8, b.player, true, textStyle, redStyle, b.cards, b.back)

	switch b.phase {
	case PhasePlayer:
		drawMenu(r, x0, y0+14, []string{"Hit", "Stay", "Quit"}, b.menuIndex, textStyle)
	case PhaseOver:
		r.DrawText(x0, y0+14, b.message, textStyle)
		drawMenu(r, x0, y0+16, []string{"Play again", "Quit"}, b.endMenuIndex, textStyle)
	}
}

func (b *Blackjack) Resize(w, h int) {
	b.width = w
	b.height = h
}

func (b *Blackjack) SetInput(state input.State) {
	b.state = state
}

func (b *Blackjack) ShouldQuit() bool {
	return b.quit
}

func (b *Blackjack) pressed(action input.Action) bool {
	key, ok := b.binds[action]
	if !ok {
		return false
	}
	return b.state.Pressed[key]
}

func (b *Blackjack) updateMenu(index *int, total int) {
	if b.pressed("up") {
		*index = (*index + total - 1) % total
	}
	if b.pressed("down") {
		*index = (*index + 1) % total
	}
}

func (b *Blackjack) startGame() {
	b.resetDeck()
	b.player = nil
	b.dealer = nil

	b.dealer = append(b.dealer, b.draw())
	b.dealer = append(b.dealer, b.draw())
	b.player = append(b.player, b.draw())
	b.player = append(b.player, b.draw())

	b.menuIndex = 0
	b.endMenuIndex = 0
	b.message = ""
	b.phase = PhasePlayer
}

func (b *Blackjack) playerHit() {
	b.player = append(b.player, b.draw())
	p := handTotal(b.player)
	if p == 21 {
		b.end("Blackjack! You win.")
		return
	}
	if p > 21 {
		b.end("Bust. You lose.")
		return
	}
}

func (b *Blackjack) runDealer() {
	for {
		d := handTotal(b.dealer)
		if d < 17 {
			b.dealer = append(b.dealer, b.draw())
			continue
		}
		break
	}
	b.resolveOutcome()
}

func (b *Blackjack) resolveOutcome() {
	p := handTotal(b.player)
	d := handTotal(b.dealer)
	if d > 21 {
		b.end("Dealer busts. You win.")
		return
	}
	if p > d {
		b.end("You win.")
		return
	}
	if p < d {
		b.end("Dealer wins.")
		return
	}
	b.end("Push.")
}

func (b *Blackjack) end(msg string) {
	b.message = msg
	b.phase = PhaseOver
}

func (b *Blackjack) resetDeck() {
	b.deck = newDeck()
	shuffle(b.deck)
}

func (b *Blackjack) draw() Card {
	if len(b.deck) == 0 {
		b.resetDeck()
	}
	c := b.deck[0]
	b.deck = b.deck[1:]
	return c
}

func (b *Blackjack) drawSplash(r *render.Renderer, style grid.Style) {
	if b.splash == nil {
		return
	}
	startY := max(1, (r.Frame.H-b.splash.H)/2-1)
	startX := max(0, (r.Frame.W-b.splash.W)/2)
	r.DrawSprite(startX, startY, b.splash)
	prompt := "press enter to continue"
	promptX := max(0, (r.Frame.W-len(prompt))/2)
	r.DrawText(promptX, startY+b.splash.H+1, prompt, style)
}

func newDeck() []Card {
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits := []rune{'♠', '♥', '♦', '♣'}
	deck := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, Card{Rank: r, Suit: s})
		}
	}
	return deck
}

func shuffle(deck []Card) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(deck) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

func handTotal(cards []Card) int {
	sum := 0
	aces := 0
	for _, c := range cards {
		switch c.Rank {
		case "A":
			sum += 1
			aces++
		case "K", "Q", "J":
			sum += 10
		default:
			var v int
			fmt.Sscanf(c.Rank, "%d", &v)
			sum += v
		}
	}
	for aces > 0 {
		if sum+9 <= 21 {
			sum += 9
		}
		aces--
	}
	return sum
}

func drawHand(r *render.Renderer, x, y int, cards []Card, revealAll bool, textStyle, redStyle grid.Style, sprites map[rune]*render.Sprite, back *render.Sprite) {
	cx := x
	for i, c := range cards {
		reveal := revealAll
		if i == 1 && !revealAll {
			reveal = false
		}
		drawCard(r, cx, y, c, reveal, textStyle, redStyle, sprites, back)
		cx += 8
	}
}

func drawCard(r *render.Renderer, x, y int, c Card, reveal bool, textStyle, redStyle grid.Style, sprites map[rune]*render.Sprite, back *render.Sprite) {
	w, h := 7, 5
	if !reveal {
		if back != nil {
			r.DrawSprite(x, y, back)
		} else {
			drawCardBox(r, x, y, w, h, textStyle)
			fillCardBack(r, x, y, w, h, textStyle)
		}
		return
	}
	suitStyle := textStyle
	if c.Suit == '♥' || c.Suit == '♦' {
		suitStyle = redStyle
	}
	if sprite, ok := sprites[c.Suit]; ok && sprite != nil {
		r.DrawSprite(x, y, sprite)
	} else {
		drawCardBox(r, x, y, w, h, textStyle)
		r.DrawText(x+3, y+2, string(c.Suit), suitStyle)
	}
	r.DrawText(x+1, y+1, padRight(c.Rank, 2), suitStyle)
}

func drawCardBox(r *render.Renderer, x, y, w, h int, style grid.Style) {
	r.Frame.Set(x, y, grid.Cell{Ch: '╭', Style: style})
	r.Frame.Set(x+w-1, y, grid.Cell{Ch: '╮', Style: style})
	r.Frame.Set(x, y+h-1, grid.Cell{Ch: '╰', Style: style})
	r.Frame.Set(x+w-1, y+h-1, grid.Cell{Ch: '╯', Style: style})
	for i := 1; i < w-1; i++ {
		r.Frame.Set(x+i, y, grid.Cell{Ch: '─', Style: style})
		r.Frame.Set(x+i, y+h-1, grid.Cell{Ch: '─', Style: style})
	}
	for j := 1; j < h-1; j++ {
		r.Frame.Set(x, y+j, grid.Cell{Ch: '│', Style: style})
		r.Frame.Set(x+w-1, y+j, grid.Cell{Ch: '│', Style: style})
	}
}

func fillCardBack(r *render.Renderer, x, y, w, h int, style grid.Style) {
	for row := 1; row < h-1; row++ {
		for col := 1; col < w-1; col++ {
			r.Frame.Set(x+col, y+row, grid.Cell{Ch: '░', Style: style})
		}
	}
}

func drawMenu(r *render.Renderer, x, y int, items []string, index int, style grid.Style) {
	for i, item := range items {
		prefix := "  "
		if i == index {
			prefix = "▶ "
		}
		r.DrawText(x, y+i, prefix+item, style)
	}
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func loadCardSprites() map[rune]*render.Sprite {
	sprites := map[rune]*render.Sprite{}
	load := func(suit rune, base string) {
		sprites[suit] = loadSprite(base)
	}
	load('♠', "card_spades")
	load('♣', "card_clubs")
	load('♥', "card_hearts")
	load('♦', "card_diamonds")
	return sprites
}

func loadSprite(base string) *render.Sprite {
	sprite, err := assets.LoadMaskedSprite(filepath.Join("demos/blackjack/assets", base))
	if err != nil {
		log.Fatal(err)
	}
	return sprite
}

func (b *Blackjack) layoutOffsets(r *render.Renderer, dealerTotal, playerTotal int) (int, int) {
	screenW := r.Frame.W
	screenH := r.Frame.H
	handW := max(handWidth(len(b.dealer)), handWidth(len(b.player)))
	labelW := max(len(fmt.Sprintf("Dealer hand (%d):", dealerTotal)), len(fmt.Sprintf("Your hand (%d):", playerTotal)))
	menuW := max(menuWidth([]string{"Hit", "Stay", "Quit"}), menuWidth([]string{"Play again", "Quit"}))
	contentW := max(handW, max(labelW, menuW))

	contentH := 15
	if b.phase == PhaseOver {
		contentH = 17
	}
	x0 := max(0, (screenW-contentW)/2)
	y0 := max(0, (screenH-contentH)/2)
	return x0, y0
}

func handWidth(count int) int {
	if count <= 0 {
		return 0
	}
	return count*8 - 1
}

func menuWidth(items []string) int {
	maxLen := 0
	for _, item := range items {
		if len(item) > maxLen {
			maxLen = len(item)
		}
	}
	return maxLen + 2
}

func main() {
	game := NewBlackjack()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: game.bg}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
