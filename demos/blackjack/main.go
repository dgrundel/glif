package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/dgrundel/glif/assets"
	"github.com/dgrundel/glif/engine"
	"github.com/dgrundel/glif/grid"
	"github.com/dgrundel/glif/input"
	"github.com/dgrundel/glif/render"
	"github.com/gdamore/tcell/v3"
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
	textStyle := grid.Style{Fg: tcell.ColorWhite, Bg: tcell.ColorReset}
	redStyle := grid.Style{Fg: tcell.NewRGBColor(220, 40, 40), Bg: tcell.ColorReset}

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

	r.DrawText(0, 0, fmt.Sprintf("Dealer hand (%d):", dealerTotal), textStyle)
	drawHand(r, 0, 1, b.dealer, b.phase != PhasePlayer, textStyle, redStyle)

	r.DrawText(0, 7, fmt.Sprintf("Your hand (%d):", playerTotal), textStyle)
	drawHand(r, 0, 8, b.player, true, textStyle, redStyle)

	switch b.phase {
	case PhasePlayer:
		drawMenu(r, 0, 14, []string{"Hit", "Stay", "Quit"}, b.menuIndex, textStyle)
	case PhaseOver:
		r.DrawText(0, 14, b.message, textStyle)
		drawMenu(r, 0, 16, []string{"Play again", "Quit"}, b.endMenuIndex, textStyle)
	}
}

func (b *Blackjack) Resize(w, h int) {
	_ = w
	_ = h
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
	startY := 1
	r.DrawSprite(0, startY, b.splash)
	prompt := "press enter to continue"
	r.DrawText(2, startY+b.splash.H+1, prompt, style)
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

func drawHand(r *render.Renderer, x, y int, cards []Card, revealAll bool, textStyle, redStyle grid.Style) {
	cx := x
	for i, c := range cards {
		reveal := revealAll
		if i == 1 && !revealAll {
			reveal = false
		}
		drawCard(r, cx, y, c, reveal, textStyle, redStyle)
		cx += 8
	}
}

func drawCard(r *render.Renderer, x, y int, c Card, reveal bool, textStyle, redStyle grid.Style) {
	w, h := 7, 5
	drawCardBox(r, x, y, w, h, textStyle)
	if !reveal {
		fillCardBack(r, x, y, w, h, textStyle)
		return
	}
	suitStyle := textStyle
	if c.Suit == '♥' || c.Suit == '♦' {
		suitStyle = redStyle
	}
	r.DrawText(x+1, y+1, padRight(c.Rank, 2), suitStyle)
	r.DrawText(x+3, y+2, string(c.Suit), suitStyle)
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

func main() {
	game := NewBlackjack()
	eng, err := engine.New(game, 0)
	if err != nil {
		log.Fatal(err)
	}
	eng.Frame.Clear = grid.Cell{Ch: ' ', Style: grid.Style{Fg: tcell.ColorReset, Bg: tcell.ColorReset}}
	eng.Frame.ClearAll()
	if err := eng.Run(game); err != nil {
		log.Fatal(err)
	}
}
